package editor

import (
	"strings"

	"github.com/DDZ-DO/vex/internal/syntax"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultTabWidth   = 4
	defaultHistoryMax = 1000
)

// Editor is the main editor component that coordinates buffer, cursor,
// selection, history, and rendering.
type Editor struct {
	buffer      *Buffer
	cursor      *Cursor
	selection   *Selection
	history     *History
	highlighter *syntax.Highlighter

	// View state
	scrollY     int // First visible line
	scrollX     int // First visible column
	width       int
	height      int

	// Settings
	tabWidth    int
	showLineNum bool
	wordWrap    bool

	// Cached highlighted lines
	highlightedLines []syntax.StyledLine
	highlightDirty   bool

	// Line number gutter width
	gutterWidth int

	// Styles
	lineNumStyle    lipgloss.Style
	cursorLineStyle lipgloss.Style
	selectionStyle  lipgloss.Style
}

// NewEditor creates a new editor instance.
func NewEditor() *Editor {
	return &Editor{
		buffer:      NewBuffer(),
		cursor:      NewCursor(),
		selection:   NewSelection(),
		history:     NewHistory(defaultHistoryMax),
		highlighter: syntax.NewHighlighter(""),

		tabWidth:    defaultTabWidth,
		showLineNum: true,
		wordWrap:    false,

		highlightDirty: true,
		gutterWidth:    4,

		lineNumStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingRight(1),
		cursorLineStyle: lipgloss.NewStyle().Background(lipgloss.Color("236")),
		selectionStyle:  lipgloss.NewStyle().Background(lipgloss.Color("24")),
	}
}

// LoadFile loads a file into the editor.
func (e *Editor) LoadFile(filepath string) error {
	buf, err := NewBufferFromFile(filepath)
	if err != nil {
		return err
	}

	e.buffer = buf
	e.cursor = NewCursor()
	e.selection.Clear()
	e.history.Clear()
	e.highlighter.SetLanguageFromPath(filepath)
	e.highlightDirty = true
	e.scrollY = 0
	e.scrollX = 0
	e.updateGutterWidth()

	return nil
}

// NewFile creates a new empty file in the editor.
func (e *Editor) NewFile() {
	e.buffer = NewBuffer()
	e.cursor = NewCursor()
	e.selection.Clear()
	e.history.Clear()
	e.highlighter.SetLanguage("")
	e.highlightDirty = true
	e.scrollY = 0
	e.scrollX = 0
	e.updateGutterWidth()
}

// SetContent sets the editor content.
func (e *Editor) SetContent(content string) {
	e.buffer.SetContent(content)
	e.cursor = NewCursor()
	e.selection.Clear()
	e.history.Clear()
	e.highlightDirty = true
	e.scrollY = 0
	e.scrollX = 0
	e.updateGutterWidth()
}

// Content returns the current buffer content.
func (e *Editor) Content() string {
	return e.buffer.Content()
}

// SetSize sets the editor viewport size.
func (e *Editor) SetSize(width, height int) {
	e.width = width
	e.height = height
}

// updateGutterWidth calculates the line number gutter width.
func (e *Editor) updateGutterWidth() {
	lineCount := e.buffer.LineCount()
	width := 2
	for lineCount > 0 {
		lineCount /= 10
		width++
	}
	if width < 4 {
		width = 4
	}
	e.gutterWidth = width
}

// InsertRune inserts a single rune at the cursor position.
func (e *Editor) InsertRune(r rune) {
	// Delete selection if active
	if e.selection.Active && !e.selection.IsEmpty() {
		e.deleteSelection()
	}

	offset := e.cursor.Offset(e.buffer)
	text := string(r)

	e.history.RecordInsert(offset, text, e.cursor.Position())
	e.buffer.Insert(offset, text)
	e.cursor.MoveRight(e.buffer)
	e.highlightDirty = true
	e.ensureCursorVisible()
	e.updateGutterWidth()
}

// InsertText inserts a string at the cursor position.
func (e *Editor) InsertText(text string) {
	if text == "" {
		return
	}

	// Delete selection if active
	if e.selection.Active && !e.selection.IsEmpty() {
		e.deleteSelection()
	}

	offset := e.cursor.Offset(e.buffer)
	e.history.RecordInsert(offset, text, e.cursor.Position())
	e.buffer.Insert(offset, text)

	// Move cursor to end of inserted text
	newOffset := offset + len([]rune(text))
	line, col := e.buffer.OffsetToPosition(newOffset)
	e.cursor.MoveTo(line, col, e.buffer)

	e.highlightDirty = true
	e.ensureCursorVisible()
	e.updateGutterWidth()
}

// InsertNewline inserts a newline with auto-indentation.
func (e *Editor) InsertNewline() {
	// Get current line indentation
	currentLine := e.buffer.Line(e.cursor.Line)
	indent := ""
	for _, r := range currentLine {
		if r == ' ' || r == '\t' {
			indent += string(r)
		} else {
			break
		}
	}

	e.InsertText("\n" + indent)
}

// InsertTab inserts a tab (as spaces or tab character based on settings).
func (e *Editor) InsertTab() {
	// Insert spaces instead of tab
	spaces := strings.Repeat(" ", e.tabWidth)
	e.InsertText(spaces)
}

// Backspace deletes the character before the cursor.
func (e *Editor) Backspace() {
	// Delete selection if active
	if e.selection.Active && !e.selection.IsEmpty() {
		e.deleteSelection()
		return
	}

	offset := e.cursor.Offset(e.buffer)
	if offset == 0 {
		return
	}

	deleted := e.buffer.Delete(offset-1, 1)
	e.history.RecordDelete(offset-1, deleted, e.cursor.Position())
	e.cursor.MoveLeft(e.buffer)
	e.highlightDirty = true
	e.ensureCursorVisible()
	e.updateGutterWidth()
}

// Delete deletes the character at the cursor.
func (e *Editor) Delete() {
	// Delete selection if active
	if e.selection.Active && !e.selection.IsEmpty() {
		e.deleteSelection()
		return
	}

	offset := e.cursor.Offset(e.buffer)
	if offset >= e.buffer.Length() {
		return
	}

	deleted := e.buffer.Delete(offset, 1)
	e.history.RecordDelete(offset, deleted, e.cursor.Position())
	e.highlightDirty = true
	e.updateGutterWidth()
}

// DeleteLine deletes the current line.
func (e *Editor) DeleteLine() {
	line := e.cursor.Line
	lineStart := e.buffer.PositionToOffset(line, 0)
	lineLen := e.buffer.LineLength(line)

	// Include the newline if not on last line
	deleteLen := lineLen
	if line < e.buffer.LineCount()-1 {
		deleteLen++ // Include newline
	} else if line > 0 {
		// On last line, delete previous newline
		lineStart--
		deleteLen++
	}

	if deleteLen > 0 {
		deleted := e.buffer.Delete(lineStart, deleteLen)
		e.history.RecordDelete(lineStart, deleted, e.cursor.Position())
		e.cursor.Clamp(e.buffer)
		e.highlightDirty = true
		e.updateGutterWidth()
	}
}

// deleteSelection deletes the selected text.
func (e *Editor) deleteSelection() {
	if !e.selection.Active || e.selection.IsEmpty() {
		return
	}

	start, _ := e.selection.Normalized()
	text := e.selection.Text(e.buffer)
	startOffset := e.buffer.PositionToOffset(start.Line, start.Column)

	e.history.RecordDelete(startOffset, text, e.cursor.Position())
	e.selection.Delete(e.buffer)
	e.cursor.MoveTo(start.Line, start.Column, e.buffer)
	e.highlightDirty = true
	e.updateGutterWidth()
}

// Undo undoes the last action.
func (e *Editor) Undo() {
	action := e.history.Undo()
	if action == nil {
		return
	}

	cursorPos := ApplyUndo(action, e.buffer)
	e.cursor.MoveTo(cursorPos.Line, cursorPos.Column, e.buffer)
	e.selection.Clear()
	e.highlightDirty = true
	e.ensureCursorVisible()
	e.updateGutterWidth()
}

// Redo redoes the last undone action.
func (e *Editor) Redo() {
	action := e.history.Redo()
	if action == nil {
		return
	}

	cursorPos := ApplyRedo(action, e.buffer)
	e.cursor.MoveTo(cursorPos.Line, cursorPos.Column, e.buffer)
	e.selection.Clear()
	e.highlightDirty = true
	e.ensureCursorVisible()
	e.updateGutterWidth()
}

// Copy returns the selected text (or current line if no selection).
func (e *Editor) Copy() string {
	if e.selection.Active && !e.selection.IsEmpty() {
		return e.selection.Text(e.buffer)
	}
	// Copy entire line
	return e.buffer.Line(e.cursor.Line) + "\n"
}

// Cut cuts the selected text (or current line if no selection).
func (e *Editor) Cut() string {
	if e.selection.Active && !e.selection.IsEmpty() {
		text := e.selection.Text(e.buffer)
		e.deleteSelection()
		return text
	}
	// Cut entire line
	text := e.buffer.Line(e.cursor.Line) + "\n"
	e.DeleteLine()
	return text
}

// Paste inserts text at cursor position.
func (e *Editor) Paste(text string) {
	e.InsertText(text)
}

// SelectAll selects all text in the buffer.
func (e *Editor) SelectAll() {
	e.selection.SelectAll(e.buffer)
}

// SelectWord selects the word at cursor.
func (e *Editor) SelectWord() {
	e.selection.SelectWord(e.buffer, e.cursor)
}

// SelectLine selects the current line.
func (e *Editor) SelectLine() {
	e.selection.SelectLine(e.buffer, e.cursor)
}

// DuplicateLine duplicates the current line or selection.
func (e *Editor) DuplicateLine() {
	if e.selection.Active && !e.selection.IsEmpty() {
		// Duplicate selection
		text := e.selection.Text(e.buffer)
		_, end := e.selection.Normalized()
		endOffset := e.buffer.PositionToOffset(end.Line, end.Column)
		e.buffer.Insert(endOffset, text)
		e.history.RecordInsert(endOffset, text, e.cursor.Position())
	} else {
		// Duplicate line
		line := e.buffer.Line(e.cursor.Line)
		lineEnd := e.buffer.PositionToOffset(e.cursor.Line+1, 0)
		if e.cursor.Line >= e.buffer.LineCount()-1 {
			lineEnd = e.buffer.Length()
			line = "\n" + line
		}
		e.buffer.Insert(lineEnd, line+"\n")
		e.history.RecordInsert(lineEnd, line+"\n", e.cursor.Position())
		e.cursor.MoveDown(e.buffer)
	}
	e.highlightDirty = true
	e.updateGutterWidth()
}

// MoveLineUp moves the current line up.
func (e *Editor) MoveLineUp() {
	if e.cursor.Line == 0 {
		return
	}

	currentLine := e.buffer.Line(e.cursor.Line)
	aboveLine := e.buffer.Line(e.cursor.Line - 1)

	// Delete both lines
	lineStart := e.buffer.PositionToOffset(e.cursor.Line-1, 0)
	deleteLen := len([]rune(aboveLine)) + 1 + len([]rune(currentLine))
	if e.cursor.Line < e.buffer.LineCount()-1 {
		deleteLen++
	}

	e.buffer.Delete(lineStart, deleteLen)

	// Insert swapped
	newContent := currentLine + "\n" + aboveLine
	if e.cursor.Line < e.buffer.LineCount() {
		newContent += "\n"
	}
	e.buffer.Insert(lineStart, newContent)

	e.cursor.Line--
	e.highlightDirty = true
	e.ensureCursorVisible()
}

// MoveLineDown moves the current line down.
func (e *Editor) MoveLineDown() {
	if e.cursor.Line >= e.buffer.LineCount()-1 {
		return
	}

	currentLine := e.buffer.Line(e.cursor.Line)
	belowLine := e.buffer.Line(e.cursor.Line + 1)

	// Delete both lines
	lineStart := e.buffer.PositionToOffset(e.cursor.Line, 0)
	deleteLen := len([]rune(currentLine)) + 1 + len([]rune(belowLine))
	if e.cursor.Line+1 < e.buffer.LineCount()-1 {
		deleteLen++
	}

	e.buffer.Delete(lineStart, deleteLen)

	// Insert swapped
	newContent := belowLine + "\n" + currentLine
	if e.cursor.Line+1 < e.buffer.LineCount() {
		newContent += "\n"
	}
	e.buffer.Insert(lineStart, newContent)

	e.cursor.Line++
	e.highlightDirty = true
	e.ensureCursorVisible()
}

// MoveCursor moves the cursor with optional selection extension.
func (e *Editor) MoveCursor(direction string, extend bool) {
	if extend && !e.selection.Active {
		e.selection.StartAt(e.cursor.Position())
	}

	switch direction {
	case "left":
		e.cursor.MoveLeft(e.buffer)
	case "right":
		e.cursor.MoveRight(e.buffer)
	case "up":
		e.cursor.MoveUp(e.buffer)
	case "down":
		e.cursor.MoveDown(e.buffer)
	case "wordLeft":
		e.cursor.MoveWordLeft(e.buffer)
	case "wordRight":
		e.cursor.MoveWordRight(e.buffer)
	case "lineStart":
		e.cursor.MoveToLineStart()
	case "lineEnd":
		e.cursor.MoveToLineEnd(e.buffer)
	case "bufferStart":
		e.cursor.MoveToBufferStart()
	case "bufferEnd":
		e.cursor.MoveToBufferEnd(e.buffer)
	}

	if extend {
		e.selection.ExtendTo(e.cursor.Position())
	} else {
		e.selection.Clear()
	}

	e.ensureCursorVisible()
}

// GoToLine moves the cursor to a specific line (1-indexed).
func (e *Editor) GoToLine(line int) {
	e.cursor.MoveToLine(line, e.buffer)
	e.selection.Clear()
	e.ensureCursorVisible()
}

// PageUp moves the view and cursor up by one page.
func (e *Editor) PageUp() {
	e.cursor.PageUp(e.height-2, e.buffer)
	e.selection.Clear()
	e.ensureCursorVisible()
}

// PageDown moves the view and cursor down by one page.
func (e *Editor) PageDown() {
	e.cursor.PageDown(e.height-2, e.buffer)
	e.selection.Clear()
	e.ensureCursorVisible()
}

// ensureCursorVisible adjusts scroll to keep cursor in view.
func (e *Editor) ensureCursorVisible() {
	// Vertical scroll
	if e.cursor.Line < e.scrollY {
		e.scrollY = e.cursor.Line
	}
	if e.cursor.Line >= e.scrollY+e.height {
		e.scrollY = e.cursor.Line - e.height + 1
	}

	// Horizontal scroll
	textWidth := e.width - e.gutterWidth - 1
	if textWidth < 1 {
		textWidth = 1
	}
	if e.cursor.Column < e.scrollX {
		e.scrollX = e.cursor.Column
	}
	if e.cursor.Column >= e.scrollX+textWidth {
		e.scrollX = e.cursor.Column - textWidth + 1
	}
}

// Scroll scrolls the view by delta lines.
func (e *Editor) Scroll(delta int) {
	e.scrollY += delta
	if e.scrollY < 0 {
		e.scrollY = 0
	}
	maxScroll := e.buffer.LineCount() - e.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if e.scrollY > maxScroll {
		e.scrollY = maxScroll
	}
}

// HandleClick handles a mouse click at the given position.
func (e *Editor) HandleClick(x, y int, shift bool) {
	// Convert screen position to buffer position
	line := e.scrollY + y
	if line >= e.buffer.LineCount() {
		line = e.buffer.LineCount() - 1
	}
	if line < 0 {
		line = 0
	}

	col := e.scrollX + x - e.gutterWidth
	if col < 0 {
		col = 0
	}
	lineLen := e.buffer.LineLength(line)
	if col > lineLen {
		col = lineLen
	}

	if shift && !e.selection.Active {
		e.selection.StartAt(e.cursor.Position())
	}

	e.cursor.MoveTo(line, col, e.buffer)

	if shift {
		e.selection.ExtendTo(e.cursor.Position())
	} else {
		e.selection.Clear()
	}
}

// HandleDrag handles mouse drag for selection.
func (e *Editor) HandleDrag(x, y int) {
	if !e.selection.Active {
		e.selection.StartAt(e.cursor.Position())
	}

	line := e.scrollY + y
	if line >= e.buffer.LineCount() {
		line = e.buffer.LineCount() - 1
	}
	if line < 0 {
		line = 0
	}

	col := e.scrollX + x - e.gutterWidth
	if col < 0 {
		col = 0
	}
	lineLen := e.buffer.LineLength(line)
	if col > lineLen {
		col = lineLen
	}

	e.cursor.MoveTo(line, col, e.buffer)
	e.selection.ExtendTo(e.cursor.Position())
	e.ensureCursorVisible()
}

// DoubleClick selects the word at the click position.
func (e *Editor) DoubleClick(x, y int) {
	e.HandleClick(x, y, false)
	e.SelectWord()
}

// TripleClick selects the entire line.
func (e *Editor) TripleClick(x, y int) {
	e.HandleClick(x, y, false)
	e.SelectLine()
}

// Save saves the buffer to its file.
func (e *Editor) Save() error {
	return e.buffer.Save()
}

// SaveAs saves the buffer to a new file.
func (e *Editor) SaveAs(filepath string) error {
	err := e.buffer.SaveAs(filepath)
	if err == nil {
		e.highlighter.SetLanguageFromPath(filepath)
		e.highlightDirty = true
	}
	return err
}

// Modified returns whether the buffer has unsaved changes.
func (e *Editor) Modified() bool {
	return e.buffer.Modified()
}

// Filepath returns the current file path.
func (e *Editor) Filepath() string {
	return e.buffer.Filepath()
}

// Language returns the detected programming language.
func (e *Editor) Language() string {
	return e.highlighter.Language()
}

// LineCount returns the number of lines.
func (e *Editor) LineCount() int {
	return e.buffer.LineCount()
}

// CursorLine returns the current cursor line (0-indexed).
func (e *Editor) CursorLine() int {
	return e.cursor.Line
}

// CursorColumn returns the current cursor column (0-indexed).
func (e *Editor) CursorColumn() int {
	return e.cursor.Column
}

// LineEnding returns the line ending style.
func (e *Editor) LineEnding() string {
	return e.buffer.LineEnding()
}

// Encoding returns the character encoding.
func (e *Editor) Encoding() string {
	return e.buffer.Encoding()
}

// updateHighlighting refreshes syntax highlighting cache.
func (e *Editor) updateHighlighting() {
	if !e.highlightDirty {
		return
	}
	e.highlightedLines = e.highlighter.Highlight(e.buffer.Content())
	e.highlightDirty = false
}

// View renders the editor view.
func (e *Editor) View() string {
	if e.width == 0 || e.height == 0 {
		return ""
	}

	e.updateHighlighting()

	var lines []string
	textWidth := e.width - e.gutterWidth - 1
	if textWidth < 1 {
		textWidth = 1
	}

	for y := 0; y < e.height; y++ {
		lineNum := e.scrollY + y
		var lineContent string

		if lineNum < e.buffer.LineCount() {
			// Render line number
			if e.showLineNum {
				lineNumStr := lipgloss.NewStyle().
					Width(e.gutterWidth - 1).
					Align(lipgloss.Right).
					Render(formatLineNum(lineNum + 1))
				if lineNum == e.cursor.Line {
					lineNumStr = lipgloss.NewStyle().
						Foreground(lipgloss.Color("252")).
						Width(e.gutterWidth - 1).
						Align(lipgloss.Right).
						Render(formatLineNum(lineNum + 1))
				} else {
					lineNumStr = e.lineNumStyle.
						Width(e.gutterWidth - 1).
						Align(lipgloss.Right).
						Render(formatLineNum(lineNum + 1))
				}
				lineContent = lineNumStr + " "
			}

			// Render line content with syntax highlighting and selection
			lineText := e.buffer.Line(lineNum)
			lineContent += e.renderLine(lineNum, lineText, textWidth)
		} else {
			// Empty line
			if e.showLineNum {
				lineContent = strings.Repeat(" ", e.gutterWidth)
			}
			lineContent += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("~")
		}

		lines = append(lines, lineContent)
	}

	return strings.Join(lines, "\n")
}

// renderLine renders a single line with syntax highlighting and selection.
func (e *Editor) renderLine(lineNum int, lineText string, maxWidth int) string {
	// Handle horizontal scrolling
	runes := []rune(lineText)
	if e.scrollX > 0 {
		if e.scrollX >= len(runes) {
			runes = nil
		} else {
			runes = runes[e.scrollX:]
		}
	}

	// Truncate to fit width
	if len(runes) > maxWidth {
		runes = runes[:maxWidth]
	}

	// Replace tabs with spaces
	lineText = e.expandTabs(string(runes))
	runes = []rune(lineText)

	// Get selection range for this line
	selStart, selEnd := e.selection.GetLineRange(lineNum, len([]rune(e.buffer.Line(lineNum))))
	if selStart != -1 {
		selStart -= e.scrollX
		selEnd -= e.scrollX
		if selStart < 0 {
			selStart = 0
		}
	}

	// Build the line with highlighting and selection
	var result strings.Builder

	// Get highlighted segments for this line
	var segments []syntax.StyledSegment
	if lineNum < len(e.highlightedLines) {
		segments = e.highlightedLines[lineNum].Segments
	}
	if len(segments) == 0 {
		segments = []syntax.StyledSegment{{Text: lineText, Style: lipgloss.NewStyle()}}
	}

	// Flatten segments accounting for scrollX
	flatRunes := make([]struct {
		r     rune
		style lipgloss.Style
	}, 0, len(runes))

	pos := 0
	for _, seg := range segments {
		for _, r := range seg.Text {
			if pos >= e.scrollX && pos < e.scrollX+maxWidth {
				// Expand tab
				if r == '\t' {
					for i := 0; i < e.tabWidth; i++ {
						flatRunes = append(flatRunes, struct {
							r     rune
							style lipgloss.Style
						}{' ', seg.Style})
					}
				} else {
					flatRunes = append(flatRunes, struct {
						r     rune
						style lipgloss.Style
					}{r, seg.Style})
				}
			}
			pos++
		}
	}

	// Render each rune with appropriate style
	for i, fr := range flatRunes {
		style := fr.style

		// Apply selection highlighting
		if selStart != -1 && i >= selStart && i < selEnd {
			style = style.Background(lipgloss.Color("24"))
		}

		// Apply cursor highlight (only if no selection)
		if lineNum == e.cursor.Line && i == e.cursor.Column-e.scrollX && !e.selection.Active {
			style = style.Reverse(true)
		}

		result.WriteString(style.Render(string(fr.r)))
	}

	// Render cursor at end of line
	cursorCol := e.cursor.Column - e.scrollX
	if lineNum == e.cursor.Line && cursorCol >= 0 && cursorCol == len(flatRunes) {
		style := lipgloss.NewStyle().Reverse(true)
		if e.selection.Active && selStart != -1 && cursorCol >= selStart && cursorCol < selEnd {
			style = style.Background(lipgloss.Color("24"))
		}
		result.WriteString(style.Render(" "))
	}

	return result.String()
}

// expandTabs replaces tabs with spaces.
func (e *Editor) expandTabs(s string) string {
	return strings.ReplaceAll(s, "\t", strings.Repeat(" ", e.tabWidth))
}

// formatLineNum formats a line number for display.
func formatLineNum(n int) string {
	return strings.TrimSpace(lipgloss.NewStyle().Render(intToStr(n)))
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// Update handles Bubble Tea messages.
func (e *Editor) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.SetSize(msg.Width, msg.Height)
	}
	return nil
}

// Buffer returns the underlying buffer (for advanced operations).
func (e *Editor) Buffer() *Buffer {
	return e.buffer
}

// Cursor returns the cursor (for position info).
func (e *Editor) Cursor() *Cursor {
	return e.cursor
}

// Selection returns the selection (for selection info).
func (e *Editor) Selection() *Selection {
	return e.selection
}

// Find searches for text and moves cursor to the match.
func (e *Editor) Find(text string, caseSensitive bool) bool {
	offset := e.cursor.Offset(e.buffer)
	found := e.buffer.FindNext(text, offset+1, caseSensitive)
	if found == -1 {
		// Wrap around
		found = e.buffer.FindNext(text, 0, caseSensitive)
	}
	if found == -1 {
		return false
	}

	line, col := e.buffer.OffsetToPosition(found)
	e.cursor.MoveTo(line, col, e.buffer)

	// Select the found text
	e.selection.SetRange(
		Position{Line: line, Column: col},
		Position{Line: line, Column: col + len([]rune(text))},
	)
	// Adjust end position for multi-line matches
	endLine, endCol := e.buffer.OffsetToPosition(found + len([]rune(text)))
	e.selection.End = Position{Line: endLine, Column: endCol}

	e.ensureCursorVisible()
	return true
}

// FindPrevious searches backwards for text.
func (e *Editor) FindPrevious(text string, caseSensitive bool) bool {
	offset := e.cursor.Offset(e.buffer)
	found := e.buffer.FindPrevious(text, offset, caseSensitive)
	if found == -1 {
		// Wrap around
		found = e.buffer.FindPrevious(text, e.buffer.Length(), caseSensitive)
	}
	if found == -1 {
		return false
	}

	line, col := e.buffer.OffsetToPosition(found)
	e.cursor.MoveTo(line, col, e.buffer)

	// Select the found text
	endLine, endCol := e.buffer.OffsetToPosition(found + len([]rune(text)))
	e.selection.SetRange(
		Position{Line: line, Column: col},
		Position{Line: endLine, Column: endCol},
	)

	e.ensureCursorVisible()
	return true
}

// Replace replaces the current selection or next occurrence.
func (e *Editor) Replace(find, replace string, caseSensitive bool) bool {
	// If we have a selection that matches, replace it
	if e.selection.Active && !e.selection.IsEmpty() {
		selectedText := e.selection.Text(e.buffer)
		matches := selectedText == find
		if !caseSensitive {
			matches = strings.EqualFold(selectedText, find)
		}
		if matches {
			e.deleteSelection()
			e.InsertText(replace)
			return true
		}
	}

	// Otherwise find next and select it
	return e.Find(find, caseSensitive)
}

// ReplaceAll replaces all occurrences.
func (e *Editor) ReplaceAll(find, replace string, caseSensitive bool) int {
	return e.buffer.ReplaceAll(find, replace, caseSensitive)
}
