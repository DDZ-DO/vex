package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SearchMode represents the current search mode.
type SearchMode int

const (
	SearchModeFind SearchMode = iota
	SearchModeReplace
	SearchModeGoToLine
)

// SearchBar provides find and replace functionality.
type SearchBar struct {
	visible      bool
	mode         SearchMode
	searchInput  string
	replaceInput string
	cursorPos    int
	focusReplace bool

	// Search options
	caseSensitive bool
	wholeWord     bool
	regex         bool

	// Match info
	currentMatch int
	totalMatches int

	width int

	// Styles
	barStyle      lipgloss.Style
	inputStyle    lipgloss.Style
	labelStyle    lipgloss.Style
	matchStyle    lipgloss.Style
	buttonStyle   lipgloss.Style
	activeStyle   lipgloss.Style
	inactiveStyle lipgloss.Style
}

// NewSearchBar creates a new search bar.
func NewSearchBar() *SearchBar {
	return &SearchBar{
		mode: SearchModeFind,

		barStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			Padding(0, 1),
		inputStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("238")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),
		labelStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingRight(1),
		matchStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			PaddingLeft(1),
		buttonStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("239")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),
		activeStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1),
		inactiveStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("239")).
			Foreground(lipgloss.Color("245")).
			Padding(0, 1),
	}
}

// SetWidth sets the search bar width.
func (s *SearchBar) SetWidth(width int) {
	s.width = width
}

// Show shows the search bar in find mode.
func (s *SearchBar) Show() {
	s.visible = true
	s.mode = SearchModeFind
	s.focusReplace = false
	s.cursorPos = len(s.searchInput)
}

// ShowReplace shows the search bar in replace mode.
func (s *SearchBar) ShowReplace() {
	s.visible = true
	s.mode = SearchModeReplace
	s.focusReplace = false
	s.cursorPos = len(s.searchInput)
}

// ShowGoToLine shows the search bar in go-to-line mode.
func (s *SearchBar) ShowGoToLine() {
	s.visible = true
	s.mode = SearchModeGoToLine
	s.searchInput = ""
	s.cursorPos = 0
}

// Hide hides the search bar.
func (s *SearchBar) Hide() {
	s.visible = false
}

// IsVisible returns whether the search bar is visible.
func (s *SearchBar) IsVisible() bool {
	return s.visible
}

// Mode returns the current search mode.
func (s *SearchBar) Mode() SearchMode {
	return s.mode
}

// Input handles text input.
func (s *SearchBar) Input(text string) {
	if s.focusReplace {
		s.replaceInput = s.replaceInput[:s.cursorPos] + text + s.replaceInput[s.cursorPos:]
	} else {
		s.searchInput = s.searchInput[:s.cursorPos] + text + s.searchInput[s.cursorPos:]
	}
	s.cursorPos += len(text)
}

// Backspace handles backspace.
func (s *SearchBar) Backspace() {
	if s.cursorPos > 0 {
		if s.focusReplace {
			s.replaceInput = s.replaceInput[:s.cursorPos-1] + s.replaceInput[s.cursorPos:]
		} else {
			s.searchInput = s.searchInput[:s.cursorPos-1] + s.searchInput[s.cursorPos:]
		}
		s.cursorPos--
	}
}

// Delete handles delete key.
func (s *SearchBar) Delete() {
	input := s.currentInput()
	if s.cursorPos < len(input) {
		if s.focusReplace {
			s.replaceInput = s.replaceInput[:s.cursorPos] + s.replaceInput[s.cursorPos+1:]
		} else {
			s.searchInput = s.searchInput[:s.cursorPos] + s.searchInput[s.cursorPos+1:]
		}
	}
}

// MoveLeft moves cursor left.
func (s *SearchBar) MoveLeft() {
	if s.cursorPos > 0 {
		s.cursorPos--
	}
}

// MoveRight moves cursor right.
func (s *SearchBar) MoveRight() {
	input := s.currentInput()
	if s.cursorPos < len(input) {
		s.cursorPos++
	}
}

// currentInput returns the current active input.
func (s *SearchBar) currentInput() string {
	if s.focusReplace {
		return s.replaceInput
	}
	return s.searchInput
}

// Tab switches focus between search and replace fields.
func (s *SearchBar) Tab() {
	if s.mode == SearchModeReplace {
		s.focusReplace = !s.focusReplace
		if s.focusReplace {
			s.cursorPos = len(s.replaceInput)
		} else {
			s.cursorPos = len(s.searchInput)
		}
	}
}

// SearchText returns the current search text.
func (s *SearchBar) SearchText() string {
	return s.searchInput
}

// ReplaceText returns the current replace text.
func (s *SearchBar) ReplaceText() string {
	return s.replaceInput
}

// SetSearchText sets the search text.
func (s *SearchBar) SetSearchText(text string) {
	s.searchInput = text
	if !s.focusReplace {
		s.cursorPos = len(text)
	}
}

// IsCaseSensitive returns whether search is case sensitive.
func (s *SearchBar) IsCaseSensitive() bool {
	return s.caseSensitive
}

// ToggleCaseSensitive toggles case sensitivity.
func (s *SearchBar) ToggleCaseSensitive() {
	s.caseSensitive = !s.caseSensitive
}

// SetMatchInfo sets the current match information.
func (s *SearchBar) SetMatchInfo(current, total int) {
	s.currentMatch = current
	s.totalMatches = total
}

// LineNumber returns the entered line number (for go-to-line mode).
func (s *SearchBar) LineNumber() int {
	if s.mode != SearchModeGoToLine {
		return 0
	}
	var n int
	fmt.Sscanf(s.searchInput, "%d", &n)
	return n
}

// View renders the search bar.
func (s *SearchBar) View() string {
	if !s.visible {
		return ""
	}

	switch s.mode {
	case SearchModeGoToLine:
		return s.renderGoToLine()
	case SearchModeReplace:
		return s.renderReplace()
	default:
		return s.renderFind()
	}
}

// renderFind renders the find bar.
func (s *SearchBar) renderFind() string {
	var parts []string

	// Search label
	parts = append(parts, s.labelStyle.Render("Find:"))

	// Search input with cursor
	input := s.searchInput
	if s.cursorPos <= len(input) {
		input = input[:s.cursorPos] + "|" + input[s.cursorPos:]
	}
	inputWidth := 30
	parts = append(parts, s.inputStyle.Width(inputWidth).Render(input))

	// Match count
	matchInfo := ""
	if s.searchInput != "" {
		if s.totalMatches > 0 {
			matchInfo = fmt.Sprintf("%d of %d", s.currentMatch, s.totalMatches)
		} else {
			matchInfo = "No results"
		}
	}
	parts = append(parts, s.matchStyle.Render(matchInfo))

	// Options
	caseBtnStyle := s.inactiveStyle
	if s.caseSensitive {
		caseBtnStyle = s.activeStyle
	}
	parts = append(parts, caseBtnStyle.Render("Aa"))

	// Hints
	parts = append(parts, s.labelStyle.Render("  Enter: Next  Shift+Enter: Prev  Esc: Close"))

	content := strings.Join(parts, " ")
	return s.barStyle.Width(s.width).Render(content)
}

// renderReplace renders the find and replace bar.
func (s *SearchBar) renderReplace() string {
	var lines []string

	// First line: Find
	var findParts []string
	findParts = append(findParts, s.labelStyle.Render("Find:   "))

	searchInput := s.searchInput
	if !s.focusReplace && s.cursorPos <= len(searchInput) {
		searchInput = searchInput[:s.cursorPos] + "|" + searchInput[s.cursorPos:]
	}
	inputStyle := s.inputStyle
	if s.focusReplace {
		inputStyle = inputStyle.Background(lipgloss.Color("236"))
	}
	findParts = append(findParts, inputStyle.Width(30).Render(searchInput))

	// Match info
	matchInfo := ""
	if s.searchInput != "" {
		if s.totalMatches > 0 {
			matchInfo = fmt.Sprintf("%d of %d", s.currentMatch, s.totalMatches)
		} else {
			matchInfo = "No results"
		}
	}
	findParts = append(findParts, s.matchStyle.Render(matchInfo))

	lines = append(lines, strings.Join(findParts, " "))

	// Second line: Replace
	var replaceParts []string
	replaceParts = append(replaceParts, s.labelStyle.Render("Replace:"))

	replaceInput := s.replaceInput
	if s.focusReplace && s.cursorPos <= len(replaceInput) {
		replaceInput = replaceInput[:s.cursorPos] + "|" + replaceInput[s.cursorPos:]
	}
	replaceStyle := s.inputStyle
	if !s.focusReplace {
		replaceStyle = replaceStyle.Background(lipgloss.Color("236"))
	}
	replaceParts = append(replaceParts, replaceStyle.Width(30).Render(replaceInput))

	// Buttons
	replaceParts = append(replaceParts, s.buttonStyle.Render("Replace"))
	replaceParts = append(replaceParts, s.buttonStyle.Render("Replace All"))

	lines = append(lines, strings.Join(replaceParts, " "))

	content := strings.Join(lines, "\n")
	return s.barStyle.Width(s.width).Render(content)
}

// renderGoToLine renders the go-to-line bar.
func (s *SearchBar) renderGoToLine() string {
	var parts []string

	parts = append(parts, s.labelStyle.Render("Go to Line:"))

	input := s.searchInput
	if s.cursorPos <= len(input) {
		input = input[:s.cursorPos] + "|" + input[s.cursorPos:]
	}
	parts = append(parts, s.inputStyle.Width(10).Render(input))

	parts = append(parts, s.labelStyle.Render("  Enter: Go  Esc: Cancel"))

	content := strings.Join(parts, " ")
	return s.barStyle.Width(s.width).Render(content)
}

// Height returns the height of the search bar.
func (s *SearchBar) Height() int {
	if !s.visible {
		return 0
	}
	if s.mode == SearchModeReplace {
		return 2
	}
	return 1
}
