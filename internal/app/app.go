package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DDZ-DO/vex/internal/config"
	"github.com/DDZ-DO/vex/internal/editor"
	"github.com/DDZ-DO/vex/internal/keybindings"
	"github.com/DDZ-DO/vex/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.design/x/clipboard"
)

// FocusArea represents which UI component has focus.
type FocusArea int

const (
	FocusEditor FocusArea = iota
	FocusSidebar
	FocusCommandPalette
	FocusSearchBar
)

// App is the root model that orchestrates all components.
type App struct {
	// Components
	editor         *editor.Editor
	titleBar       *ui.TitleBar
	statusBar      *ui.StatusBar
	sidebar        *ui.Sidebar
	commandPalette *ui.CommandPalette
	searchBar      *ui.SearchBar

	// Configuration
	config      *config.Config
	keyBindings *keybindings.KeyBindings

	// State
	focus       FocusArea
	width       int
	height      int
	quitting    bool
	message     string
	messageTime time.Time

	// Clipboard
	clipboardInit bool
}

// New creates a new App instance.
func New() *App {
	cfg, _ := config.Load()

	app := &App{
		editor:         editor.NewEditor(),
		titleBar:       ui.NewTitleBar(),
		statusBar:      ui.NewStatusBar(),
		sidebar:        ui.NewSidebar(),
		commandPalette: ui.NewCommandPalette(),
		searchBar:      ui.NewSearchBar(),
		config:         cfg,
		keyBindings:    keybindings.NewKeyBindings(),
		focus:          FocusEditor,
	}

	// Initialize clipboard
	if err := clipboard.Init(); err == nil {
		app.clipboardInit = true
	}

	// Set initial sidebar visibility from config
	if !cfg.ShowSidebar {
		app.sidebar.Hide()
	}

	return app
}

// LoadFile loads a file into the editor.
func (a *App) LoadFile(filepath string) error {
	err := a.editor.LoadFile(filepath)
	if err != nil {
		return err
	}

	// Load directory into sidebar
	dir := filepath
	if info, err := os.Stat(filepath); err == nil && !info.IsDir() {
		dir = filepath[:len(filepath)-len(info.Name())-1]
		if dir == "" {
			dir = "."
		}
	}
	a.sidebar.LoadDirectory(dir)

	return nil
}

// GoToLine moves to a specific line (1-indexed).
func (a *App) GoToLine(line int) {
	a.editor.GoToLine(line)
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
	)
}

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Clear old messages
	if a.message != "" && time.Since(a.messageTime) > 3*time.Second {
		a.message = ""
		a.statusBar.ClearMessage()
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.handleResize(msg.Width, msg.Height)
		return a, nil

	case tea.KeyMsg:
		return a.handleKeyPress(msg)

	case tea.MouseMsg:
		return a.handleMouse(msg)
	}

	return a, nil
}

// handleResize handles window resize events.
func (a *App) handleResize(width, height int) {
	a.width = width
	a.height = height

	// Calculate component sizes
	sidebarWidth := a.sidebar.Width()
	editorWidth := width - sidebarWidth
	editorHeight := height - 3 // title bar + status bar + search bar

	// Update component sizes
	a.titleBar.SetWidth(width)
	a.statusBar.SetWidth(width)
	a.sidebar.SetSize(sidebarWidth, height-2) // Exclude title and status
	a.editor.SetSize(editorWidth, editorHeight)
	a.commandPalette.SetSize(width, height)
	a.searchBar.SetWidth(width)
}

// handleKeyPress handles keyboard input.
func (a *App) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle Escape first - closes overlays
	if msg.Type == tea.KeyEsc {
		if a.commandPalette.IsVisible() {
			a.commandPalette.Hide()
			a.focus = FocusEditor
			return a, nil
		}
		if a.searchBar.IsVisible() {
			a.searchBar.Hide()
			a.focus = FocusEditor
			a.handleResize(a.width, a.height)
			return a, nil
		}
		// Clear selection
		a.editor.Selection().Clear()
		return a, nil
	}

	// Handle command palette
	if a.commandPalette.IsVisible() {
		return a.handleCommandPaletteKey(msg)
	}

	// Handle search bar
	if a.searchBar.IsVisible() {
		return a.handleSearchBarKey(msg)
	}

	// Handle sidebar focus
	if a.focus == FocusSidebar {
		return a.handleSidebarKey(msg)
	}

	// Look up keybinding
	action := a.keyBindings.Lookup(msg)

	switch action {
	// File operations
	case keybindings.ActionSave:
		return a.save()
	case keybindings.ActionNew:
		a.editor.NewFile()
		a.showMessage("New file", ui.MessageInfo)
		return a, nil
	case keybindings.ActionQuit:
		return a.quit()

	// Edit operations
	case keybindings.ActionUndo:
		a.editor.Undo()
		return a, nil
	case keybindings.ActionRedo:
		a.editor.Redo()
		return a, nil
	case keybindings.ActionCut:
		text := a.editor.Cut()
		a.copyToClipboard(text)
		return a, nil
	case keybindings.ActionCopy:
		text := a.editor.Copy()
		a.copyToClipboard(text)
		return a, nil
	case keybindings.ActionPaste:
		text := a.pasteFromClipboard()
		a.editor.Paste(text)
		return a, nil
	case keybindings.ActionSelectAll:
		a.editor.SelectAll()
		return a, nil
	case keybindings.ActionDuplicateLine:
		a.editor.DuplicateLine()
		return a, nil
	case keybindings.ActionDeleteLine:
		a.editor.DeleteLine()
		return a, nil
	case keybindings.ActionSelectLine:
		a.editor.SelectLine()
		return a, nil

	// Navigation
	case keybindings.ActionMoveLeft:
		a.editor.MoveCursor("left", false)
		return a, nil
	case keybindings.ActionMoveRight:
		a.editor.MoveCursor("right", false)
		return a, nil
	case keybindings.ActionMoveUp:
		a.editor.MoveCursor("up", false)
		return a, nil
	case keybindings.ActionMoveDown:
		a.editor.MoveCursor("down", false)
		return a, nil
	case keybindings.ActionMoveWordLeft:
		a.editor.MoveCursor("wordLeft", false)
		return a, nil
	case keybindings.ActionMoveWordRight:
		a.editor.MoveCursor("wordRight", false)
		return a, nil
	case keybindings.ActionMoveLineStart:
		a.editor.MoveCursor("lineStart", false)
		return a, nil
	case keybindings.ActionMoveLineEnd:
		a.editor.MoveCursor("lineEnd", false)
		return a, nil
	case keybindings.ActionMoveBufferStart:
		a.editor.MoveCursor("bufferStart", false)
		return a, nil
	case keybindings.ActionMoveBufferEnd:
		a.editor.MoveCursor("bufferEnd", false)
		return a, nil
	case keybindings.ActionPageUp:
		a.editor.PageUp()
		return a, nil
	case keybindings.ActionPageDown:
		a.editor.PageDown()
		return a, nil
	case keybindings.ActionGoToLine:
		a.searchBar.ShowGoToLine()
		a.focus = FocusSearchBar
		a.handleResize(a.width, a.height)
		return a, nil

	// Selection
	case keybindings.ActionSelectLeft:
		a.editor.MoveCursor("left", true)
		return a, nil
	case keybindings.ActionSelectRight:
		a.editor.MoveCursor("right", true)
		return a, nil
	case keybindings.ActionSelectUp:
		a.editor.MoveCursor("up", true)
		return a, nil
	case keybindings.ActionSelectDown:
		a.editor.MoveCursor("down", true)
		return a, nil
	case keybindings.ActionSelectWordLeft:
		a.editor.MoveCursor("wordLeft", true)
		return a, nil
	case keybindings.ActionSelectWordRight:
		a.editor.MoveCursor("wordRight", true)
		return a, nil
	case keybindings.ActionSelectLineStart:
		a.editor.MoveCursor("lineStart", true)
		return a, nil
	case keybindings.ActionSelectLineEnd:
		a.editor.MoveCursor("lineEnd", true)
		return a, nil

	// Search
	case keybindings.ActionFind:
		a.searchBar.Show()
		a.focus = FocusSearchBar
		a.handleResize(a.width, a.height)
		return a, nil
	case keybindings.ActionReplace:
		a.searchBar.ShowReplace()
		a.focus = FocusSearchBar
		a.handleResize(a.width, a.height)
		return a, nil
	case keybindings.ActionFindNext:
		if a.searchBar.SearchText() != "" {
			a.editor.Find(a.searchBar.SearchText(), a.searchBar.IsCaseSensitive())
		}
		return a, nil
	case keybindings.ActionFindPrevious:
		if a.searchBar.SearchText() != "" {
			a.editor.FindPrevious(a.searchBar.SearchText(), a.searchBar.IsCaseSensitive())
		}
		return a, nil

	// View
	case keybindings.ActionToggleSidebar:
		a.sidebar.Toggle()
		a.handleResize(a.width, a.height)
		return a, nil
	case keybindings.ActionCommandPalette:
		a.commandPalette.Show()
		a.focus = FocusCommandPalette
		return a, nil

	// Text input
	case keybindings.ActionInsertNewline:
		a.editor.InsertNewline()
		return a, nil
	case keybindings.ActionInsertTab:
		a.editor.InsertTab()
		return a, nil
	case keybindings.ActionBackspace:
		a.editor.Backspace()
		return a, nil
	case keybindings.ActionDelete:
		a.editor.Delete()
		return a, nil
	}

	// Handle regular character input
	if msg.Type == tea.KeyRunes {
		for _, r := range msg.Runes {
			a.editor.InsertRune(r)
		}
		return a, nil
	}

	// Handle Alt+Arrow for line movement
	if msg.Alt {
		switch msg.Type {
		case tea.KeyUp:
			a.editor.MoveLineUp()
			return a, nil
		case tea.KeyDown:
			a.editor.MoveLineDown()
			return a, nil
		}
	}

	return a, nil
}

// handleCommandPaletteKey handles key input when command palette is focused.
func (a *App) handleCommandPaletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		a.commandPalette.MoveUp()
	case tea.KeyDown:
		a.commandPalette.MoveDown()
	case tea.KeyEnter:
		cmd := a.commandPalette.Select()
		a.focus = FocusEditor
		if cmd != nil {
			return a.executeCommand(cmd.ID)
		}
	case tea.KeyBackspace:
		a.commandPalette.Backspace()
	case tea.KeyDelete:
		a.commandPalette.Delete()
	case tea.KeyLeft:
		a.commandPalette.MoveLeft()
	case tea.KeyRight:
		a.commandPalette.MoveRight()
	case tea.KeyRunes:
		a.commandPalette.Input(string(msg.Runes))
	}
	return a, nil
}

// handleSearchBarKey handles key input when search bar is focused.
func (a *App) handleSearchBarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if a.searchBar.Mode() == ui.SearchModeGoToLine {
			lineNum := a.searchBar.LineNumber()
			if lineNum > 0 {
				a.editor.GoToLine(lineNum)
			}
			a.searchBar.Hide()
			a.focus = FocusEditor
			a.handleResize(a.width, a.height)
		} else {
			// Find next
			text := a.searchBar.SearchText()
			if text != "" {
				a.editor.Find(text, a.searchBar.IsCaseSensitive())
			}
		}
	case tea.KeyTab:
		a.searchBar.Tab()
	case tea.KeyBackspace:
		a.searchBar.Backspace()
	case tea.KeyDelete:
		a.searchBar.Delete()
	case tea.KeyLeft:
		a.searchBar.MoveLeft()
	case tea.KeyRight:
		a.searchBar.MoveRight()
	case tea.KeyRunes:
		a.searchBar.Input(string(msg.Runes))
		// Live search
		if a.searchBar.Mode() != ui.SearchModeGoToLine {
			text := a.searchBar.SearchText()
			if text != "" {
				a.editor.Find(text, a.searchBar.IsCaseSensitive())
			}
		}
	}
	return a, nil
}

// handleSidebarKey handles key input when sidebar is focused.
func (a *App) handleSidebarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		a.sidebar.MoveUp()
	case tea.KeyDown:
		a.sidebar.MoveDown()
	case tea.KeyEnter:
		path := a.sidebar.Enter()
		if path != "" {
			if err := a.editor.LoadFile(path); err != nil {
				a.showMessage("Error: "+err.Error(), ui.MessageError)
			}
			a.focus = FocusEditor
		}
	case tea.KeyTab, tea.KeyRight:
		a.focus = FocusEditor
	}
	return a, nil
}

// handleMouse handles mouse events.
func (a *App) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Action {
	case tea.MouseActionPress:
		// Check if click is in sidebar
		if a.sidebar.IsVisible() && msg.X < a.sidebar.Width() {
			a.focus = FocusSidebar
			path := a.sidebar.HandleClick(msg.Y - 1) // Adjust for title bar
			if path != "" {
				if err := a.editor.LoadFile(path); err != nil {
					a.showMessage("Error: "+err.Error(), ui.MessageError)
				}
				a.focus = FocusEditor
			}
			return a, nil
		}

		// Click in editor area
		a.focus = FocusEditor
		editorX := msg.X - a.sidebar.Width()
		editorY := msg.Y - 1 // Adjust for title bar
		shift := msg.Ctrl    // Bubble Tea doesn't have Shift detection in mouse, use Ctrl as workaround
		a.editor.HandleClick(editorX, editorY, shift)

	case tea.MouseActionMotion:
		if msg.Button == tea.MouseButtonLeft {
			editorX := msg.X - a.sidebar.Width()
			editorY := msg.Y - 1
			a.editor.HandleDrag(editorX, editorY)
		}

	case tea.MouseActionRelease:
		// Double click detection would go here
		// For now, single click only

	}

	// Handle scroll wheel
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if msg.X < a.sidebar.Width() && a.sidebar.IsVisible() {
			a.sidebar.ScrollUp(3)
		} else {
			a.editor.Scroll(-3)
		}
	case tea.MouseButtonWheelDown:
		if msg.X < a.sidebar.Width() && a.sidebar.IsVisible() {
			a.sidebar.ScrollDown(3)
		} else {
			a.editor.Scroll(3)
		}
	}

	return a, nil
}

// executeCommand executes a command by ID.
func (a *App) executeCommand(id string) (tea.Model, tea.Cmd) {
	switch id {
	case "file.save":
		return a.save()
	case "file.new":
		a.editor.NewFile()
		a.showMessage("New file", ui.MessageInfo)
	case "app.quit":
		return a.quit()
	case "edit.undo":
		a.editor.Undo()
	case "edit.redo":
		a.editor.Redo()
	case "edit.cut":
		text := a.editor.Cut()
		a.copyToClipboard(text)
	case "edit.copy":
		text := a.editor.Copy()
		a.copyToClipboard(text)
	case "edit.paste":
		text := a.pasteFromClipboard()
		a.editor.Paste(text)
	case "edit.selectAll":
		a.editor.SelectAll()
	case "edit.duplicateLine":
		a.editor.DuplicateLine()
	case "edit.deleteLine":
		a.editor.DeleteLine()
	case "edit.moveLineUp":
		a.editor.MoveLineUp()
	case "edit.moveLineDown":
		a.editor.MoveLineDown()
	case "search.find":
		a.searchBar.Show()
		a.focus = FocusSearchBar
		a.handleResize(a.width, a.height)
	case "search.replace":
		a.searchBar.ShowReplace()
		a.focus = FocusSearchBar
		a.handleResize(a.width, a.height)
	case "search.findNext":
		if a.searchBar.SearchText() != "" {
			a.editor.Find(a.searchBar.SearchText(), a.searchBar.IsCaseSensitive())
		}
	case "search.findPrevious":
		if a.searchBar.SearchText() != "" {
			a.editor.FindPrevious(a.searchBar.SearchText(), a.searchBar.IsCaseSensitive())
		}
	case "nav.goToLine":
		a.searchBar.ShowGoToLine()
		a.focus = FocusSearchBar
		a.handleResize(a.width, a.height)
	case "view.toggleSidebar":
		a.sidebar.Toggle()
		a.handleResize(a.width, a.height)
	case "view.commandPalette":
		a.commandPalette.Show()
		a.focus = FocusCommandPalette
	}
	return a, nil
}

// save saves the current file.
func (a *App) save() (tea.Model, tea.Cmd) {
	if a.editor.Filepath() == "" {
		a.showMessage("No file name - use Save As", ui.MessageWarning)
		return a, nil
	}

	if err := a.editor.Save(); err != nil {
		a.showMessage("Error saving: "+err.Error(), ui.MessageError)
	} else {
		a.showMessage("Saved "+filepath.Base(a.editor.Filepath()), ui.MessageInfo)
	}
	return a, nil
}

// quit attempts to quit the application.
func (a *App) quit() (tea.Model, tea.Cmd) {
	if a.editor.Modified() {
		a.showMessage("Unsaved changes! Press Ctrl+Q again to quit without saving", ui.MessageWarning)
		// In a real implementation, we'd track double-press
		// For now, just warn and allow quit
	}
	a.quitting = true
	return a, tea.Quit
}

// showMessage displays a status message.
func (a *App) showMessage(msg string, msgType ui.MessageType) {
	a.message = msg
	a.messageTime = time.Now()
	a.statusBar.SetMessage(msg, msgType)
}

// copyToClipboard copies text to the system clipboard.
func (a *App) copyToClipboard(text string) {
	if a.clipboardInit {
		clipboard.Write(clipboard.FmtText, []byte(text))
	}
}

// pasteFromClipboard retrieves text from the system clipboard.
func (a *App) pasteFromClipboard() string {
	if a.clipboardInit {
		data := clipboard.Read(clipboard.FmtText)
		return string(data)
	}
	return ""
}

// View implements tea.Model.
func (a *App) View() string {
	if a.quitting {
		return ""
	}

	var sections []string

	// Title bar
	a.titleBar.SetFile(a.editor.Filepath(), a.editor.Modified())
	sections = append(sections, a.titleBar.View())

	// Main content area (sidebar + editor)
	var mainContent string
	if a.sidebar.IsVisible() {
		sidebarView := a.sidebar.View()
		editorView := a.editor.View()

		// Join sidebar and editor horizontally
		sidebarLines := strings.Split(sidebarView, "\n")
		editorLines := strings.Split(editorView, "\n")

		// Ensure equal number of lines
		maxLines := len(sidebarLines)
		if len(editorLines) > maxLines {
			maxLines = len(editorLines)
		}

		var combinedLines []string
		for i := 0; i < maxLines; i++ {
			sLine := ""
			if i < len(sidebarLines) {
				sLine = sidebarLines[i]
			}
			eLine := ""
			if i < len(editorLines) {
				eLine = editorLines[i]
			}
			combinedLines = append(combinedLines, sLine+eLine)
		}
		mainContent = strings.Join(combinedLines, "\n")
	} else {
		mainContent = a.editor.View()
	}
	sections = append(sections, mainContent)

	// Search bar (if visible)
	if a.searchBar.IsVisible() {
		sections = append(sections, a.searchBar.View())
	}

	// Status bar
	a.statusBar.SetPosition(a.editor.CursorLine(), a.editor.CursorColumn(), a.editor.LineCount())
	a.statusBar.SetLanguage(a.editor.Language())
	a.statusBar.SetEncoding(a.editor.Encoding())
	a.statusBar.SetLineEnding(a.editor.LineEnding())
	sections = append(sections, a.statusBar.View())

	// Main view
	view := strings.Join(sections, "\n")

	// Overlay command palette if visible
	if a.commandPalette.IsVisible() {
		paletteView := a.commandPalette.View()
		view = a.overlayView(view, paletteView)
	}

	return view
}

// overlayView overlays a smaller view on top of the main view.
func (a *App) overlayView(base, overlay string) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	// Calculate position (center horizontally, near top vertically)
	overlayWidth := 0
	for _, line := range overlayLines {
		w := lipgloss.Width(line)
		if w > overlayWidth {
			overlayWidth = w
		}
	}

	startX := (a.width - overlayWidth) / 2
	startY := 3 // Below title bar

	if startX < 0 {
		startX = 0
	}

	// Overlay the palette
	for i, overlayLine := range overlayLines {
		lineIdx := startY + i
		if lineIdx < len(baseLines) {
			baseLine := baseLines[lineIdx]
			baseRunes := []rune(baseLine)

			// Pad base line if needed
			for len(baseRunes) < startX {
				baseRunes = append(baseRunes, ' ')
			}

			// Insert overlay
			overlayRunes := []rune(overlayLine)
			newLine := string(baseRunes[:startX]) + string(overlayRunes)

			// Add rest of base line if there's room
			endX := startX + len(overlayRunes)
			if endX < len(baseRunes) {
				newLine += string(baseRunes[endX:])
			}

			baseLines[lineIdx] = newLine
		}
	}

	return strings.Join(baseLines, "\n")
}

// Run runs the application.
func Run(filepath string, line int) error {
	app := New()

	// Load file if specified
	if filepath != "" {
		if err := app.LoadFile(filepath); err != nil {
			// If file doesn't exist, create new buffer with that path
			if os.IsNotExist(err) {
				app.editor.NewFile()
				app.editor.Buffer().SetFilepath(filepath)
				// Try to load parent directory into sidebar
				dir := filepath
				if idx := strings.LastIndex(filepath, string(os.PathSeparator)); idx > 0 {
					dir = filepath[:idx]
				} else {
					dir = "."
				}
				app.sidebar.LoadDirectory(dir)
			} else {
				return fmt.Errorf("error loading file: %w", err)
			}
		}

		// Go to line if specified
		if line > 0 {
			app.GoToLine(line)
		}
	} else {
		// Load current directory into sidebar
		cwd, _ := os.Getwd()
		app.sidebar.LoadDirectory(cwd)
	}

	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}
