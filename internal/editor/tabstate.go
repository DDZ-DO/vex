package editor

import (
	"path/filepath"

	"github.com/DDZ-DO/vex/internal/syntax"
)

// TabState holds all state for a single open file/tab.
type TabState struct {
	buffer      *Buffer
	cursor      *Cursor
	selection   *Selection
	history     *History
	highlighter *syntax.Highlighter

	// View state per tab
	scrollX int
	scrollY int
}

// NewTabState creates a new empty tab.
func NewTabState() *TabState {
	return &TabState{
		buffer:      NewBuffer(),
		cursor:      NewCursor(),
		selection:   NewSelection(),
		history:     NewHistory(defaultHistoryMax),
		highlighter: syntax.NewHighlighter(""),
	}
}

// NewTabStateFromFile creates a tab with content loaded from a file.
func NewTabStateFromFile(path string) (*TabState, error) {
	buf, err := NewBufferFromFile(path)
	if err != nil {
		return nil, err
	}

	ts := &TabState{
		buffer:      buf,
		cursor:      NewCursor(),
		selection:   NewSelection(),
		history:     NewHistory(defaultHistoryMax),
		highlighter: syntax.NewHighlighter(""),
	}
	ts.highlighter.SetLanguageFromPath(path)

	return ts, nil
}

// Buffer returns the buffer.
func (ts *TabState) Buffer() *Buffer {
	return ts.buffer
}

// Cursor returns the cursor.
func (ts *TabState) Cursor() *Cursor {
	return ts.cursor
}

// Selection returns the selection.
func (ts *TabState) Selection() *Selection {
	return ts.selection
}

// History returns the history.
func (ts *TabState) History() *History {
	return ts.history
}

// Highlighter returns the syntax highlighter.
func (ts *TabState) Highlighter() *syntax.Highlighter {
	return ts.highlighter
}

// ScrollX returns horizontal scroll position.
func (ts *TabState) ScrollX() int {
	return ts.scrollX
}

// ScrollY returns vertical scroll position.
func (ts *TabState) ScrollY() int {
	return ts.scrollY
}

// SetScrollX sets horizontal scroll position.
func (ts *TabState) SetScrollX(x int) {
	ts.scrollX = x
}

// SetScrollY sets vertical scroll position.
func (ts *TabState) SetScrollY(y int) {
	ts.scrollY = y
}

// Filepath returns the file path.
func (ts *TabState) Filepath() string {
	return ts.buffer.Filepath()
}

// Modified returns whether the buffer has unsaved changes.
// Uses history save-point tracking for accurate detection after undo/redo.
func (ts *TabState) Modified() bool {
	return !ts.history.IsAtSavePoint()
}

// MarkSaved marks the current state as saved.
func (ts *TabState) MarkSaved() {
	ts.history.MarkSaved()
}

// Name returns a display name for the tab.
func (ts *TabState) Name() string {
	path := ts.buffer.Filepath()
	if path == "" {
		return "Untitled"
	}
	return filepath.Base(path)
}
