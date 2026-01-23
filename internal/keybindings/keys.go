package keybindings

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Action represents an editor action triggered by a keybinding.
type Action string

const (
	// File actions
	ActionSave      Action = "file.save"
	ActionSaveAs    Action = "file.saveAs"
	ActionNew       Action = "file.new"
	ActionOpen      Action = "file.open"
	ActionClose     Action = "file.close"
	ActionQuit      Action = "app.quit"

	// Edit actions
	ActionUndo          Action = "edit.undo"
	ActionRedo          Action = "edit.redo"
	ActionCut           Action = "edit.cut"
	ActionCopy          Action = "edit.copy"
	ActionPaste         Action = "edit.paste"
	ActionSelectAll     Action = "edit.selectAll"
	ActionDuplicateLine Action = "edit.duplicateLine"
	ActionDeleteLine    Action = "edit.deleteLine"
	ActionMoveLineUp    Action = "edit.moveLineUp"
	ActionMoveLineDown  Action = "edit.moveLineDown"

	// Navigation actions
	ActionMoveLeft       Action = "nav.moveLeft"
	ActionMoveRight      Action = "nav.moveRight"
	ActionMoveUp         Action = "nav.moveUp"
	ActionMoveDown       Action = "nav.moveDown"
	ActionMoveWordLeft   Action = "nav.moveWordLeft"
	ActionMoveWordRight  Action = "nav.moveWordRight"
	ActionMoveLineStart  Action = "nav.moveLineStart"
	ActionMoveLineEnd    Action = "nav.moveLineEnd"
	ActionMoveBufferStart Action = "nav.moveBufferStart"
	ActionMoveBufferEnd  Action = "nav.moveBufferEnd"
	ActionPageUp         Action = "nav.pageUp"
	ActionPageDown       Action = "nav.pageDown"
	ActionGoToLine       Action = "nav.goToLine"

	// Selection actions
	ActionSelectLeft       Action = "select.left"
	ActionSelectRight      Action = "select.right"
	ActionSelectUp         Action = "select.up"
	ActionSelectDown       Action = "select.down"
	ActionSelectWordLeft   Action = "select.wordLeft"
	ActionSelectWordRight  Action = "select.wordRight"
	ActionSelectLineStart  Action = "select.lineStart"
	ActionSelectLineEnd    Action = "select.lineEnd"
	ActionSelectLine       Action = "select.line"

	// Search actions
	ActionFind         Action = "search.find"
	ActionFindNext     Action = "search.findNext"
	ActionFindPrevious Action = "search.findPrevious"
	ActionReplace      Action = "search.replace"

	// View actions
	ActionToggleSidebar   Action = "view.toggleSidebar"
	ActionCommandPalette  Action = "view.commandPalette"

	// Text input
	ActionInsertNewline Action = "insert.newline"
	ActionInsertTab     Action = "insert.tab"
	ActionBackspace     Action = "edit.backspace"
	ActionDelete        Action = "edit.delete"

	// No action
	ActionNone Action = ""
)

// Binding represents a key binding.
type Binding struct {
	Key    tea.KeyType
	Runes  string
	Alt    bool
	Ctrl   bool
	Shift  bool
	Action Action
}

// KeyBindings manages keyboard shortcuts.
type KeyBindings struct {
	bindings []Binding
}

// NewKeyBindings creates a new keybindings manager with default VSCode-style bindings.
func NewKeyBindings() *KeyBindings {
	kb := &KeyBindings{}
	kb.loadDefaults()
	return kb
}

// loadDefaults loads the default VSCode-style keybindings.
func (kb *KeyBindings) loadDefaults() {
	kb.bindings = []Binding{
		// File operations
		{Key: tea.KeyCtrlS, Action: ActionSave},
		{Key: tea.KeyCtrlN, Action: ActionNew},
		{Key: tea.KeyCtrlO, Action: ActionOpen},
		{Key: tea.KeyCtrlW, Action: ActionClose},
		{Key: tea.KeyCtrlQ, Action: ActionQuit},

		// Edit operations
		{Key: tea.KeyCtrlZ, Action: ActionUndo},
		{Key: tea.KeyCtrlY, Action: ActionRedo},
		{Key: tea.KeyCtrlX, Action: ActionCut},
		{Key: tea.KeyCtrlC, Action: ActionCopy},
		{Key: tea.KeyCtrlV, Action: ActionPaste},
		{Key: tea.KeyCtrlA, Action: ActionSelectAll},
		{Key: tea.KeyCtrlD, Action: ActionDuplicateLine},
		{Key: tea.KeyCtrlK, Ctrl: true, Shift: true, Action: ActionDeleteLine},
		{Key: tea.KeyCtrlL, Action: ActionSelectLine},

		// Navigation
		{Key: tea.KeyLeft, Action: ActionMoveLeft},
		{Key: tea.KeyRight, Action: ActionMoveRight},
		{Key: tea.KeyUp, Action: ActionMoveUp},
		{Key: tea.KeyDown, Action: ActionMoveDown},
		{Key: tea.KeyHome, Action: ActionMoveLineStart},
		{Key: tea.KeyEnd, Action: ActionMoveLineEnd},
		{Key: tea.KeyPgUp, Action: ActionPageUp},
		{Key: tea.KeyPgDown, Action: ActionPageDown},
		{Key: tea.KeyCtrlG, Action: ActionGoToLine},

		// Word navigation (Ctrl+Arrow)
		{Key: tea.KeyCtrlLeft, Action: ActionMoveWordLeft},
		{Key: tea.KeyCtrlRight, Action: ActionMoveWordRight},

		// Buffer start/end (Ctrl+Home/End)
		{Key: tea.KeyCtrlHome, Action: ActionMoveBufferStart},
		{Key: tea.KeyCtrlEnd, Action: ActionMoveBufferEnd},

		// Selection (Shift+Arrow)
		{Key: tea.KeyShiftLeft, Action: ActionSelectLeft},
		{Key: tea.KeyShiftRight, Action: ActionSelectRight},
		{Key: tea.KeyShiftUp, Action: ActionSelectUp},
		{Key: tea.KeyShiftDown, Action: ActionSelectDown},

		// Search
		{Key: tea.KeyCtrlF, Action: ActionFind},
		{Key: tea.KeyCtrlH, Action: ActionReplace},
		{Key: tea.KeyF3, Action: ActionFindNext},

		// View
		{Key: tea.KeyCtrlB, Action: ActionToggleSidebar},
		{Key: tea.KeyCtrlP, Action: ActionCommandPalette},

		// Text input
		{Key: tea.KeyEnter, Action: ActionInsertNewline},
		{Key: tea.KeyTab, Action: ActionInsertTab},
		{Key: tea.KeyBackspace, Action: ActionBackspace},
		{Key: tea.KeyDelete, Action: ActionDelete},
	}
}

// Lookup looks up the action for a key message.
func (kb *KeyBindings) Lookup(msg tea.KeyMsg) Action {
	// Check for exact matches first
	for _, binding := range kb.bindings {
		if kb.matches(binding, msg) {
			return binding.Action
		}
	}

	return ActionNone
}

// matches checks if a binding matches a key message.
func (kb *KeyBindings) matches(binding Binding, msg tea.KeyMsg) bool {
	// Check key type
	if binding.Key != 0 && msg.Type == binding.Key {
		return true
	}

	// Check runes
	if binding.Runes != "" && msg.String() == binding.Runes {
		return true
	}

	return false
}

// GetBindingForAction returns the keybinding string for an action.
func (kb *KeyBindings) GetBindingForAction(action Action) string {
	for _, binding := range kb.bindings {
		if binding.Action == action {
			return kb.formatBinding(binding)
		}
	}
	return ""
}

// formatBinding formats a binding as a human-readable string.
func (kb *KeyBindings) formatBinding(binding Binding) string {
	var parts []string

	if binding.Ctrl {
		parts = append(parts, "Ctrl")
	}
	if binding.Alt {
		parts = append(parts, "Alt")
	}
	if binding.Shift {
		parts = append(parts, "Shift")
	}

	keyName := getKeyName(binding.Key)
	if keyName != "" {
		parts = append(parts, keyName)
	} else if binding.Runes != "" {
		parts = append(parts, binding.Runes)
	}

	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "+"
		}
		result += part
	}
	return result
}

// getKeyName returns the name of a key type.
func getKeyName(key tea.KeyType) string {
	switch key {
	case tea.KeyCtrlA:
		return "Ctrl+A"
	case tea.KeyCtrlB:
		return "Ctrl+B"
	case tea.KeyCtrlC:
		return "Ctrl+C"
	case tea.KeyCtrlD:
		return "Ctrl+D"
	case tea.KeyCtrlE:
		return "Ctrl+E"
	case tea.KeyCtrlF:
		return "Ctrl+F"
	case tea.KeyCtrlG:
		return "Ctrl+G"
	case tea.KeyCtrlH:
		return "Ctrl+H"
	case tea.KeyCtrlK:
		return "Ctrl+K"
	case tea.KeyCtrlL:
		return "Ctrl+L"
	case tea.KeyCtrlN:
		return "Ctrl+N"
	case tea.KeyCtrlO:
		return "Ctrl+O"
	case tea.KeyCtrlP:
		return "Ctrl+P"
	case tea.KeyCtrlQ:
		return "Ctrl+Q"
	case tea.KeyCtrlS:
		return "Ctrl+S"
	case tea.KeyCtrlV:
		return "Ctrl+V"
	case tea.KeyCtrlW:
		return "Ctrl+W"
	case tea.KeyCtrlX:
		return "Ctrl+X"
	case tea.KeyCtrlY:
		return "Ctrl+Y"
	case tea.KeyCtrlZ:
		return "Ctrl+Z"
	case tea.KeyEnter:
		return "Enter"
	case tea.KeyTab:
		return "Tab"
	case tea.KeyBackspace:
		return "Backspace"
	case tea.KeyDelete:
		return "Delete"
	case tea.KeyLeft:
		return "Left"
	case tea.KeyRight:
		return "Right"
	case tea.KeyUp:
		return "Up"
	case tea.KeyDown:
		return "Down"
	case tea.KeyHome:
		return "Home"
	case tea.KeyEnd:
		return "End"
	case tea.KeyPgUp:
		return "PageUp"
	case tea.KeyPgDown:
		return "PageDown"
	case tea.KeyF3:
		return "F3"
	case tea.KeyEsc:
		return "Esc"
	case tea.KeyShiftLeft:
		return "Shift+Left"
	case tea.KeyShiftRight:
		return "Shift+Right"
	case tea.KeyShiftUp:
		return "Shift+Up"
	case tea.KeyShiftDown:
		return "Shift+Down"
	case tea.KeyCtrlLeft:
		return "Ctrl+Left"
	case tea.KeyCtrlRight:
		return "Ctrl+Right"
	case tea.KeyCtrlHome:
		return "Ctrl+Home"
	case tea.KeyCtrlEnd:
		return "Ctrl+End"
	default:
		return ""
	}
}

// IsModifier returns true if the key message is just a modifier key.
func IsModifier(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyCtrlD:
		// These might be signals, but we handle them as actions
		return false
	}
	return false
}

// IsMovementKey returns true if the key is a movement key.
func IsMovementKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyLeft, tea.KeyRight, tea.KeyUp, tea.KeyDown,
		tea.KeyHome, tea.KeyEnd, tea.KeyPgUp, tea.KeyPgDown,
		tea.KeyCtrlLeft, tea.KeyCtrlRight, tea.KeyCtrlHome, tea.KeyCtrlEnd,
		tea.KeyShiftLeft, tea.KeyShiftRight, tea.KeyShiftUp, tea.KeyShiftDown:
		return true
	}
	return false
}
