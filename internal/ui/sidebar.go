package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	DefaultSidebarWidth = 25
	MinSidebarWidth     = 15
	MaxSidebarWidth     = 60
)

// Sidebar represents the file explorer sidebar.
type Sidebar struct {
	fileTree *FileTree
	width    int
	height   int
	visible  bool

	// Selection
	selectedIndex int
	scrollOffset  int

	// Styles
	titleStyle    lipgloss.Style
	itemStyle     lipgloss.Style
	selectedStyle lipgloss.Style
	dirStyle      lipgloss.Style
	borderStyle   lipgloss.Style
}

// NewSidebar creates a new sidebar.
func NewSidebar() *Sidebar {
	return &Sidebar{
		fileTree: NewFileTree(),
		width:    DefaultSidebarWidth,
		visible:  true,

		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("238")).
			Padding(0, 1),
		itemStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		selectedStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")),
		dirStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("117")).
			Bold(true),
		borderStyle: lipgloss.NewStyle().
			BorderRight(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")),
	}
}

// LoadDirectory loads a directory into the sidebar.
func (s *Sidebar) LoadDirectory(path string) error {
	return s.fileTree.LoadDirectory(path)
}

// SetSize sets the sidebar dimensions.
func (s *Sidebar) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetWidth sets the sidebar width.
func (s *Sidebar) SetWidth(width int) {
	if width < MinSidebarWidth {
		width = MinSidebarWidth
	}
	if width > MaxSidebarWidth {
		width = MaxSidebarWidth
	}
	s.width = width
}

// Width returns the sidebar width.
func (s *Sidebar) Width() int {
	if !s.visible {
		return 0
	}
	return s.width
}

// Height returns the sidebar height.
func (s *Sidebar) Height() int {
	return s.height
}

// Toggle toggles sidebar visibility.
func (s *Sidebar) Toggle() {
	s.visible = !s.visible
}

// Show shows the sidebar.
func (s *Sidebar) Show() {
	s.visible = true
}

// Hide hides the sidebar.
func (s *Sidebar) Hide() {
	s.visible = false
}

// IsVisible returns whether the sidebar is visible.
func (s *Sidebar) IsVisible() bool {
	return s.visible
}

// MoveUp moves selection up.
func (s *Sidebar) MoveUp() {
	if s.selectedIndex > 0 {
		s.selectedIndex--
		s.ensureVisible()
	}
}

// MoveDown moves selection down.
func (s *Sidebar) MoveDown() {
	nodes := s.fileTree.GetVisibleNodes()
	if s.selectedIndex < len(nodes)-1 {
		s.selectedIndex++
		s.ensureVisible()
	}
}

// ensureVisible ensures the selected item is visible.
func (s *Sidebar) ensureVisible() {
	contentHeight := s.height - 2 // Account for title and bottom hint
	if contentHeight < 1 {
		contentHeight = 1
	}

	if s.selectedIndex < s.scrollOffset {
		s.scrollOffset = s.selectedIndex
	}
	if s.selectedIndex >= s.scrollOffset+contentHeight {
		s.scrollOffset = s.selectedIndex - contentHeight + 1
	}
}

// ToggleSelected toggles the selected directory or returns the selected file path.
func (s *Sidebar) ToggleSelected() string {
	nodes := s.fileTree.GetVisibleNodes()
	if s.selectedIndex >= len(nodes) {
		return ""
	}

	node := nodes[s.selectedIndex]
	if node.IsDir {
		s.fileTree.Toggle(node.Path)
		return ""
	}

	return node.Path
}

// Enter handles enter key on selected item.
func (s *Sidebar) Enter() string {
	return s.ToggleSelected()
}

// GetSelectedPath returns the currently selected file path.
func (s *Sidebar) GetSelectedPath() string {
	nodes := s.fileTree.GetVisibleNodes()
	if s.selectedIndex >= len(nodes) {
		return ""
	}
	return nodes[s.selectedIndex].Path
}

// SelectPath selects a specific path in the sidebar.
func (s *Sidebar) SelectPath(path string) {
	nodes := s.fileTree.GetVisibleNodes()
	for i, node := range nodes {
		if node.Path == path {
			s.selectedIndex = i
			s.ensureVisible()
			return
		}
	}
}

// Refresh reloads the file tree.
func (s *Sidebar) Refresh() error {
	return s.fileTree.Refresh()
}

// View renders the sidebar.
func (s *Sidebar) View() string {
	if !s.visible {
		return ""
	}

	var lines []string
	contentWidth := s.width - 2 // Account for border

	// Title
	title := s.titleStyle.Width(contentWidth).Render("EXPLORER")
	lines = append(lines, title)

	// File tree
	nodes := s.fileTree.GetVisibleNodes()
	contentHeight := s.height - 2 // Title + hint

	// Render visible nodes
	for i := s.scrollOffset; i < len(nodes) && i < s.scrollOffset+contentHeight; i++ {
		node := nodes[i]
		depth := s.fileTree.GetNodeDepth(node)

		// Build line
		indent := strings.Repeat("  ", depth)
		icon := ""
		name := node.Name

		if node.IsDir {
			if s.fileTree.IsExpanded(node.Path) {
				icon = "- "
			} else {
				icon = "+ "
			}
		} else {
			icon = "  "
		}

		line := indent + icon + name

		// Truncate if too long
		if len(line) > contentWidth {
			line = line[:contentWidth-3] + "..."
		}

		// Pad to width
		if len(line) < contentWidth {
			line += strings.Repeat(" ", contentWidth-len(line))
		}

		// Apply style
		var styledLine string
		if i == s.selectedIndex {
			styledLine = s.selectedStyle.Render(line)
		} else if node.IsDir {
			styledLine = s.dirStyle.Render(line)
		} else {
			styledLine = s.itemStyle.Render(line)
		}

		lines = append(lines, styledLine)
	}

	// Fill remaining space
	for len(lines) < s.height-1 {
		lines = append(lines, strings.Repeat(" ", contentWidth))
	}

	// Hint at bottom
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(contentWidth).
		Render("(Ctrl+B)")
	lines = append(lines, hint)

	// Join and add border
	content := strings.Join(lines, "\n")
	return s.borderStyle.Render(content)
}

// HandleClick handles a mouse click at the given y position.
func (s *Sidebar) HandleClick(y int) string {
	if y == 0 {
		// Clicked on title
		return ""
	}

	// Adjust for title
	index := s.scrollOffset + y - 1

	nodes := s.fileTree.GetVisibleNodes()
	if index >= len(nodes) {
		return ""
	}

	s.selectedIndex = index
	return s.ToggleSelected()
}

// ScrollUp scrolls the sidebar up.
func (s *Sidebar) ScrollUp(amount int) {
	s.scrollOffset -= amount
	if s.scrollOffset < 0 {
		s.scrollOffset = 0
	}
}

// ScrollDown scrolls the sidebar down.
func (s *Sidebar) ScrollDown(amount int) {
	nodes := s.fileTree.GetVisibleNodes()
	maxOffset := len(nodes) - (s.height - 2)
	if maxOffset < 0 {
		maxOffset = 0
	}

	s.scrollOffset += amount
	if s.scrollOffset > maxOffset {
		s.scrollOffset = maxOffset
	}
}
