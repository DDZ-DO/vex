package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/DDZ-DO/vex/internal/app"
)

var version = "dev"

func main() {
	args := os.Args[1:]

	// Handle flags
	if len(args) > 0 {
		switch args[0] {
		case "-h", "--help":
			printHelp()
			return
		case "-v", "--version":
			printVersion()
			return
		}
	}

	// Parse file path and optional line number
	var filepath string
	var line int

	if len(args) > 0 {
		filepath, line = parseFileArg(args[0])
	}

	// Run the editor
	if err := app.RunWithVersion(filepath, line, version); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// parseFileArg parses a file argument that may include a line number.
// Supports formats: "file.go", "file.go:42"
func parseFileArg(arg string) (string, int) {
	// Check for :line suffix
	if idx := strings.LastIndex(arg, ":"); idx > 0 {
		filepath := arg[:idx]
		lineStr := arg[idx+1:]

		// Check if what follows is a number
		if line, err := strconv.Atoi(lineStr); err == nil && line > 0 {
			return filepath, line
		}
	}

	return arg, 0
}

func printHelp() {
	help := `vex - A modern terminal text editor

USAGE:
    vex [FILE][:LINE]

ARGUMENTS:
    FILE        File to open (optional)
    :LINE       Line number to jump to (optional)

OPTIONS:
    -h, --help      Show this help message
    -v, --version   Show version information

EXAMPLES:
    vex                 Open editor with empty buffer
    vex file.go         Open file.go
    vex file.go:42      Open file.go at line 42

KEYBINDINGS:
    File Operations:
        Ctrl+S          Save file
        Ctrl+N          New file
        Ctrl+O          Open file
        Ctrl+W          Close file
        Ctrl+Q          Quit

    Editing:
        Ctrl+Z          Undo
        Ctrl+Y          Redo
        Ctrl+X          Cut
        Ctrl+C          Copy
        Ctrl+V          Paste
        Ctrl+A          Select all
        Ctrl+D          Duplicate line
        Ctrl+L          Delete line
        Alt+Up/Down     Move line up/down

    Navigation:
        Ctrl+G          Go to line
        Ctrl+Home       Go to start of file
        Ctrl+End        Go to end of file
        Ctrl+Left/Right Move by word

    Search:
        Ctrl+F          Find
        Ctrl+H          Find and replace
        F3              Find next
        Shift+F3        Find previous

    View:
        Ctrl+B          Toggle sidebar
        Ctrl+P          Command palette

For more information, visit: https://github.com/DDZ-DO/vex
`
	fmt.Print(help)
}

func printVersion() {
	fmt.Printf("vex version %s\n", version)
}
