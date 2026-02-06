# vex

A modern terminal text editor with intuitive keybindings.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-blue)

## Features

- VSCode-style keybindings (Ctrl+S, Ctrl+C/V, etc.)
- Syntax highlighting for 200+ languages
- Full mouse support
- Built-in file explorer
- Search & replace
- Fast and lightweight
- No modal editing - always in edit mode

## Installation

### Quick Install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/DDZ-DO/vex/main/install.sh | bash
```

### Go Install

```bash
go install github.com/DDZ-DO/vex/cmd/vex@latest
```

### From Source

```bash
git clone https://github.com/DDZ-DO/vex.git
cd vex
make install
```

## Usage

```bash
# Open editor with empty buffer
vex

# Open a file
vex file.go

# Open at specific line
vex file.go:42

# Show help
vex --help

# Show version
vex --version
```

## Keybindings

### File Operations

| Shortcut | Action |
|----------|--------|
| Ctrl+S | Save |
| Ctrl+N | New file |
| Ctrl+O | Open file |
| Ctrl+W | Close file |
| Ctrl+Q | Quit |

### Editing

| Shortcut | Action |
|----------|--------|
| Ctrl+Z | Undo |
| Ctrl+Y | Redo |
| Ctrl+X | Cut |
| Ctrl+C | Copy |
| Ctrl+V | Paste |
| Ctrl+A | Select all |
| Ctrl+D | Duplicate line |
| Ctrl+L | Delete line |
| Ctrl+L | Select line |
| Alt+Up/Down | Move line up/down |

### Navigation

| Shortcut | Action |
|----------|--------|
| Ctrl+G | Go to line |
| Ctrl+Home | Go to start of file |
| Ctrl+End | Go to end of file |
| Ctrl+Left/Right | Move by word |
| Home/End | Start/end of line |
| Page Up/Down | Scroll by page |

### Search

| Shortcut | Action |
|----------|--------|
| Ctrl+F | Find |
| Ctrl+H | Find and replace |
| F3 | Find next |
| Shift+F3 | Find previous |

### View

| Shortcut | Action |
|----------|--------|
| Ctrl+B | Toggle sidebar |
| Ctrl+P | Command palette |

See [KEYBINDINGS.md](docs/KEYBINDINGS.md) for full reference.

## Mouse Support

- **Left click**: Position cursor
- **Drag**: Select text
- **Double click**: Select word
- **Triple click**: Select line
- **Scroll wheel**: Scroll content
- **Sidebar click**: Open file

## Architecture

vex is built with a modular architecture:

- **Buffer**: Gap buffer for efficient text editing
- **Editor**: Coordinates buffer, cursor, selection, and history
- **UI Components**: Title bar, status bar, sidebar, command palette, search bar
- **Syntax Highlighting**: Chroma integration for 200+ languages

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for technical details.

## Dependencies

All dependencies are MIT-licensed:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Chroma](https://github.com/alecthomas/chroma) - Syntax highlighting
- [golang.design/x/clipboard](https://golang.design/x/clipboard) - Clipboard access
- [fuzzy](https://github.com/sahilm/fuzzy) - Fuzzy matching

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT License - see [LICENSE](LICENSE)
