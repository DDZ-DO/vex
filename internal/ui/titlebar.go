package ui

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TitleBar represents the title bar component.
type TitleBar struct {
	width    int
	filename string
	filepath string
	modified bool

	style lipgloss.Style
}

// NewTitleBar creates a new title bar.
func NewTitleBar() *TitleBar {
	return &TitleBar{
		style: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Bold(true).
			Padding(0, 1),
	}
}

// SetWidth sets the title bar width.
func (t *TitleBar) SetWidth(width int) {
	t.width = width
}

// SetFile sets the current file information.
func (t *TitleBar) SetFile(path string, modified bool) {
	t.filepath = path
	t.modified = modified

	if path == "" {
		t.filename = "untitled"
	} else {
		t.filename = filepath.Base(path)
	}
}

// View renders the title bar.
func (t *TitleBar) View() string {
	// Build title
	title := "vex"
	if t.filepath != "" {
		// Show relative path or just filename
		title = "vex - " + t.shortenPath(t.filepath)
	} else {
		title = "vex - untitled"
	}

	// Modified indicator
	if t.modified {
		title += " [+]"
	}

	// Center the title
	padding := t.width - lipgloss.Width(title)
	if padding < 0 {
		padding = 0
	}

	leftPad := padding / 2
	rightPad := padding - leftPad

	content := strings.Repeat(" ", leftPad) + title + strings.Repeat(" ", rightPad)

	return t.style.Width(t.width).Render(content)
}

// shortenPath shortens a file path for display.
func (t *TitleBar) shortenPath(path string) string {
	// If path is short enough, return as-is
	maxLen := t.width - 20
	if maxLen < 20 {
		maxLen = 20
	}

	if len(path) <= maxLen {
		return path
	}

	// Try to shorten by replacing home with ~
	home, _ := filepath.Abs("~")
	if strings.HasPrefix(path, home) {
		path = "~" + path[len(home):]
	}

	if len(path) <= maxLen {
		return path
	}

	// Truncate from the left
	return "..." + path[len(path)-maxLen+3:]
}
