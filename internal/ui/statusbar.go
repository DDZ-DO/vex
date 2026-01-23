package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBar represents the status bar component at the bottom of the editor.
type StatusBar struct {
	width int

	// Editor info
	line       int
	column     int
	totalLines int
	language   string
	encoding   string
	lineEnding string
	tabWidth   int

	// Message
	message     string
	messageType MessageType

	style        lipgloss.Style
	errorStyle   lipgloss.Style
	warningStyle lipgloss.Style
	infoStyle    lipgloss.Style
}

// MessageType represents the type of status message.
type MessageType int

const (
	MessageNone MessageType = iota
	MessageInfo
	MessageWarning
	MessageError
)

// NewStatusBar creates a new status bar.
func NewStatusBar() *StatusBar {
	return &StatusBar{
		encoding:   "UTF-8",
		lineEnding: "LF",
		tabWidth:   4,
		language:   "plain",

		style: lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			Foreground(lipgloss.Color("252")),
		errorStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("196")).
			Foreground(lipgloss.Color("231")),
		warningStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("214")).
			Foreground(lipgloss.Color("232")),
		infoStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("39")).
			Foreground(lipgloss.Color("231")),
	}
}

// SetWidth sets the status bar width.
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// SetPosition sets the cursor position.
func (s *StatusBar) SetPosition(line, column, totalLines int) {
	s.line = line
	s.column = column
	s.totalLines = totalLines
}

// SetLanguage sets the language indicator.
func (s *StatusBar) SetLanguage(language string) {
	s.language = language
}

// SetEncoding sets the encoding indicator.
func (s *StatusBar) SetEncoding(encoding string) {
	s.encoding = encoding
}

// SetLineEnding sets the line ending indicator (LF or CRLF).
func (s *StatusBar) SetLineEnding(lineEnding string) {
	if lineEnding == "\r\n" {
		s.lineEnding = "CRLF"
	} else {
		s.lineEnding = "LF"
	}
}

// SetTabWidth sets the tab width indicator.
func (s *StatusBar) SetTabWidth(tabWidth int) {
	s.tabWidth = tabWidth
}

// SetMessage sets a temporary message to display.
func (s *StatusBar) SetMessage(message string, msgType MessageType) {
	s.message = message
	s.messageType = msgType
}

// ClearMessage clears the current message.
func (s *StatusBar) ClearMessage() {
	s.message = ""
	s.messageType = MessageNone
}

// View renders the status bar.
func (s *StatusBar) View() string {
	// If there's a message, show it
	if s.message != "" {
		return s.renderMessage()
	}

	// Build left side: position info
	position := fmt.Sprintf(" Ln %d, Col %d", s.line+1, s.column+1)

	// Build right side: language, encoding, line ending, tab width, help hint
	right := fmt.Sprintf("%s | %s | %s | Spaces: %d  Ctrl+P: Commands ",
		s.language, s.encoding, s.lineEnding, s.tabWidth)

	// Calculate spacing
	spacing := s.width - lipgloss.Width(position) - lipgloss.Width(right)
	if spacing < 0 {
		spacing = 0
	}

	content := position + strings.Repeat(" ", spacing) + right

	return s.style.Width(s.width).Render(content)
}

// renderMessage renders the message display.
func (s *StatusBar) renderMessage() string {
	var style lipgloss.Style
	switch s.messageType {
	case MessageError:
		style = s.errorStyle
	case MessageWarning:
		style = s.warningStyle
	case MessageInfo:
		style = s.infoStyle
	default:
		style = s.style
	}

	msg := " " + s.message
	if len(msg) < s.width {
		msg += strings.Repeat(" ", s.width-len(msg))
	}

	return style.Width(s.width).Render(msg)
}

// Height returns the height of the status bar (always 1).
func (s *StatusBar) Height() int {
	return 1
}
