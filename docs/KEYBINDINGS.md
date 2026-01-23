# vex Keybindings Reference

vex uses VSCode-style keybindings for a familiar editing experience. All shortcuts work in edit mode - there are no modal modes like in vim.

## File Operations

| Shortcut | Action | Description |
|----------|--------|-------------|
| Ctrl+S | Save | Save current file |
| Ctrl+Shift+S | Save As | Save to new file |
| Ctrl+N | New | Create new file |
| Ctrl+O | Open | Open file |
| Ctrl+W | Close | Close current file |
| Ctrl+Q | Quit | Exit editor |

## Editing

| Shortcut | Action | Description |
|----------|--------|-------------|
| Ctrl+Z | Undo | Undo last action |
| Ctrl+Y | Redo | Redo last undone action |
| Ctrl+Shift+Z | Redo | Alternative redo |
| Ctrl+X | Cut | Cut selection or line |
| Ctrl+C | Copy | Copy selection or line |
| Ctrl+V | Paste | Paste from clipboard |
| Ctrl+A | Select All | Select entire document |
| Ctrl+D | Duplicate | Duplicate line or selection |
| Ctrl+L | Select Line | Select current line |
| Ctrl+Shift+K | Delete Line | Delete current line |
| Tab | Indent | Insert tab/spaces |
| Shift+Tab | Outdent | Remove indentation |

## Line Operations

| Shortcut | Action | Description |
|----------|--------|-------------|
| Alt+Up | Move Line Up | Move current line up |
| Alt+Down | Move Line Down | Move current line down |
| Ctrl+Enter | Insert Line Below | Insert new line below |
| Ctrl+Shift+Enter | Insert Line Above | Insert new line above |

## Navigation

| Shortcut | Action | Description |
|----------|--------|-------------|
| Arrow Keys | Move Cursor | Move by character/line |
| Ctrl+Left | Word Left | Move to previous word |
| Ctrl+Right | Word Right | Move to next word |
| Home | Line Start | Go to line start |
| End | Line End | Go to line end |
| Ctrl+Home | File Start | Go to file start |
| Ctrl+End | File End | Go to file end |
| Ctrl+G | Go to Line | Jump to specific line |
| Page Up | Page Up | Scroll up one page |
| Page Down | Page Down | Scroll down one page |

## Selection

| Shortcut | Action | Description |
|----------|--------|-------------|
| Shift+Arrow | Extend Selection | Select while moving |
| Shift+Ctrl+Left | Select Word Left | Select previous word |
| Shift+Ctrl+Right | Select Word Right | Select next word |
| Shift+Home | Select to Line Start | Select to line start |
| Shift+End | Select to Line End | Select to line end |
| Shift+Ctrl+Home | Select to File Start | Select to file start |
| Shift+Ctrl+End | Select to File End | Select to file end |

## Search & Replace

| Shortcut | Action | Description |
|----------|--------|-------------|
| Ctrl+F | Find | Open search bar |
| Ctrl+H | Replace | Open find and replace |
| F3 | Find Next | Go to next match |
| Shift+F3 | Find Previous | Go to previous match |
| Enter | Find Next | (in search bar) |
| Escape | Close | Close search bar |

## View & UI

| Shortcut | Action | Description |
|----------|--------|-------------|
| Ctrl+B | Toggle Sidebar | Show/hide file explorer |
| Ctrl+P | Command Palette | Open command palette |
| Escape | Close Overlay | Close palette/search/selection |

## Command Palette

When the command palette is open:

| Key | Action |
|-----|--------|
| Type | Filter commands |
| Up/Down | Navigate list |
| Enter | Execute command |
| Escape | Close palette |

## Mouse

| Action | Description |
|--------|-------------|
| Click | Position cursor |
| Double-click | Select word |
| Triple-click | Select line |
| Drag | Select text |
| Scroll wheel | Scroll content |
| Sidebar click | Open file |

## Tips

1. **No modes**: Unlike vim, vex is always in insert mode. Just start typing.

2. **Quick save**: Ctrl+S saves immediately. No confirmation needed.

3. **Command palette**: Can't remember a shortcut? Press Ctrl+P and search for the command.

4. **Line operations**: Alt+Arrow for moving lines is very useful for reorganizing code.

5. **Word navigation**: Ctrl+Left/Right jumps by words, much faster than arrow keys.
