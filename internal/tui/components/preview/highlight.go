package preview

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// defaultStyle is the syntax highlighting style to use.
// "monokai" works well on dark terminals; alternatives: "dracula", "native", "vim"
const defaultStyle = "monokai"

// defaultFormatter is the terminal formatter to use.
// "terminal256" provides good color support on most terminals.
const defaultFormatter = "terminal256"

// Highlight applies syntax highlighting to code based on the filename.
// Returns the original content if highlighting fails or no lexer is found.
func Highlight(filename string, content []byte) string {
	text := string(content)

	// Try to find a lexer for this file
	lexer := lexers.Match(filename)
	if lexer == nil {
		// Try to detect from content
		lexer = lexers.Analyse(text) //nolint:misspell // Chroma API uses British spelling
	}
	if lexer == nil {
		// No syntax highlighting available
		return text
	}

	// Coalesce runs of identical token types for cleaner output
	lexer = chroma.Coalesce(lexer)

	// Get style and formatter
	style := styles.Get(defaultStyle)
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get(defaultFormatter)
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Tokenize the content
	iterator, err := lexer.Tokenise(nil, text)
	if err != nil {
		return text
	}

	// Format to a buffer
	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return text
	}

	return buf.String()
}

// CanHighlight returns true if syntax highlighting is available for the given filename.
func CanHighlight(filename string) bool {
	return lexers.Match(filename) != nil
}

// GetLanguage returns the detected language name for a filename, or empty string if unknown.
func GetLanguage(filename string) string {
	lexer := lexers.Match(filename)
	if lexer == nil {
		return ""
	}
	config := lexer.Config()
	if config == nil {
		return ""
	}
	return config.Name
}

// commonHighlightExtensions is a quick lookup for files we know can be highlighted.
// This avoids calling lexers.Match for common cases.
var commonHighlightExtensions = map[string]bool{
	".go":    true,
	".py":    true,
	".js":    true,
	".ts":    true,
	".jsx":   true,
	".tsx":   true,
	".java":  true,
	".c":     true,
	".cpp":   true,
	".h":     true,
	".hpp":   true,
	".rs":    true,
	".rb":    true,
	".php":   true,
	".sh":    true,
	".bash":  true,
	".zsh":   true,
	".yaml":  true,
	".yml":   true,
	".json":  true,
	".xml":   true,
	".html":  true,
	".css":   true,
	".scss":  true,
	".md":    true,
	".sql":   true,
	".toml":  true,
	".ini":   true,
	".conf":  true,
	".swift": true,
	".kt":    true,
	".scala": true,
	".lua":   true,
	".vim":   true,
	".zig":   true,
}

// IsLikelyHighlightable does a quick check if a file extension is commonly highlighted.
func IsLikelyHighlightable(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return commonHighlightExtensions[ext]
}
