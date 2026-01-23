# Contributing to vex

Thanks for your interest in contributing to vex!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/[your-user]/vex.git`
3. Create a branch: `git checkout -b feature/my-feature`
4. Make your changes
5. Run tests: `make test`
6. Commit: `git commit -m "Add my feature"`
7. Push: `git push origin feature/my-feature`
8. Open a Pull Request

## Development Setup

### Prerequisites

- Go 1.22 or later
- Make (optional, for convenience)

### Building

```bash
# Build the binary
make build

# Or directly with go
go build -o bin/vex ./cmd/vex
```

### Running Tests

```bash
make test
```

### Running the Linter

```bash
make lint
```

## Code Style

- Run `gofmt` before committing
- Follow standard Go conventions
- Add tests for new features
- Keep functions small and focused
- Write clear, descriptive commit messages

## Project Structure

```
vex/
├── cmd/vex/           # CLI entry point
├── internal/
│   ├── app/           # Main application orchestration
│   ├── editor/        # Core editor components
│   │   ├── buffer.go  # Gap buffer implementation
│   │   ├── cursor.go  # Cursor management
│   │   ├── selection.go # Text selection
│   │   ├── history.go # Undo/redo
│   │   └── editor.go  # Editor component
│   ├── syntax/        # Syntax highlighting
│   ├── ui/            # UI components
│   ├── keybindings/   # Keyboard shortcuts
│   └── config/        # Configuration
└── docs/              # Documentation
```

## Reporting Issues

- Check existing issues first
- Include Go version, OS, and terminal
- Provide steps to reproduce
- Include error messages if applicable

## Feature Requests

- Open an issue describing the feature
- Explain the use case
- Consider if it aligns with vex's philosophy (simple, intuitive, VSCode-like)

## Pull Request Guidelines

1. Keep PRs focused on a single change
2. Update documentation if needed
3. Add tests for new functionality
4. Ensure all tests pass
5. Follow the existing code style

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
