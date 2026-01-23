package editor

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"unicode/utf8"
)

const (
	initialGapSize = 1024
	LineEndingLF   = "\n"
	LineEndingCRLF = "\r\n"
)

// Buffer implements a Gap Buffer for efficient text editing operations.
// The gap buffer maintains a gap (empty space) in the middle of the data
// that moves to the cursor position, making insertions and deletions O(1)
// at the cursor location.
type Buffer struct {
	data       []rune
	gapStart   int
	gapEnd     int
	lines      []int  // Cache of line start positions (byte offsets into content)
	modified   bool
	filepath   string
	encoding   string
	lineEnding string
}

// NewBuffer creates a new empty buffer.
func NewBuffer() *Buffer {
	b := &Buffer{
		data:       make([]rune, initialGapSize),
		gapStart:   0,
		gapEnd:     initialGapSize,
		lines:      []int{0},
		encoding:   "UTF-8",
		lineEnding: LineEndingLF,
	}
	return b
}

// NewBufferFromFile creates a buffer with content loaded from a file.
func NewBufferFromFile(filepath string) (*Buffer, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	b := NewBuffer()
	b.filepath = filepath
	b.detectLineEnding(content)
	b.SetContent(string(content))
	b.modified = false

	return b, nil
}

// detectLineEnding detects whether the file uses LF or CRLF line endings.
func (b *Buffer) detectLineEnding(content []byte) {
	if bytes.Contains(content, []byte("\r\n")) {
		b.lineEnding = LineEndingCRLF
	} else {
		b.lineEnding = LineEndingLF
	}
}

// SetContent replaces the entire buffer content.
func (b *Buffer) SetContent(content string) {
	// Normalize line endings to LF internally
	content = strings.ReplaceAll(content, "\r\n", "\n")
	runes := []rune(content)

	b.data = make([]rune, len(runes)+initialGapSize)
	copy(b.data, runes)
	b.gapStart = len(runes)
	b.gapEnd = len(b.data)
	b.rebuildLineIndex()
	b.modified = true
}

// Content returns the full buffer content as a string.
func (b *Buffer) Content() string {
	result := make([]rune, b.Length())
	copy(result, b.data[:b.gapStart])
	copy(result[b.gapStart:], b.data[b.gapEnd:])
	return string(result)
}

// Length returns the number of runes in the buffer (excluding the gap).
func (b *Buffer) Length() int {
	return len(b.data) - (b.gapEnd - b.gapStart)
}

// gapSize returns the current size of the gap.
func (b *Buffer) gapSize() int {
	return b.gapEnd - b.gapStart
}

// moveGapTo moves the gap to the specified position.
func (b *Buffer) moveGapTo(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > b.Length() {
		pos = b.Length()
	}

	if pos < b.gapStart {
		// Move gap left: shift data from [pos, gapStart) to after gap
		moveCount := b.gapStart - pos
		copy(b.data[b.gapEnd-moveCount:b.gapEnd], b.data[pos:b.gapStart])
		b.gapStart = pos
		b.gapEnd -= moveCount
	} else if pos > b.gapStart {
		// Move gap right: shift data from [gapEnd, gapEnd+moveCount) to before gap
		moveCount := pos - b.gapStart
		copy(b.data[b.gapStart:b.gapStart+moveCount], b.data[b.gapEnd:b.gapEnd+moveCount])
		b.gapStart += moveCount
		b.gapEnd += moveCount
	}
}

// expandGap ensures the gap has at least minSize capacity.
func (b *Buffer) expandGap(minSize int) {
	if b.gapSize() >= minSize {
		return
	}

	// Double the gap size or use minSize, whichever is larger
	newGapSize := b.gapSize() * 2
	if newGapSize < minSize {
		newGapSize = minSize
	}
	if newGapSize < initialGapSize {
		newGapSize = initialGapSize
	}

	additionalSpace := newGapSize - b.gapSize()
	newData := make([]rune, len(b.data)+additionalSpace)

	// Copy data before gap
	copy(newData, b.data[:b.gapStart])
	// Copy data after gap to new position
	copy(newData[b.gapStart+newGapSize:], b.data[b.gapEnd:])

	b.gapEnd = b.gapStart + newGapSize
	b.data = newData
}

// Insert inserts text at the specified position.
func (b *Buffer) Insert(pos int, text string) {
	if len(text) == 0 {
		return
	}

	runes := []rune(text)
	b.moveGapTo(pos)
	b.expandGap(len(runes))

	copy(b.data[b.gapStart:], runes)
	b.gapStart += len(runes)
	b.modified = true
	b.rebuildLineIndex()
}

// Delete removes count runes starting at pos.
func (b *Buffer) Delete(pos, count int) string {
	if count <= 0 || pos < 0 || pos >= b.Length() {
		return ""
	}

	if pos+count > b.Length() {
		count = b.Length() - pos
	}

	b.moveGapTo(pos)

	// The deleted text is now at gapEnd
	deleted := string(b.data[b.gapEnd : b.gapEnd+count])
	b.gapEnd += count
	b.modified = true
	b.rebuildLineIndex()

	return deleted
}

// RuneAt returns the rune at the specified position.
func (b *Buffer) RuneAt(pos int) rune {
	if pos < 0 || pos >= b.Length() {
		return 0
	}

	if pos < b.gapStart {
		return b.data[pos]
	}
	return b.data[pos+(b.gapEnd-b.gapStart)]
}

// Substring returns a substring from start to end positions.
func (b *Buffer) Substring(start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > b.Length() {
		end = b.Length()
	}
	if start >= end {
		return ""
	}

	result := make([]rune, end-start)
	for i := start; i < end; i++ {
		result[i-start] = b.RuneAt(i)
	}
	return string(result)
}

// rebuildLineIndex rebuilds the line start position cache.
func (b *Buffer) rebuildLineIndex() {
	b.lines = []int{0}
	pos := 0

	for i := 0; i < b.Length(); i++ {
		r := b.RuneAt(i)
		pos++
		if r == '\n' {
			b.lines = append(b.lines, i+1)
		}
	}
}

// LineCount returns the number of lines in the buffer.
func (b *Buffer) LineCount() int {
	return len(b.lines)
}

// Line returns the content of the specified line (0-indexed).
func (b *Buffer) Line(lineNum int) string {
	if lineNum < 0 || lineNum >= len(b.lines) {
		return ""
	}

	start := b.lines[lineNum]
	var end int
	if lineNum+1 < len(b.lines) {
		end = b.lines[lineNum+1] - 1 // Exclude newline
	} else {
		end = b.Length()
	}

	if end < start {
		return ""
	}

	return b.Substring(start, end)
}

// LineLength returns the length of the specified line (excluding newline).
func (b *Buffer) LineLength(lineNum int) int {
	return utf8.RuneCountInString(b.Line(lineNum))
}

// PositionToOffset converts a line/column position to a buffer offset.
func (b *Buffer) PositionToOffset(line, col int) int {
	if line < 0 || line >= len(b.lines) {
		return 0
	}

	offset := b.lines[line] + col
	maxOffset := b.Length()

	// Don't go past end of line
	if line+1 < len(b.lines) {
		lineEnd := b.lines[line+1] - 1
		if offset > lineEnd {
			offset = lineEnd
		}
	}

	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}

	return offset
}

// OffsetToPosition converts a buffer offset to line/column.
func (b *Buffer) OffsetToPosition(offset int) (line, col int) {
	if offset < 0 {
		return 0, 0
	}
	if offset >= b.Length() {
		if len(b.lines) == 0 {
			return 0, 0
		}
		lastLine := len(b.lines) - 1
		return lastLine, b.LineLength(lastLine)
	}

	// Binary search for the line
	lo, hi := 0, len(b.lines)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if b.lines[mid] <= offset {
			lo = mid
		} else {
			hi = mid - 1
		}
	}

	return lo, offset - b.lines[lo]
}

// Modified returns whether the buffer has been modified since last save.
func (b *Buffer) Modified() bool {
	return b.modified
}

// SetModified sets the modified flag.
func (b *Buffer) SetModified(modified bool) {
	b.modified = modified
}

// Filepath returns the file path associated with this buffer.
func (b *Buffer) Filepath() string {
	return b.filepath
}

// SetFilepath sets the file path for this buffer.
func (b *Buffer) SetFilepath(filepath string) {
	b.filepath = filepath
}

// Encoding returns the character encoding of the buffer.
func (b *Buffer) Encoding() string {
	return b.encoding
}

// LineEnding returns the line ending style (LF or CRLF).
func (b *Buffer) LineEnding() string {
	return b.lineEnding
}

// Save writes the buffer content to the associated file.
func (b *Buffer) Save() error {
	if b.filepath == "" {
		return os.ErrInvalid
	}
	return b.SaveAs(b.filepath)
}

// SaveAs writes the buffer content to the specified file.
func (b *Buffer) SaveAs(filepath string) error {
	content := b.Content()

	// Convert line endings if necessary
	if b.lineEnding == LineEndingCRLF {
		content = strings.ReplaceAll(content, "\n", "\r\n")
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(content)
	if err != nil {
		return err
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	b.filepath = filepath
	b.modified = false
	return nil
}

// WordAt returns the word at the given position and its start/end offsets.
func (b *Buffer) WordAt(pos int) (word string, start, end int) {
	if pos < 0 || pos >= b.Length() {
		return "", pos, pos
	}

	// Find word start
	start = pos
	for start > 0 && isWordChar(b.RuneAt(start-1)) {
		start--
	}

	// Find word end
	end = pos
	for end < b.Length() && isWordChar(b.RuneAt(end)) {
		end++
	}

	return b.Substring(start, end), start, end
}

// isWordChar returns true if the rune is part of a word.
func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_'
}

// FindNext finds the next occurrence of text starting from pos.
// Returns -1 if not found.
func (b *Buffer) FindNext(text string, pos int, caseSensitive bool) int {
	if text == "" {
		return -1
	}

	content := b.Content()
	searchText := text

	if !caseSensitive {
		content = strings.ToLower(content)
		searchText = strings.ToLower(text)
	}

	if pos < 0 {
		pos = 0
	}

	idx := strings.Index(content[pos:], searchText)
	if idx == -1 {
		return -1
	}
	return pos + idx
}

// FindPrevious finds the previous occurrence of text before pos.
// Returns -1 if not found.
func (b *Buffer) FindPrevious(text string, pos int, caseSensitive bool) int {
	if text == "" || pos <= 0 {
		return -1
	}

	content := b.Content()
	searchText := text

	if !caseSensitive {
		content = strings.ToLower(content)
		searchText = strings.ToLower(text)
	}

	if pos > len(content) {
		pos = len(content)
	}

	idx := strings.LastIndex(content[:pos], searchText)
	return idx
}

// Replace replaces the first occurrence of old with new starting from pos.
// Returns the position after replacement, or -1 if not found.
func (b *Buffer) Replace(old, new string, pos int, caseSensitive bool) int {
	found := b.FindNext(old, pos, caseSensitive)
	if found == -1 {
		return -1
	}

	b.Delete(found, len([]rune(old)))
	b.Insert(found, new)
	return found + len([]rune(new))
}

// ReplaceAll replaces all occurrences of old with new.
// Returns the number of replacements made.
func (b *Buffer) ReplaceAll(old, new string, caseSensitive bool) int {
	count := 0
	pos := 0

	for {
		found := b.FindNext(old, pos, caseSensitive)
		if found == -1 {
			break
		}

		b.Delete(found, len([]rune(old)))
		b.Insert(found, new)
		pos = found + len([]rune(new))
		count++
	}

	return count
}
