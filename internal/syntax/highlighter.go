package syntax

import (
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/charmbracelet/lipgloss"
)

// StyledSegment represents a segment of text with a style.
type StyledSegment struct {
	Text  string
	Style lipgloss.Style
}

// StyledLine represents a line of styled segments.
type StyledLine struct {
	Segments []StyledSegment
}

// Highlighter provides syntax highlighting using Chroma.
type Highlighter struct {
	lexer    chroma.Lexer
	language string
	theme    *Theme
}

// Theme defines colors for syntax highlighting.
type Theme struct {
	Default     lipgloss.Style
	Keyword     lipgloss.Style
	Name        lipgloss.Style
	Function    lipgloss.Style
	String      lipgloss.Style
	Number      lipgloss.Style
	Comment     lipgloss.Style
	Operator    lipgloss.Style
	Punctuation lipgloss.Style
	Type        lipgloss.Style
	Error       lipgloss.Style
	Background  lipgloss.Color
}

// DefaultTheme returns the default syntax highlighting theme.
func DefaultTheme() *Theme {
	return &Theme{
		Default:     lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		Keyword:     lipgloss.NewStyle().Foreground(lipgloss.Color("204")),
		Name:        lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		Function:    lipgloss.NewStyle().Foreground(lipgloss.Color("117")),
		String:      lipgloss.NewStyle().Foreground(lipgloss.Color("149")),
		Number:      lipgloss.NewStyle().Foreground(lipgloss.Color("209")),
		Comment:     lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Italic(true),
		Operator:    lipgloss.NewStyle().Foreground(lipgloss.Color("204")),
		Punctuation: lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		Type:        lipgloss.NewStyle().Foreground(lipgloss.Color("81")),
		Error:       lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		Background:  lipgloss.Color("235"),
	}
}

// MonokaiTheme returns a Monokai-inspired theme.
func MonokaiTheme() *Theme {
	return &Theme{
		Default:     lipgloss.NewStyle().Foreground(lipgloss.Color("231")),
		Keyword:     lipgloss.NewStyle().Foreground(lipgloss.Color("197")),
		Name:        lipgloss.NewStyle().Foreground(lipgloss.Color("231")),
		Function:    lipgloss.NewStyle().Foreground(lipgloss.Color("148")),
		String:      lipgloss.NewStyle().Foreground(lipgloss.Color("186")),
		Number:      lipgloss.NewStyle().Foreground(lipgloss.Color("141")),
		Comment:     lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Italic(true),
		Operator:    lipgloss.NewStyle().Foreground(lipgloss.Color("197")),
		Punctuation: lipgloss.NewStyle().Foreground(lipgloss.Color("231")),
		Type:        lipgloss.NewStyle().Foreground(lipgloss.Color("81")),
		Error:       lipgloss.NewStyle().Foreground(lipgloss.Color("197")),
		Background:  lipgloss.Color("235"),
	}
}

// NewHighlighter creates a new syntax highlighter for the given file path.
func NewHighlighter(filepath string) *Highlighter {
	h := &Highlighter{
		theme: DefaultTheme(),
	}
	h.SetLanguageFromPath(filepath)
	return h
}

// NewHighlighterWithLanguage creates a highlighter for a specific language.
func NewHighlighterWithLanguage(language string) *Highlighter {
	h := &Highlighter{
		theme: DefaultTheme(),
	}
	h.SetLanguage(language)
	return h
}

// SetLanguageFromPath detects and sets the language based on file extension.
func (h *Highlighter) SetLanguageFromPath(path string) {
	if path == "" {
		h.lexer = lexers.Fallback
		h.language = "plain"
		return
	}

	// Try to get lexer by filename
	h.lexer = lexers.Match(path)
	if h.lexer == nil {
		h.lexer = lexers.Fallback
		h.language = "plain"
		return
	}

	h.lexer = chroma.Coalesce(h.lexer)
	config := h.lexer.Config()
	if config != nil {
		h.language = config.Name
	}
}

// SetLanguage sets the language for highlighting.
func (h *Highlighter) SetLanguage(language string) {
	if language == "" || language == "plain" {
		h.lexer = lexers.Fallback
		h.language = "plain"
		return
	}

	h.lexer = lexers.Get(language)
	if h.lexer == nil {
		h.lexer = lexers.Fallback
		h.language = "plain"
		return
	}

	h.lexer = chroma.Coalesce(h.lexer)
	h.language = language
}

// SetTheme sets the highlighting theme.
func (h *Highlighter) SetTheme(theme *Theme) {
	h.theme = theme
}

// Language returns the current language name.
func (h *Highlighter) Language() string {
	return h.language
}

// HighlightLine highlights a single line of code.
func (h *Highlighter) HighlightLine(line string) StyledLine {
	if h.lexer == nil {
		return StyledLine{
			Segments: []StyledSegment{{Text: line, Style: h.theme.Default}},
		}
	}

	iterator, err := h.lexer.Tokenise(nil, line)
	if err != nil {
		return StyledLine{
			Segments: []StyledSegment{{Text: line, Style: h.theme.Default}},
		}
	}

	var segments []StyledSegment
	for token := iterator(); token != chroma.EOF; token = iterator() {
		style := h.styleForToken(token.Type)
		segments = append(segments, StyledSegment{
			Text:  token.Value,
			Style: style,
		})
	}

	return StyledLine{Segments: segments}
}

// Highlight highlights entire content and returns styled lines.
func (h *Highlighter) Highlight(content string) []StyledLine {
	lines := strings.Split(content, "\n")
	result := make([]StyledLine, len(lines))

	// For performance, tokenize the entire content and distribute
	if h.lexer == nil {
		for i, line := range lines {
			result[i] = StyledLine{
				Segments: []StyledSegment{{Text: line, Style: h.theme.Default}},
			}
		}
		return result
	}

	iterator, err := h.lexer.Tokenise(nil, content)
	if err != nil {
		for i, line := range lines {
			result[i] = StyledLine{
				Segments: []StyledSegment{{Text: line, Style: h.theme.Default}},
			}
		}
		return result
	}

	currentLine := 0
	var currentSegments []StyledSegment

	for token := iterator(); token != chroma.EOF; token = iterator() {
		style := h.styleForToken(token.Type)
		tokenLines := strings.Split(token.Value, "\n")

		for i, part := range tokenLines {
			if i > 0 {
				// New line - save current and start new
				result[currentLine] = StyledLine{Segments: currentSegments}
				currentLine++
				currentSegments = nil
			}
			if part != "" {
				currentSegments = append(currentSegments, StyledSegment{
					Text:  part,
					Style: style,
				})
			}
		}
	}

	// Save last line
	if currentLine < len(result) {
		result[currentLine] = StyledLine{Segments: currentSegments}
	}

	return result
}

// HighlightLines highlights a range of lines (0-indexed).
func (h *Highlighter) HighlightLines(content string, startLine, endLine int) []StyledLine {
	lines := strings.Split(content, "\n")

	if startLine < 0 {
		startLine = 0
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	if startLine >= endLine {
		return nil
	}

	// Extract the subset of lines
	subset := strings.Join(lines[startLine:endLine], "\n")
	return h.Highlight(subset)
}

// styleForToken returns the appropriate style for a Chroma token type.
func (h *Highlighter) styleForToken(tokenType chroma.TokenType) lipgloss.Style {
	switch {
	case tokenType == chroma.Keyword ||
		tokenType == chroma.KeywordConstant ||
		tokenType == chroma.KeywordDeclaration ||
		tokenType == chroma.KeywordNamespace ||
		tokenType == chroma.KeywordPseudo ||
		tokenType == chroma.KeywordReserved ||
		tokenType == chroma.KeywordType:
		return h.theme.Keyword

	case tokenType == chroma.NameFunction ||
		tokenType == chroma.NameFunctionMagic:
		return h.theme.Function

	case tokenType == chroma.NameBuiltin ||
		tokenType == chroma.NameBuiltinPseudo:
		return h.theme.Function

	case tokenType == chroma.NameClass ||
		tokenType == chroma.NameException:
		return h.theme.Type

	case tokenType == chroma.String ||
		tokenType == chroma.StringAffix ||
		tokenType == chroma.StringBacktick ||
		tokenType == chroma.StringChar ||
		tokenType == chroma.StringDelimiter ||
		tokenType == chroma.StringDoc ||
		tokenType == chroma.StringDouble ||
		tokenType == chroma.StringEscape ||
		tokenType == chroma.StringHeredoc ||
		tokenType == chroma.StringInterpol ||
		tokenType == chroma.StringOther ||
		tokenType == chroma.StringRegex ||
		tokenType == chroma.StringSingle ||
		tokenType == chroma.StringSymbol:
		return h.theme.String

	case tokenType == chroma.Number ||
		tokenType == chroma.NumberBin ||
		tokenType == chroma.NumberFloat ||
		tokenType == chroma.NumberHex ||
		tokenType == chroma.NumberInteger ||
		tokenType == chroma.NumberIntegerLong ||
		tokenType == chroma.NumberOct:
		return h.theme.Number

	case tokenType == chroma.Comment ||
		tokenType == chroma.CommentHashbang ||
		tokenType == chroma.CommentMultiline ||
		tokenType == chroma.CommentPreproc ||
		tokenType == chroma.CommentPreprocFile ||
		tokenType == chroma.CommentSingle ||
		tokenType == chroma.CommentSpecial:
		return h.theme.Comment

	case tokenType == chroma.Operator ||
		tokenType == chroma.OperatorWord:
		return h.theme.Operator

	case tokenType == chroma.Punctuation:
		return h.theme.Punctuation

	case tokenType == chroma.Error:
		return h.theme.Error

	default:
		return h.theme.Default
	}
}

// Render renders a styled line to a string with ANSI codes.
func (sl *StyledLine) Render() string {
	var sb strings.Builder
	for _, seg := range sl.Segments {
		sb.WriteString(seg.Style.Render(seg.Text))
	}
	return sb.String()
}

// DetectLanguage detects the programming language from file path or content.
func DetectLanguage(path string, content string) string {
	// Try by file extension first
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "Go"
	case ".py":
		return "Python"
	case ".js":
		return "JavaScript"
	case ".ts":
		return "TypeScript"
	case ".jsx":
		return "JSX"
	case ".tsx":
		return "TSX"
	case ".rs":
		return "Rust"
	case ".c":
		return "C"
	case ".cpp", ".cc", ".cxx":
		return "C++"
	case ".h", ".hpp":
		return "C++"
	case ".java":
		return "Java"
	case ".rb":
		return "Ruby"
	case ".php":
		return "PHP"
	case ".sh", ".bash":
		return "Bash"
	case ".zsh":
		return "Zsh"
	case ".json":
		return "JSON"
	case ".yaml", ".yml":
		return "YAML"
	case ".toml":
		return "TOML"
	case ".xml":
		return "XML"
	case ".html", ".htm":
		return "HTML"
	case ".css":
		return "CSS"
	case ".scss":
		return "SCSS"
	case ".md", ".markdown":
		return "Markdown"
	case ".sql":
		return "SQL"
	case ".swift":
		return "Swift"
	case ".kt", ".kts":
		return "Kotlin"
	case ".lua":
		return "Lua"
	case ".zig":
		return "Zig"
	case ".nim":
		return "Nim"
	case ".ex", ".exs":
		return "Elixir"
	case ".erl", ".hrl":
		return "Erlang"
	case ".hs":
		return "Haskell"
	case ".ml", ".mli":
		return "OCaml"
	case ".r", ".R":
		return "R"
	case ".jl":
		return "Julia"
	case ".pl", ".pm":
		return "Perl"
	case ".scala":
		return "Scala"
	case ".clj", ".cljs":
		return "Clojure"
	case ".vim":
		return "VimL"
	case ".el":
		return "EmacsLisp"
	case ".dockerfile":
		return "Docker"
	case ".tf":
		return "Terraform"
	case ".proto":
		return "Protocol Buffer"
	case ".graphql", ".gql":
		return "GraphQL"
	}

	// Check for specific filenames
	filename := strings.ToLower(filepath.Base(path))
	switch filename {
	case "makefile", "gnumakefile":
		return "Makefile"
	case "dockerfile":
		return "Docker"
	case "cmakelists.txt":
		return "CMake"
	case ".gitignore", ".dockerignore":
		return "Ignore List"
	case "go.mod":
		return "Go Module"
	case "go.sum":
		return "Go Checksum"
	case "cargo.toml":
		return "TOML"
	case "package.json", "tsconfig.json":
		return "JSON"
	}

	// Try to detect from shebang
	if len(content) > 2 && content[:2] == "#!" {
		firstLine := strings.Split(content, "\n")[0]
		if strings.Contains(firstLine, "python") {
			return "Python"
		}
		if strings.Contains(firstLine, "node") || strings.Contains(firstLine, "deno") {
			return "JavaScript"
		}
		if strings.Contains(firstLine, "ruby") {
			return "Ruby"
		}
		if strings.Contains(firstLine, "bash") || strings.Contains(firstLine, "sh") {
			return "Bash"
		}
		if strings.Contains(firstLine, "perl") {
			return "Perl"
		}
		if strings.Contains(firstLine, "php") {
			return "PHP"
		}
	}

	return "plain"
}
