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
	tabBar         *ui.TabBar
	statusBar      *ui.StatusBar
	sidebar        *ui.Sidebar
	commandPalette *ui.CommandPalette
	searchBar      *ui.SearchBar

	// Configuration
	config      *config.Config
	keyBindings *keybindings.KeyBindings

	// State
	focus           FocusArea
	width           int
	height          int
	quitting        bool
	pendingQuit     bool // True when waiting for quit confirmation
	pendingCloseTab bool // True when waiting for close tab confirmation
	message         string
	messageTime     time.Time

	// Clipboard
	clipboardInit bool
}

// New creates a new App instance.
func New() *App {
	cfg, _ := config.Load()

	app := &App{
		editor:         editor.NewEditor(),
		titleBar:       ui.NewTitleBar(),
		tabBar:         ui.NewTabBar(),
		statusBar:      ui.NewStatusBar(),
		sidebar:        ui.NewSidebar(),
		commandPalette: ui.NewCommandPalette(),
		searchBar:      ui.NewSearchBar(),
		config:         cfg,
		keyBindings:    keybindings.NewKeyBindings(),
		focus:          FocusEditor,
	}

	// Initialize clipboard with panic recovery for headless systems
	// (clipboard.Init may panic when CGO_ENABLED=0 or no X11/Wayland)
	func() {
		defer func() {
			recover()
		}()
		if err := clipboard.Init(); err == nil {
			app.clipboardInit = true
		}
	}()

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
	sidebarWidth := a.sidebar.Width() // Returns 0 when hidden
	editorWidth := width - sidebarWidth
	tabBarHeight := a.tabBar.Height()
	editorHeight := height - 2 - tabBarHeight // title bar + status bar + tab bar

	// Update component sizes
	a.titleBar.SetWidth(width)
	a.tabBar.SetWidth(width)
	a.statusBar.SetWidth(width)
	// Only update sidebar height, preserve width (SetSize only for height)
	a.sidebar.SetHeight(height - 2 - tabBarHeight) // Exclude title, status, and tab bar
	a.editor.SetSize(editorWidth, editorHeight)
	a.commandPalette.SetSize(width, height)
	a.searchBar.SetWidth(width)
}

// handleKeyPress handles keyboard input.
func (a *App) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle pending quit confirmation
	if a.pendingQuit {
		switch msg.Type {
		case tea.KeyEsc:
			a.cancelQuit()
			return a, nil
		case tea.KeyCtrlS:
			return a.saveAndQuit()
		case tea.KeyCtrlQ:
			// Quit without saving (handled in quit())
			return a.quit()
		default:
			// Any other key cancels the quit
			a.cancelQuit()
			// Don't return - let the key be processed normally
		}
	}

	// Handle pending close tab confirmation
	if a.pendingCloseTab {
		switch msg.Type {
		case tea.KeyEsc:
			a.pendingCloseTab = false
			a.showMessage("Schließen abgebrochen", ui.MessageInfo)
			return a, nil
		case tea.KeyCtrlS:
			// Save then close
			if _, cmd := a.save(); cmd == nil {
				a.pendingCloseTab = false
				return a.closeTab()
			}
			return a, nil
		case tea.KeyCtrlW:
			// Force close (handled in closeTab)
			return a.closeTab()
		default:
			// Any other key cancels the close
			a.pendingCloseTab = false
			a.showMessage("Schließen abgebrochen", ui.MessageInfo)
			// Don't return - let the key be processed normally
		}
	}

	// Handle Escape first - closes overlays and cancels pending actions
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
	case keybindings.ActionOpen:
		a.searchBar.ShowOpen()
		a.focus = FocusSearchBar
		a.handleResize(a.width, a.height)
		return a, nil

	// Tab operations
	case keybindings.ActionCloseTab:
		return a.closeTab()
	case keybindings.ActionNextTab:
		a.editor.TabManager().NextTab()
		a.highlightDirty()
		a.handleResize(a.width, a.height)
		return a, nil
	case keybindings.ActionPrevTab:
		a.editor.TabManager().PrevTab()
		a.highlightDirty()
		a.handleResize(a.width, a.height)
		return a, nil
	case keybindings.ActionSaveAll:
		return a.saveAll()

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
		// Reset focus to editor if sidebar is now hidden
		if !a.sidebar.IsVisible() {
			a.focus = FocusEditor
		}
		a.handleResize(a.width, a.height)
		return a, nil
	case keybindings.ActionCommandPalette:
		a.commandPalette.Show()
		a.focus = FocusCommandPalette
		return a, nil
	case keybindings.ActionFocusExplorer:
		// Toggle focus between editor and explorer
		if a.focus == FocusEditor {
			// Show sidebar if hidden, then focus it
			if !a.sidebar.IsVisible() {
				a.sidebar.Toggle()
				a.handleResize(a.width, a.height)
			}
			a.focus = FocusSidebar
			a.showMessage("Fokus: Explorer", ui.MessageInfo)
		} else {
			a.focus = FocusEditor
			a.showMessage("Fokus: Editor", ui.MessageInfo)
		}
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

	// Handle space key (Bubble Tea treats it as special key, not rune)
	if msg.Type == tea.KeySpace {
		a.editor.InsertRune(' ')
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
		switch a.searchBar.Mode() {
		case ui.SearchModeGoToLine:
			lineNum := a.searchBar.LineNumber()
			if lineNum > 0 {
				a.editor.GoToLine(lineNum)
			}
			a.searchBar.Hide()
			a.focus = FocusEditor
			a.handleResize(a.width, a.height)
		case ui.SearchModeSaveAs:
			filePath := a.searchBar.FilePath()
			if filePath != "" {
				if err := a.editor.SaveAs(filePath); err != nil {
					a.showMessage("Fehler beim Speichern: "+err.Error(), ui.MessageError)
				} else {
					a.showMessage("Gespeichert: "+filepath.Base(filePath), ui.MessageInfo)
					a.sidebar.Refresh()
				}
			}
			a.searchBar.Hide()
			a.focus = FocusEditor
			a.handleResize(a.width, a.height)
		case ui.SearchModeOpen:
			filePath := a.searchBar.FilePath()
			if filePath != "" {
				if err := a.editor.LoadFile(filePath); err != nil {
					a.showMessage("Fehler beim Öffnen: "+err.Error(), ui.MessageError)
				} else {
					a.showMessage("Geöffnet: "+filepath.Base(filePath), ui.MessageInfo)
				}
			}
			a.searchBar.Hide()
			a.focus = FocusEditor
			a.handleResize(a.width, a.height)
		default:
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
		// Live search (only for find/replace modes)
		mode := a.searchBar.Mode()
		if mode == ui.SearchModeFind || mode == ui.SearchModeReplace {
			text := a.searchBar.SearchText()
			if text != "" {
				a.editor.Find(text, a.searchBar.IsCaseSensitive())
			}
		}
	case tea.KeySpace:
		a.searchBar.Input(" ")
	}
	return a, nil
}

// handleSidebarKey handles key input when sidebar is focused.
func (a *App) handleSidebarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check for global shortcuts first
	action := a.keyBindings.Lookup(msg)
	switch action {
	case keybindings.ActionToggleSidebar:
		a.sidebar.Toggle()
		if !a.sidebar.IsVisible() {
			a.focus = FocusEditor
		}
		a.handleResize(a.width, a.height)
		return a, nil
	case keybindings.ActionQuit:
		return a.quit()
	case keybindings.ActionCommandPalette:
		a.commandPalette.Show()
		a.focus = FocusCommandPalette
		return a, nil
	case keybindings.ActionFocusExplorer:
		// Toggle back to editor
		a.focus = FocusEditor
		a.showMessage("Fokus: Editor (von Sidebar)", ui.MessageInfo)
		return a, nil
	}

	// Sidebar-specific keys
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
	case tea.KeyLeft:
		// Collapse current directory or move focus to editor
		a.focus = FocusEditor
	}
	return a, nil
}

// handleMouse handles mouse events.
func (a *App) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	tabBarHeight := a.tabBar.Height()

	switch msg.Action {
	case tea.MouseActionPress:
		// Check if click is in tab bar (row 1 if visible)
		if tabBarHeight > 0 && msg.Y == 1 {
			tabIdx := a.tabBar.HandleClick(msg.X)
			if tabIdx >= 0 {
				a.editor.TabManager().SwitchTab(tabIdx)
				a.highlightDirty()
				a.handleResize(a.width, a.height)
			}
			return a, nil
		}

		// Adjust Y for tab bar
		adjustedY := msg.Y - 1 - tabBarHeight // Adjust for title bar and tab bar
		if adjustedY < 0 {
			return a, nil // Click on title bar or tab bar, ignore
		}

		// Check if click is in sidebar
		if a.sidebar.IsVisible() && msg.X < a.sidebar.Width() {
			a.focus = FocusSidebar
			path := a.sidebar.HandleClick(adjustedY)
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
		editorY := adjustedY
		shift := msg.Ctrl // Bubble Tea doesn't have Shift detection in mouse, use Ctrl as workaround
		a.editor.HandleClick(editorX, editorY, shift)

	case tea.MouseActionMotion:
		if msg.Button == tea.MouseButtonLeft {
			editorX := msg.X - a.sidebar.Width()
			editorY := msg.Y - 1 - tabBarHeight
			if editorY >= 0 {
				a.editor.HandleDrag(editorX, editorY)
			}
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
	case "file.saveAs":
		a.searchBar.ShowSaveAs(a.editor.Filepath())
		a.focus = FocusSearchBar
		a.handleResize(a.width, a.height)
	case "file.open":
		a.searchBar.ShowOpen()
		a.focus = FocusSearchBar
		a.handleResize(a.width, a.height)
	case "file.close":
		a.editor.NewFile()
		a.showMessage("Datei geschlossen", ui.MessageInfo)
	case "nav.goToStart":
		a.editor.MoveCursor("bufferStart", false)
	case "nav.goToEnd":
		a.editor.MoveCursor("bufferEnd", false)
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

// saveAll saves all modified tabs.
func (a *App) saveAll() (tea.Model, tea.Cmd) {
	if err := a.editor.TabManager().SaveAll(); err != nil {
		a.showMessage("Fehler beim Speichern: "+err.Error(), ui.MessageError)
	} else {
		a.showMessage("Alle Dateien gespeichert", ui.MessageInfo)
	}
	return a, nil
}

// closeTab closes the current tab.
func (a *App) closeTab() (tea.Model, tea.Cmd) {
	tab := a.editor.TabManager().ActiveTab()
	if tab == nil {
		return a, nil
	}

	if tab.Modified() {
		// Second Ctrl+W forces close
		if a.pendingCloseTab {
			a.pendingCloseTab = false
			a.editor.TabManager().CloseActiveTab()
			a.highlightDirty()
			a.handleResize(a.width, a.height)
			a.showMessage("Tab verworfen", ui.MessageInfo)
			return a, nil
		}

		// First Ctrl+W - show confirmation
		a.pendingCloseTab = true
		a.showMessage("Ungespeicherte Änderungen! Ctrl+S: Speichern | Ctrl+W: Verwerfen | Esc: Abbrechen", ui.MessageWarning)
		return a, nil
	}

	a.pendingCloseTab = false
	a.editor.TabManager().CloseActiveTab()
	a.highlightDirty()
	a.handleResize(a.width, a.height)
	a.showMessage("Tab geschlossen", ui.MessageInfo)
	return a, nil
}

// highlightDirty marks highlighting as needing refresh.
func (a *App) highlightDirty() {
	a.editor.MarkHighlightDirty()
}

// updateTabBar updates the tab bar with current tab information.
func (a *App) updateTabBar() {
	tm := a.editor.TabManager()
	tabs := tm.Tabs()
	activeIdx := tm.ActiveIndex()

	var tabInfos []ui.TabInfo
	for i, tab := range tabs {
		tabInfos = append(tabInfos, ui.TabInfo{
			Name:     tab.Name(),
			Path:     tab.Filepath(),
			Modified: tab.Modified(),
			Active:   i == activeIdx,
		})
	}
	a.tabBar.SetTabs(tabInfos)
}

// updateOpenEditors updates the sidebar open editors section.
func (a *App) updateOpenEditors() {
	tm := a.editor.TabManager()
	tabs := tm.Tabs()
	activeIdx := tm.ActiveIndex()

	var openTabs []ui.OpenTabInfo
	for i, tab := range tabs {
		openTabs = append(openTabs, ui.OpenTabInfo{
			Name:     tab.Name(),
			Path:     tab.Filepath(),
			Modified: tab.Modified(),
			IsNew:    tab.Filepath() == "", // New file if no path
			Active:   i == activeIdx,
		})
	}
	a.sidebar.SetOpenTabs(openTabs)
}

// quit attempts to quit the application.
func (a *App) quit() (tea.Model, tea.Cmd) {
	// Check if any tab has unsaved changes
	if !a.editor.TabManager().IsModified() {
		a.quitting = true
		return a, tea.Quit
	}

	// Unsaved changes - check if already pending
	if a.pendingQuit {
		// Second Ctrl+Q - quit without saving
		a.quitting = true
		return a, tea.Quit
	}

	// First Ctrl+Q with unsaved changes - show warning
	a.pendingQuit = true
	modifiedCount := len(a.editor.TabManager().GetModifiedTabs())
	a.showMessage(fmt.Sprintf("%d ungespeicherte Tab(s)! Ctrl+S: Speichern & Beenden | Ctrl+Q: Verwerfen | Esc: Abbrechen", modifiedCount), ui.MessageWarning)
	return a, nil
}

// saveAndQuit saves the file and quits.
func (a *App) saveAndQuit() (tea.Model, tea.Cmd) {
	if a.editor.Filepath() == "" {
		a.showMessage("Kein Dateiname - nutze Save As", ui.MessageWarning)
		a.pendingQuit = false
		return a, nil
	}

	if err := a.editor.Save(); err != nil {
		a.showMessage("Fehler beim Speichern: "+err.Error(), ui.MessageError)
		a.pendingQuit = false
		return a, nil
	}

	a.quitting = true
	return a, tea.Quit
}

// cancelQuit cancels the pending quit.
func (a *App) cancelQuit() {
	a.pendingQuit = false
	a.showMessage("Beenden abgebrochen", ui.MessageInfo)
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

	// Tab bar (only shown when multiple tabs)
	a.updateTabBar()
	if tabBarView := a.tabBar.View(); tabBarView != "" {
		sections = append(sections, tabBarView)
	}

	// Update sidebar modified indicators and open editors section
	a.sidebar.SetModifiedFiles(a.editor.TabManager().GetModifiedPaths())
	a.updateOpenEditors()

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

// overlayView overlays a smaller view on top of the main view using lipgloss.Place.
func (a *App) overlayView(base, overlay string) string {
	// Use lipgloss.Place to center the overlay horizontally and position it near the top
	// First, place the overlay in a full-screen sized box
	placed := lipgloss.Place(
		a.width,
		a.height,
		lipgloss.Center, // horizontal center
		lipgloss.Top,    // vertical top
		overlay,
		lipgloss.WithWhitespaceChars(" "),
	)

	// Now we need to composite: show base where overlay is transparent
	// Since lipgloss.Place fills with spaces, we merge line by line
	baseLines := strings.Split(base, "\n")
	placedLines := strings.Split(placed, "\n")

	result := make([]string, len(baseLines))
	for i, baseLine := range baseLines {
		if i < len(placedLines) {
			// Check if placed line has content (non-space characters after stripping ANSI)
			placedLine := placedLines[i]
			if strings.TrimSpace(stripAnsi(placedLine)) != "" {
				result[i] = placedLine
			} else {
				result[i] = baseLine
			}
		} else {
			result[i] = baseLine
		}
	}

	return strings.Join(result, "\n")
}

// stripAnsi removes ANSI escape codes from a string for comparison.
func stripAnsi(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
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
