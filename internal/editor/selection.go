package editor

// Selection represents a text selection in the buffer.
type Selection struct {
	Active bool     // Whether a selection is active
	Start  Position // Start of selection (anchor point)
	End    Position // End of selection (cursor point)
}

// NewSelection creates a new empty selection.
func NewSelection() *Selection {
	return &Selection{
		Active: false,
	}
}

// Clear removes the selection.
func (s *Selection) Clear() {
	s.Active = false
	s.Start = Position{}
	s.End = Position{}
}

// StartAt begins a selection at the specified position.
func (s *Selection) StartAt(pos Position) {
	s.Active = true
	s.Start = pos
	s.End = pos
}

// ExtendTo extends the selection to the specified position.
func (s *Selection) ExtendTo(pos Position) {
	if !s.Active {
		s.StartAt(pos)
		return
	}
	s.End = pos
}

// SetRange sets the selection to cover from start to end.
func (s *Selection) SetRange(start, end Position) {
	s.Active = true
	s.Start = start
	s.End = end
}

// Normalized returns the selection with Start before End.
func (s *Selection) Normalized() (start, end Position) {
	if !s.Active {
		return s.Start, s.End
	}

	if s.Start.Line < s.End.Line ||
		(s.Start.Line == s.End.Line && s.Start.Column <= s.End.Column) {
		return s.Start, s.End
	}
	return s.End, s.Start
}

// IsEmpty returns true if the selection has zero length.
func (s *Selection) IsEmpty() bool {
	if !s.Active {
		return true
	}
	return s.Start.Line == s.End.Line && s.Start.Column == s.End.Column
}

// Contains checks if a position is within the selection.
func (s *Selection) Contains(pos Position) bool {
	if !s.Active || s.IsEmpty() {
		return false
	}

	start, end := s.Normalized()

	// Before start?
	if pos.Line < start.Line || (pos.Line == start.Line && pos.Column < start.Column) {
		return false
	}

	// After end?
	if pos.Line > end.Line || (pos.Line == end.Line && pos.Column >= end.Column) {
		return false
	}

	return true
}

// ContainsLine checks if any part of the selection is on the given line.
func (s *Selection) ContainsLine(line int) bool {
	if !s.Active {
		return false
	}

	start, end := s.Normalized()
	return line >= start.Line && line <= end.Line
}

// GetLineRange returns the selection range for a specific line.
// Returns startCol, endCol for the selection on that line.
// Returns -1, -1 if the line is not part of the selection.
func (s *Selection) GetLineRange(line int, lineLength int) (startCol, endCol int) {
	if !s.Active || !s.ContainsLine(line) {
		return -1, -1
	}

	start, end := s.Normalized()

	if line < start.Line || line > end.Line {
		return -1, -1
	}

	// Start column for this line
	if line == start.Line {
		startCol = start.Column
	} else {
		startCol = 0
	}

	// End column for this line
	if line == end.Line {
		endCol = end.Column
	} else {
		endCol = lineLength + 1 // Include newline in selection visual
	}

	return startCol, endCol
}

// Text returns the selected text from the buffer.
func (s *Selection) Text(buf *Buffer) string {
	if !s.Active || s.IsEmpty() {
		return ""
	}

	start, end := s.Normalized()
	startOffset := buf.PositionToOffset(start.Line, start.Column)
	endOffset := buf.PositionToOffset(end.Line, end.Column)

	return buf.Substring(startOffset, endOffset)
}

// Delete removes the selected text from the buffer and returns it.
func (s *Selection) Delete(buf *Buffer) string {
	if !s.Active || s.IsEmpty() {
		return ""
	}

	text := s.Text(buf)
	start, _ := s.Normalized()
	startOffset := buf.PositionToOffset(start.Line, start.Column)

	buf.Delete(startOffset, len([]rune(text)))
	s.Clear()

	return text
}

// SelectWord selects the word at the cursor position.
func (s *Selection) SelectWord(buf *Buffer, cursor *Cursor) {
	offset := cursor.Offset(buf)
	_, start, end := buf.WordAt(offset)

	if start == end {
		s.Clear()
		return
	}

	startLine, startCol := buf.OffsetToPosition(start)
	endLine, endCol := buf.OffsetToPosition(end)

	s.SetRange(
		Position{Line: startLine, Column: startCol},
		Position{Line: endLine, Column: endCol},
	)
}

// SelectLine selects the entire line at the cursor position.
func (s *Selection) SelectLine(buf *Buffer, cursor *Cursor) {
	line := cursor.Line
	lineLen := buf.LineLength(line)

	s.SetRange(
		Position{Line: line, Column: 0},
		Position{Line: line, Column: lineLen},
	)

	// If not on the last line, include the newline
	if line < buf.LineCount()-1 {
		s.End = Position{Line: line + 1, Column: 0}
	}
}

// SelectAll selects all text in the buffer.
func (s *Selection) SelectAll(buf *Buffer) {
	lastLine := buf.LineCount() - 1
	if lastLine < 0 {
		lastLine = 0
	}

	s.SetRange(
		Position{Line: 0, Column: 0},
		Position{Line: lastLine, Column: buf.LineLength(lastLine)},
	)
}
