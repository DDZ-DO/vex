package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// Command represents a command in the palette.
type Command struct {
	ID          string
	Label       string
	Category    string
	Keybinding  string
	Description string
}

// CommandPalette provides a fuzzy-searchable command palette.
type CommandPalette struct {
	visible     bool
	input       string
	cursorPos   int
	commands    []Command
	filtered    []Command
	selected    int
	scrollOffset int

	width  int
	height int

	// Styles
	overlayStyle   lipgloss.Style
	inputStyle     lipgloss.Style
	itemStyle      lipgloss.Style
	selectedStyle  lipgloss.Style
	keybindStyle   lipgloss.Style
	categoryStyle  lipgloss.Style
}

// NewCommandPalette creates a new command palette.
func NewCommandPalette() *CommandPalette {
	return &CommandPalette{
		commands: defaultCommands(),

		overlayStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1),
		inputStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("238")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),
		itemStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),
		selectedStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1),
		keybindStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		categoryStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("109")),
	}
}

// defaultCommands returns the default set of commands.
func defaultCommands() []Command {
	return []Command{
		// File operations
		{ID: "file.save", Label: "Save File", Category: "File", Keybinding: "Ctrl+S"},
		{ID: "file.saveAs", Label: "Save As...", Category: "File", Keybinding: "Ctrl+Shift+S"},
		{ID: "file.new", Label: "New File", Category: "File", Keybinding: "Ctrl+N"},
		{ID: "file.open", Label: "Open File", Category: "File", Keybinding: "Ctrl+O"},
		{ID: "file.close", Label: "Close File", Category: "File", Keybinding: "Ctrl+W"},

		// Edit operations
		{ID: "edit.undo", Label: "Undo", Category: "Edit", Keybinding: "Ctrl+Z"},
		{ID: "edit.redo", Label: "Redo", Category: "Edit", Keybinding: "Ctrl+Y"},
		{ID: "edit.cut", Label: "Cut", Category: "Edit", Keybinding: "Ctrl+X"},
		{ID: "edit.copy", Label: "Copy", Category: "Edit", Keybinding: "Ctrl+C"},
		{ID: "edit.paste", Label: "Paste", Category: "Edit", Keybinding: "Ctrl+V"},
		{ID: "edit.selectAll", Label: "Select All", Category: "Edit", Keybinding: "Ctrl+A"},
		{ID: "edit.duplicateLine", Label: "Duplicate Line", Category: "Edit", Keybinding: "Ctrl+D"},
		{ID: "edit.deleteLine", Label: "Delete Line", Category: "Edit", Keybinding: "Ctrl+Shift+K"},
		{ID: "edit.moveLineUp", Label: "Move Line Up", Category: "Edit", Keybinding: "Alt+Up"},
		{ID: "edit.moveLineDown", Label: "Move Line Down", Category: "Edit", Keybinding: "Alt+Down"},

		// Search operations
		{ID: "search.find", Label: "Find", Category: "Search", Keybinding: "Ctrl+F"},
		{ID: "search.replace", Label: "Find and Replace", Category: "Search", Keybinding: "Ctrl+H"},
		{ID: "search.findNext", Label: "Find Next", Category: "Search", Keybinding: "F3"},
		{ID: "search.findPrevious", Label: "Find Previous", Category: "Search", Keybinding: "Shift+F3"},

		// Navigation
		{ID: "nav.goToLine", Label: "Go to Line", Category: "Go", Keybinding: "Ctrl+G"},
		{ID: "nav.goToStart", Label: "Go to Start", Category: "Go", Keybinding: "Ctrl+Home"},
		{ID: "nav.goToEnd", Label: "Go to End", Category: "Go", Keybinding: "Ctrl+End"},

		// View
		{ID: "view.toggleSidebar", Label: "Toggle Sidebar", Category: "View", Keybinding: "Ctrl+B"},
		{ID: "view.commandPalette", Label: "Command Palette", Category: "View", Keybinding: "Ctrl+P"},

		// Application
		{ID: "app.quit", Label: "Quit", Category: "Application", Keybinding: "Ctrl+Q"},
	}
}

// SetCommands sets the available commands.
func (cp *CommandPalette) SetCommands(commands []Command) {
	cp.commands = commands
	cp.updateFilter()
}

// AddCommand adds a command to the palette.
func (cp *CommandPalette) AddCommand(cmd Command) {
	cp.commands = append(cp.commands, cmd)
	cp.updateFilter()
}

// SetSize sets the palette dimensions.
func (cp *CommandPalette) SetSize(width, height int) {
	cp.width = width
	cp.height = height
}

// Show shows the command palette.
func (cp *CommandPalette) Show() {
	cp.visible = true
	cp.input = ""
	cp.cursorPos = 0
	cp.selected = 0
	cp.scrollOffset = 0
	cp.updateFilter()
}

// Hide hides the command palette.
func (cp *CommandPalette) Hide() {
	cp.visible = false
}

// Toggle toggles the command palette visibility.
func (cp *CommandPalette) Toggle() {
	if cp.visible {
		cp.Hide()
	} else {
		cp.Show()
	}
}

// IsVisible returns whether the palette is visible.
func (cp *CommandPalette) IsVisible() bool {
	return cp.visible
}

// Input handles text input.
func (cp *CommandPalette) Input(s string) {
	cp.input = cp.input[:cp.cursorPos] + s + cp.input[cp.cursorPos:]
	cp.cursorPos += len(s)
	cp.selected = 0
	cp.scrollOffset = 0
	cp.updateFilter()
}

// Backspace handles backspace.
func (cp *CommandPalette) Backspace() {
	if cp.cursorPos > 0 {
		cp.input = cp.input[:cp.cursorPos-1] + cp.input[cp.cursorPos:]
		cp.cursorPos--
		cp.selected = 0
		cp.scrollOffset = 0
		cp.updateFilter()
	}
}

// Delete handles delete key.
func (cp *CommandPalette) Delete() {
	if cp.cursorPos < len(cp.input) {
		cp.input = cp.input[:cp.cursorPos] + cp.input[cp.cursorPos+1:]
		cp.updateFilter()
	}
}

// MoveLeft moves cursor left.
func (cp *CommandPalette) MoveLeft() {
	if cp.cursorPos > 0 {
		cp.cursorPos--
	}
}

// MoveRight moves cursor right.
func (cp *CommandPalette) MoveRight() {
	if cp.cursorPos < len(cp.input) {
		cp.cursorPos++
	}
}

// MoveUp moves selection up.
func (cp *CommandPalette) MoveUp() {
	if cp.selected > 0 {
		cp.selected--
		cp.ensureVisible()
	}
}

// MoveDown moves selection down.
func (cp *CommandPalette) MoveDown() {
	if cp.selected < len(cp.filtered)-1 {
		cp.selected++
		cp.ensureVisible()
	}
}

// ensureVisible ensures selected item is visible.
func (cp *CommandPalette) ensureVisible() {
	maxVisible := cp.maxVisibleItems()
	if cp.selected < cp.scrollOffset {
		cp.scrollOffset = cp.selected
	}
	if cp.selected >= cp.scrollOffset+maxVisible {
		cp.scrollOffset = cp.selected - maxVisible + 1
	}
}

// maxVisibleItems returns the maximum number of visible items.
func (cp *CommandPalette) maxVisibleItems() int {
	// Palette height - input line - padding
	max := 10
	if cp.height > 0 {
		max = cp.height/2 - 3
		if max < 5 {
			max = 5
		}
		if max > 15 {
			max = 15
		}
	}
	return max
}

// Select returns the selected command and hides the palette.
func (cp *CommandPalette) Select() *Command {
	if cp.selected >= len(cp.filtered) {
		return nil
	}
	cmd := cp.filtered[cp.selected]
	cp.Hide()
	return &cmd
}

// GetSelectedCommand returns the currently selected command without hiding.
func (cp *CommandPalette) GetSelectedCommand() *Command {
	if cp.selected >= len(cp.filtered) {
		return nil
	}
	return &cp.filtered[cp.selected]
}

// updateFilter updates the filtered command list based on input.
func (cp *CommandPalette) updateFilter() {
	if cp.input == "" {
		cp.filtered = cp.commands
		return
	}

	// Use fuzzy matching
	var labels []string
	for _, cmd := range cp.commands {
		labels = append(labels, cmd.Label+" "+cmd.Category)
	}

	matches := fuzzy.Find(cp.input, labels)
	cp.filtered = make([]Command, len(matches))
	for i, match := range matches {
		cp.filtered[i] = cp.commands[match.Index]
	}
}

// View renders the command palette.
func (cp *CommandPalette) View() string {
	if !cp.visible {
		return ""
	}

	paletteWidth := cp.width - 20
	if paletteWidth > 60 {
		paletteWidth = 60
	}
	if paletteWidth < 40 {
		paletteWidth = 40
	}

	var lines []string

	// Input line with cursor
	inputLine := cp.input
	if cp.cursorPos < len(inputLine) {
		inputLine = inputLine[:cp.cursorPos] + "|" + inputLine[cp.cursorPos:]
	} else {
		inputLine += "|"
	}
	prompt := "> " + inputLine
	lines = append(lines, cp.inputStyle.Width(paletteWidth).Render(prompt))

	// Separator
	lines = append(lines, strings.Repeat("-", paletteWidth))

	// Command list
	maxVisible := cp.maxVisibleItems()
	endIdx := cp.scrollOffset + maxVisible
	if endIdx > len(cp.filtered) {
		endIdx = len(cp.filtered)
	}

	for i := cp.scrollOffset; i < endIdx; i++ {
		cmd := cp.filtered[i]
		line := cp.renderCommandLine(cmd, paletteWidth, i == cp.selected)
		lines = append(lines, line)
	}

	// Show "no results" if empty
	if len(cp.filtered) == 0 {
		noResults := cp.itemStyle.Width(paletteWidth).Render("No matching commands")
		lines = append(lines, noResults)
	}

	// Scroll indicators
	if cp.scrollOffset > 0 {
		lines = append([]string{lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  ...")}, lines...)
	}
	if endIdx < len(cp.filtered) {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  ..."))
	}

	content := strings.Join(lines, "\n")
	return cp.overlayStyle.Render(content)
}

// renderCommandLine renders a single command line.
func (cp *CommandPalette) renderCommandLine(cmd Command, width int, selected bool) string {
	label := cmd.Label
	keybind := ""
	if cmd.Keybinding != "" {
		keybind = cp.keybindStyle.Render(cmd.Keybinding)
	}

	// Calculate available width for label
	labelWidth := width - lipgloss.Width(keybind) - 4
	if len(label) > labelWidth {
		label = label[:labelWidth-3] + "..."
	}

	// Build line with label and keybinding
	spacing := width - len(label) - lipgloss.Width(keybind) - 2
	if spacing < 1 {
		spacing = 1
	}

	line := label + strings.Repeat(" ", spacing) + keybind

	if selected {
		return cp.selectedStyle.Width(width).Render(line)
	}
	return cp.itemStyle.Width(width).Render(line)
}

// HandleClick handles a click at the given position.
func (cp *CommandPalette) HandleClick(y int) *Command {
	// Account for input line and separator
	idx := cp.scrollOffset + y - 2
	if idx >= 0 && idx < len(cp.filtered) {
		cp.selected = idx
		return cp.Select()
	}
	return nil
}
