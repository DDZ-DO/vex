package editor

// Position represents a position in the buffer.
type Position struct {
	Line   int
	Column int
}

// Cursor manages the cursor position and movement within a buffer.
type Cursor struct {
	Line         int
	Column       int
	PreferredCol int // Preferred column for vertical movement
}

// NewCursor creates a new cursor at position (0, 0).
func NewCursor() *Cursor {
	return &Cursor{
		Line:         0,
		Column:       0,
		PreferredCol: 0,
	}
}

// Position returns the current cursor position.
func (c *Cursor) Position() Position {
	return Position{Line: c.Line, Column: c.Column}
}

// SetPosition sets the cursor to the specified position.
func (c *Cursor) SetPosition(line, col int) {
	c.Line = line
	c.Column = col
	c.PreferredCol = col
}

// MoveTo moves the cursor to the specified position, clamping to buffer bounds.
func (c *Cursor) MoveTo(line, col int, buf *Buffer) {
	// Clamp line
	if line < 0 {
		line = 0
	}
	maxLine := buf.LineCount() - 1
	if maxLine < 0 {
		maxLine = 0
	}
	if line > maxLine {
		line = maxLine
	}

	// Clamp column
	if col < 0 {
		col = 0
	}
	lineLen := buf.LineLength(line)
	if col > lineLen {
		col = lineLen
	}

	c.Line = line
	c.Column = col
	c.PreferredCol = col
}

// MoveLeft moves the cursor one position to the left.
func (c *Cursor) MoveLeft(buf *Buffer) {
	if c.Column > 0 {
		c.Column--
	} else if c.Line > 0 {
		c.Line--
		c.Column = buf.LineLength(c.Line)
	}
	c.PreferredCol = c.Column
}

// MoveRight moves the cursor one position to the right.
func (c *Cursor) MoveRight(buf *Buffer) {
	lineLen := buf.LineLength(c.Line)
	if c.Column < lineLen {
		c.Column++
	} else if c.Line < buf.LineCount()-1 {
		c.Line++
		c.Column = 0
	}
	c.PreferredCol = c.Column
}

// MoveUp moves the cursor one line up.
func (c *Cursor) MoveUp(buf *Buffer) {
	if c.Line > 0 {
		c.Line--
		c.Column = c.clampColumn(c.PreferredCol, buf)
	}
}

// MoveDown moves the cursor one line down.
func (c *Cursor) MoveDown(buf *Buffer) {
	if c.Line < buf.LineCount()-1 {
		c.Line++
		c.Column = c.clampColumn(c.PreferredCol, buf)
	}
}

// MoveToLineStart moves the cursor to the beginning of the current line.
func (c *Cursor) MoveToLineStart() {
	c.Column = 0
	c.PreferredCol = 0
}

// MoveToLineEnd moves the cursor to the end of the current line.
func (c *Cursor) MoveToLineEnd(buf *Buffer) {
	c.Column = buf.LineLength(c.Line)
	c.PreferredCol = c.Column
}

// MoveToFirstNonWhitespace moves cursor to first non-whitespace character.
func (c *Cursor) MoveToFirstNonWhitespace(buf *Buffer) {
	line := buf.Line(c.Line)
	for i, r := range line {
		if r != ' ' && r != '\t' {
			c.Column = i
			c.PreferredCol = i
			return
		}
	}
	// If all whitespace or empty, go to start
	c.Column = 0
	c.PreferredCol = 0
}

// MoveWordLeft moves the cursor to the beginning of the previous word.
func (c *Cursor) MoveWordLeft(buf *Buffer) {
	offset := buf.PositionToOffset(c.Line, c.Column)

	if offset == 0 {
		return
	}

	// Skip any whitespace/non-word chars before current position
	offset--
	for offset > 0 && !isWordChar(buf.RuneAt(offset)) {
		offset--
	}

	// Move to the start of the word
	for offset > 0 && isWordChar(buf.RuneAt(offset-1)) {
		offset--
	}

	c.Line, c.Column = buf.OffsetToPosition(offset)
	c.PreferredCol = c.Column
}

// MoveWordRight moves the cursor to the beginning of the next word.
func (c *Cursor) MoveWordRight(buf *Buffer) {
	offset := buf.PositionToOffset(c.Line, c.Column)
	length := buf.Length()

	if offset >= length {
		return
	}

	// Skip current word if on a word
	for offset < length && isWordChar(buf.RuneAt(offset)) {
		offset++
	}

	// Skip non-word chars
	for offset < length && !isWordChar(buf.RuneAt(offset)) {
		offset++
	}

	c.Line, c.Column = buf.OffsetToPosition(offset)
	c.PreferredCol = c.Column
}

// MoveToBufferStart moves the cursor to the beginning of the buffer.
func (c *Cursor) MoveToBufferStart() {
	c.Line = 0
	c.Column = 0
	c.PreferredCol = 0
}

// MoveToBufferEnd moves the cursor to the end of the buffer.
func (c *Cursor) MoveToBufferEnd(buf *Buffer) {
	c.Line = buf.LineCount() - 1
	if c.Line < 0 {
		c.Line = 0
	}
	c.Column = buf.LineLength(c.Line)
	c.PreferredCol = c.Column
}

// MoveToLine moves the cursor to the specified line (1-indexed for UI).
func (c *Cursor) MoveToLine(lineNum int, buf *Buffer) {
	// Convert from 1-indexed (user-facing) to 0-indexed
	line := lineNum - 1
	if line < 0 {
		line = 0
	}
	maxLine := buf.LineCount() - 1
	if line > maxLine {
		line = maxLine
	}
	c.Line = line
	c.Column = 0
	c.PreferredCol = 0
}

// PageUp moves the cursor up by pageSize lines.
func (c *Cursor) PageUp(pageSize int, buf *Buffer) {
	c.Line -= pageSize
	if c.Line < 0 {
		c.Line = 0
	}
	c.Column = c.clampColumn(c.PreferredCol, buf)
}

// PageDown moves the cursor down by pageSize lines.
func (c *Cursor) PageDown(pageSize int, buf *Buffer) {
	c.Line += pageSize
	maxLine := buf.LineCount() - 1
	if c.Line > maxLine {
		c.Line = maxLine
	}
	c.Column = c.clampColumn(c.PreferredCol, buf)
}

// clampColumn clamps a column value to be within the current line.
func (c *Cursor) clampColumn(col int, buf *Buffer) int {
	lineLen := buf.LineLength(c.Line)
	if col > lineLen {
		return lineLen
	}
	if col < 0 {
		return 0
	}
	return col
}

// Offset returns the cursor position as a buffer offset.
func (c *Cursor) Offset(buf *Buffer) int {
	return buf.PositionToOffset(c.Line, c.Column)
}

// Clamp ensures the cursor is within valid bounds for the buffer.
func (c *Cursor) Clamp(buf *Buffer) {
	// Clamp line
	if c.Line < 0 {
		c.Line = 0
	}
	maxLine := buf.LineCount() - 1
	if maxLine < 0 {
		maxLine = 0
	}
	if c.Line > maxLine {
		c.Line = maxLine
	}

	// Clamp column
	if c.Column < 0 {
		c.Column = 0
	}
	lineLen := buf.LineLength(c.Line)
	if c.Column > lineLen {
		c.Column = lineLen
	}
}
