package editor

import (
	"time"
)

// ActionType represents the type of edit action.
type ActionType int

const (
	ActionInsert ActionType = iota
	ActionDelete
	ActionReplace
)

// EditAction represents a single undoable edit operation.
type EditAction struct {
	Type       ActionType
	Position   int       // Buffer offset where action occurred
	Text       string    // Text that was inserted or should be inserted on undo
	OldText    string    // Text that was replaced/deleted (for undo)
	Timestamp  time.Time
	CursorPos  Position  // Cursor position before the action
}

// History manages undo/redo stacks for edit operations.
type History struct {
	undoStack    []EditAction
	redoStack    []EditAction
	maxSize      int
	groupTimeout time.Duration // Time window for grouping actions
}

// NewHistory creates a new history with the specified max size.
func NewHistory(maxSize int) *History {
	return &History{
		undoStack:    make([]EditAction, 0, maxSize),
		redoStack:    make([]EditAction, 0, maxSize),
		maxSize:      maxSize,
		groupTimeout: 500 * time.Millisecond,
	}
}

// Push adds an action to the undo stack.
// It may merge with the previous action if they are similar and recent.
func (h *History) Push(action EditAction) {
	action.Timestamp = time.Now()

	// Clear redo stack on new action
	h.redoStack = h.redoStack[:0]

	// Try to merge with previous action
	if len(h.undoStack) > 0 {
		last := &h.undoStack[len(h.undoStack)-1]
		if h.canMerge(last, &action) {
			h.merge(last, &action)
			return
		}
	}

	// Add new action
	h.undoStack = append(h.undoStack, action)

	// Trim if over max size
	if len(h.undoStack) > h.maxSize {
		h.undoStack = h.undoStack[1:]
	}
}

// canMerge checks if two actions can be merged into one.
func (h *History) canMerge(prev, next *EditAction) bool {
	// Must be same type
	if prev.Type != next.Type {
		return false
	}

	// Must be within timeout
	if next.Timestamp.Sub(prev.Timestamp) > h.groupTimeout {
		return false
	}

	// Only merge consecutive inserts or deletes
	switch prev.Type {
	case ActionInsert:
		// Merge consecutive character inserts
		prevEnd := prev.Position + len([]rune(prev.Text))
		return next.Position == prevEnd && len([]rune(next.Text)) == 1
	case ActionDelete:
		// Merge consecutive deletes (backspace)
		return next.Position == prev.Position-1 || next.Position == prev.Position
	}

	return false
}

// merge combines two actions into the first one.
func (h *History) merge(prev, next *EditAction) {
	switch prev.Type {
	case ActionInsert:
		prev.Text += next.Text
	case ActionDelete:
		if next.Position < prev.Position {
			// Backspace - prepend text
			prev.OldText = next.OldText + prev.OldText
			prev.Position = next.Position
		} else {
			// Delete forward - append text
			prev.OldText += next.OldText
		}
	}
	prev.Timestamp = next.Timestamp
}

// Undo returns the action to undo, or nil if stack is empty.
func (h *History) Undo() *EditAction {
	if len(h.undoStack) == 0 {
		return nil
	}

	// Pop from undo stack
	action := h.undoStack[len(h.undoStack)-1]
	h.undoStack = h.undoStack[:len(h.undoStack)-1]

	// Push to redo stack
	h.redoStack = append(h.redoStack, action)

	return &action
}

// Redo returns the action to redo, or nil if stack is empty.
func (h *History) Redo() *EditAction {
	if len(h.redoStack) == 0 {
		return nil
	}

	// Pop from redo stack
	action := h.redoStack[len(h.redoStack)-1]
	h.redoStack = h.redoStack[:len(h.redoStack)-1]

	// Push to undo stack
	h.undoStack = append(h.undoStack, action)

	return &action
}

// CanUndo returns true if there are actions to undo.
func (h *History) CanUndo() bool {
	return len(h.undoStack) > 0
}

// CanRedo returns true if there are actions to redo.
func (h *History) CanRedo() bool {
	return len(h.redoStack) > 0
}

// Clear removes all history.
func (h *History) Clear() {
	h.undoStack = h.undoStack[:0]
	h.redoStack = h.redoStack[:0]
}

// UndoCount returns the number of actions in the undo stack.
func (h *History) UndoCount() int {
	return len(h.undoStack)
}

// RedoCount returns the number of actions in the redo stack.
func (h *History) RedoCount() int {
	return len(h.redoStack)
}

// BeginGroup starts a new action group (prevents merging with previous).
func (h *History) BeginGroup() {
	if len(h.undoStack) > 0 {
		// Set timestamp to past to prevent merging
		h.undoStack[len(h.undoStack)-1].Timestamp = time.Time{}
	}
}

// RecordInsert records an insert action.
func (h *History) RecordInsert(pos int, text string, cursorPos Position) {
	h.Push(EditAction{
		Type:      ActionInsert,
		Position:  pos,
		Text:      text,
		CursorPos: cursorPos,
	})
}

// RecordDelete records a delete action.
func (h *History) RecordDelete(pos int, deletedText string, cursorPos Position) {
	h.Push(EditAction{
		Type:      ActionDelete,
		Position:  pos,
		OldText:   deletedText,
		CursorPos: cursorPos,
	})
}

// RecordReplace records a replace action.
func (h *History) RecordReplace(pos int, oldText, newText string, cursorPos Position) {
	h.Push(EditAction{
		Type:      ActionReplace,
		Position:  pos,
		Text:      newText,
		OldText:   oldText,
		CursorPos: cursorPos,
	})
}

// ApplyUndo applies an undo action to the buffer and returns the cursor position.
func ApplyUndo(action *EditAction, buf *Buffer) Position {
	switch action.Type {
	case ActionInsert:
		// Undo insert = delete the inserted text
		buf.Delete(action.Position, len([]rune(action.Text)))
	case ActionDelete:
		// Undo delete = insert the deleted text back
		buf.Insert(action.Position, action.OldText)
	case ActionReplace:
		// Undo replace = replace new text with old text
		buf.Delete(action.Position, len([]rune(action.Text)))
		buf.Insert(action.Position, action.OldText)
	}
	return action.CursorPos
}

// ApplyRedo applies a redo action to the buffer and returns the new cursor position.
func ApplyRedo(action *EditAction, buf *Buffer) Position {
	switch action.Type {
	case ActionInsert:
		// Redo insert = insert the text again
		buf.Insert(action.Position, action.Text)
		line, col := buf.OffsetToPosition(action.Position + len([]rune(action.Text)))
		return Position{Line: line, Column: col}
	case ActionDelete:
		// Redo delete = delete the text again
		buf.Delete(action.Position, len([]rune(action.OldText)))
		line, col := buf.OffsetToPosition(action.Position)
		return Position{Line: line, Column: col}
	case ActionReplace:
		// Redo replace = replace old text with new text
		buf.Delete(action.Position, len([]rune(action.OldText)))
		buf.Insert(action.Position, action.Text)
		line, col := buf.OffsetToPosition(action.Position + len([]rune(action.Text)))
		return Position{Line: line, Column: col}
	}
	return action.CursorPos
}
