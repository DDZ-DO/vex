package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TabInfo contains display information for a tab.
type TabInfo struct {
	Name     string
	Path     string
	Modified bool
	Active   bool
}

// TabBar displays horizontal tabs for open files.
type TabBar struct {
	width    int
	tabs     []TabInfo
	visible  bool
	maxWidth int

	// Styles
	activeStyle   lipgloss.Style
	inactiveStyle lipgloss.Style
	modifiedStyle lipgloss.Style
	bgStyle       lipgloss.Style
}

// NewTabBar creates a new tab bar.
func NewTabBar() *TabBar {
	return &TabBar{
		visible:  true,
		maxWidth: 20,

		activeStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),
		inactiveStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("234")).
			Foreground(lipgloss.Color("245")).
			Padding(0, 1),
		modifiedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
		bgStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("234")),
	}
}

// SetWidth sets the tab bar width.
func (t *TabBar) SetWidth(width int) {
	t.width = width
}

// SetTabs sets the tabs to display.
func (t *TabBar) SetTabs(tabs []TabInfo) {
	t.tabs = tabs
}

// Show shows the tab bar.
func (t *TabBar) Show() {
	t.visible = true
}

// Hide hides the tab bar.
func (t *TabBar) Hide() {
	t.visible = false
}

// IsVisible returns whether the tab bar is visible.
func (t *TabBar) IsVisible() bool {
	return t.visible
}

// Height returns the height of the tab bar (1 if visible, 0 if hidden).
func (t *TabBar) Height() int {
	if !t.visible || len(t.tabs) <= 1 {
		return 0
	}
	return 1
}

// HandleClick returns the index of the clicked tab, or -1 if none.
func (t *TabBar) HandleClick(x int) int {
	if !t.visible || len(t.tabs) <= 1 {
		return -1
	}

	// Calculate tab positions
	currentX := 0
	for i, tab := range t.tabs {
		name := t.truncateName(tab.Name)
		if tab.Modified {
			name = "● " + name
		}
		tabWidth := len(name) + 2 // +2 for padding

		if x >= currentX && x < currentX+tabWidth {
			return i
		}
		currentX += tabWidth + 1 // +1 for separator
	}
	return -1
}

// truncateName truncates a name if too long.
func (t *TabBar) truncateName(name string) string {
	maxLen := t.maxWidth - 3
	if maxLen < 4 {
		maxLen = 4
	}
	if len(name) > maxLen {
		truncLen := maxLen - 3 // Leave room for "..."
		if truncLen < 1 {
			truncLen = 1
		}
		return name[:truncLen] + "..."
	}
	return name
}

// View renders the tab bar.
func (t *TabBar) View() string {
	// Don't show tab bar if only one tab
	if !t.visible || len(t.tabs) <= 1 {
		return ""
	}

	var parts []string

	for _, tab := range t.tabs {
		name := t.truncateName(tab.Name)

		// Add modified indicator
		if tab.Modified {
			name = t.modifiedStyle.Render("●") + " " + name
		}

		// Apply style based on active state
		var styled string
		if tab.Active {
			styled = t.activeStyle.Render(name)
		} else {
			styled = t.inactiveStyle.Render(name)
		}

		parts = append(parts, styled)
	}

	// Join tabs with separator
	content := strings.Join(parts, " ")

	// Pad to full width
	contentLen := lipgloss.Width(content)
	if contentLen < t.width {
		padding := strings.Repeat(" ", t.width-contentLen)
		content = content + t.bgStyle.Render(padding)
	}

	return t.bgStyle.Render(content)
}
