# vex Architecture

This document describes the technical architecture of vex.

## Overview

vex is built using the Bubble Tea TUI framework with a component-based architecture. Each major feature is implemented as a separate component that can be composed together.

## Technology Stack

- **Language**: Go 1.22+
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) - MIT
- **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss) - MIT
- **Syntax Highlighting**: [Chroma](https://github.com/alecthomas/chroma) - MIT
- **Clipboard**: [golang.design/x/clipboard](https://golang.design/x/clipboard) - MIT
- **Fuzzy Matching**: [fuzzy](https://github.com/sahilm/fuzzy) - MIT

## Component Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                          App                                 │
│  (Orchestrates all components, handles global state)         │
├─────────────────────────────────────────────────────────────┤
│  ┌────────────┐  ┌────────────────────────────────────────┐ │
│  │  TitleBar  │  │               TabBar                    │ │
│  └────────────┘  └────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐  ┌──────────────────────────────────────┐ │
│  │   Sidebar    │  │              Editor                   │ │
│  │              │  │                                       │ │
│  │  ┌────────┐  │  │  ┌────────────────────────────────┐ │ │
│  │  │FileTree│  │  │  │          TabManager            │ │ │
│  │  └────────┘  │  │  │  ┌──────────────────────────┐ │ │ │
│  │              │  │  │  │       TabState(s)        │ │ │ │
│  │              │  │  │  │ Buffer, Cursor, Selection│ │ │ │
│  │              │  │  │  │ History, Highlighter     │ │ │ │
│  │              │  │  │  └──────────────────────────┘ │ │ │
│  │              │  │  └────────────────────────────────┘ │ │
│  └──────────────┘  └──────────────────────────────────────┘ │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│  ┌───────────────┐  ┌─────────────────────┐                  │
│  │CommandPalette │  │     SearchBar       │                  │
│  └───────────────┘  └─────────────────────┘                  │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                      StatusBar                          │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### Buffer (`internal/editor/buffer.go`)

The buffer uses a **Gap Buffer** data structure for efficient text editing:

```
Before inserting at position 5:
[H][e][l][l][o][_GAP_GAP_GAP_][ ][W][o][r][l][d]
              ^gap start      ^gap end

After inserting "X" at position 5:
[H][e][l][l][o][X][_GAP_GAP_][W][o][r][l][d]
                 ^gap start  ^gap end
```

**Key features:**
- O(1) insertions and deletions at cursor position
- Automatic gap expansion when needed
- Line index caching for fast line operations
- UTF-8/rune-based for proper Unicode handling

### Cursor (`internal/editor/cursor.go`)

Manages cursor position with:
- Line/column tracking
- Preferred column for vertical movement
- Word-wise navigation
- Buffer boundary clamping

### Selection (`internal/editor/selection.go`)

Handles text selection:
- Start/end positions (anchor and cursor)
- Normalized range access (start always before end)
- Word and line selection
- Selection-aware editing operations

### History (`internal/editor/history.go`)

Implements undo/redo with:
- Action types: Insert, Delete, Replace
- Automatic action merging for consecutive typing
- Configurable stack size
- Cursor position restoration

### TabState (`internal/editor/tabstate.go`)

Encapsulates all state for a single open file:
- Buffer (file content)
- Cursor position
- Selection state
- Undo/redo history
- Syntax highlighter
- Scroll position (X/Y)

### TabManager (`internal/editor/tabmanager.go`)

Manages multiple open tabs:
- Tab list and active tab tracking
- Add/close/switch tabs
- Find tabs by file path
- Track modified tabs
- Save all functionality

### Editor (`internal/editor/editor.go`)

The main editing component that coordinates:
- TabManager for multi-file editing
- Buffer operations (delegated to active tab)
- Cursor movement
- Selection management
- History tracking
- Syntax highlighting
- View rendering

### Syntax Highlighter (`internal/syntax/highlighter.go`)

Chroma-based syntax highlighting:
- Language detection from file extension
- 200+ language support
- Theme customization
- Efficient line-by-line highlighting

## UI Components

### TitleBar (`internal/ui/titlebar.go`)
- Displays filename and modified indicator
- Centered title with app name

### TabBar (`internal/ui/tabbar.go`)
- Horizontal tab display (shown when multiple tabs open)
- Tab name with modified indicator (●)
- Active tab highlighting
- Click to switch tabs

### StatusBar (`internal/ui/statusbar.go`)
- Cursor position (line, column)
- Language detection
- Encoding and line ending info
- Status messages

### Sidebar (`internal/ui/sidebar.go`)
- File tree navigation
- Directory expansion/collapse
- File selection and opening
- Toggle visibility

### CommandPalette (`internal/ui/commandpalette.go`)
- Fuzzy search over commands
- Keybinding display
- Category organization

### SearchBar (`internal/ui/searchbar.go`)
- Find mode
- Find and replace mode
- Go to line mode
- Case sensitivity toggle

## Data Flow

```
User Input (KeyMsg/MouseMsg)
        │
        ▼
   ┌─────────┐
   │   App   │ ─────► Route to focused component
   └─────────┘
        │
        ▼
  ┌──────────┐
  │ Component│ ─────► Process input, update state
  └──────────┘
        │
        ▼
  ┌──────────┐
  │  View()  │ ─────► Render to string
  └──────────┘
        │
        ▼
   Terminal Output
```

## Key Design Decisions

### No Modal Editing
Unlike vim, vex is always in edit mode. This aligns with VSCode/Sublime behavior and reduces cognitive load.

### Gap Buffer
Chosen over rope or piece table for:
- Simpler implementation
- Good performance for typical editing patterns
- Efficient memory usage for moderate file sizes

### Component Composition
Each UI element is a separate component with its own state and rendering. The App component orchestrates them.

### Bubble Tea Model
Following the Elm architecture:
- `Init()` - Initialize state
- `Update()` - Handle messages, return new state
- `View()` - Render to string

## Performance Considerations

1. **Lazy highlighting**: Only visible lines are styled
2. **Line index caching**: Buffer maintains line offsets
3. **Efficient rendering**: Only changed areas are re-rendered
4. **Gap buffer efficiency**: O(1) at cursor, amortized for gap moves

## Future Improvements

- Split views
- Language server protocol (LSP) support
- Plugin system
- Configuration file support
- Themes
