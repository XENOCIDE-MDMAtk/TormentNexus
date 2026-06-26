package ctxharvester

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Language represents a supported programming language for syntax-aware chunking.
type Language string

const (
	LangGo         Language = "go"
	LangPython     Language = "python"
	LangTypeScript Language = "typescript"
	LangJavaScript Language = "javascript"
	LangGeneric    Language = "generic"
)

// detectLanguage guesses the language from the file name.
func detectLanguage(filename string) Language {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		return LangGo
	case ".py":
		return LangPython
	case ".ts", ".tsx":
		return LangTypeScript
	case ".js", ".jsx":
		return LangJavaScript
	default:
		return LangGeneric
	}
}

// castChunk chunks source code files based on their syntactic structure,
// prepending contextual headers to each chunk.
func castChunk(content string, filename string, targetSize int) []string {
	lang := detectLanguage(filename)
	if lang == LangGeneric {
		// Fall back to standard word-based chunking
		return semanticChunk(content, targetSize, targetSize/10)
	}

	lines := strings.Split(content, "\n")
	var chunks []string

	switch lang {
	case LangGo:
		chunks = chunkGo(lines, filename, targetSize)
	case LangPython:
		chunks = chunkPython(lines, filename, targetSize)
	case LangTypeScript, LangJavaScript:
		chunks = chunkJSFamily(lines, filename, targetSize)
	default:
		chunks = semanticChunk(content, targetSize, targetSize/10)
	}

	return chunks
}

// helper to calculate approximate word count
func countWords(text string) int {
	return len(strings.Fields(text))
}

// chunkGo parses Go files line-by-line using brace tracking.
func chunkGo(lines []string, filename string, targetSize int) []string {
	var chunks []string
	var headerLines []string
	var currentBlock []string
	
	inImport := false
	braceCount := 0
	wordCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Extract package and imports for the global context header
		if strings.HasPrefix(trimmed, "package ") {
			headerLines = append(headerLines, line)
			continue
		}
		if strings.HasPrefix(trimmed, "import (") {
			inImport = true
			headerLines = append(headerLines, line)
			continue
		}
		if inImport {
			headerLines = append(headerLines, line)
			if trimmed == ")" {
				inImport = false
			}
			continue
		}
		if strings.HasPrefix(trimmed, "import ") {
			headerLines = append(headerLines, line)
			continue
		}

		// Track braces to identify top-level code blocks (functions, structs, etc.)
		openCount := strings.Count(line, "{")
		closeCount := strings.Count(line, "}")
		braceCount += openCount - closeCount

		currentBlock = append(currentBlock, line)
		wordCount += countWords(line)

		// When a top-level block closes, or if the chunk size exceeds target, flush it
		if braceCount <= 0 && len(currentBlock) > 0 {
			if wordCount >= targetSize || len(chunks) == 0 {
				chunks = append(chunks, buildChunkWithHeader(headerLines, currentBlock, filename, "go"))
				currentBlock = nil
				wordCount = 0
			}
		}
	}

	// Flush remaining
	if len(currentBlock) > 0 {
		chunks = append(chunks, buildChunkWithHeader(headerLines, currentBlock, filename, "go"))
	}

	return chunks
}

// chunkPython parses Python files based on indentation changes.
func chunkPython(lines []string, filename string, targetSize int) []string {
	var chunks []string
	var headerLines []string
	var currentBlock []string
	wordCount := 0

	defRe := regexp.MustCompile(`^(class\s+|def\s+|@)`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Extract imports and module-level docs for context header
		if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "from ") {
			headerLines = append(headerLines, line)
			continue
		}

		// Detect a new class or function definition
		isNewBlock := defRe.MatchString(line)

		if isNewBlock && len(currentBlock) > 0 && wordCount >= targetSize {
			chunks = append(chunks, buildChunkWithHeader(headerLines, currentBlock, filename, "python"))
			currentBlock = nil
			wordCount = 0
		}

		currentBlock = append(currentBlock, line)
		wordCount += countWords(line)
	}

	if len(currentBlock) > 0 {
		chunks = append(chunks, buildChunkWithHeader(headerLines, currentBlock, filename, "python"))
	}

	return chunks
}

// chunkJSFamily parses JavaScript and TypeScript files.
func chunkJSFamily(lines []string, filename string, targetSize int) []string {
	var chunks []string
	var headerLines []string
	var currentBlock []string
	braceCount := 0
	wordCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Capture imports for context header
		if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "const ") && strings.Contains(trimmed, "require(") {
			headerLines = append(headerLines, line)
			continue
		}

		openCount := strings.Count(line, "{")
		closeCount := strings.Count(line, "}")
		braceCount += openCount - closeCount

		currentBlock = append(currentBlock, line)
		wordCount += countWords(line)

		if braceCount <= 0 && len(currentBlock) > 0 {
			if wordCount >= targetSize || len(chunks) == 0 {
				chunks = append(chunks, buildChunkWithHeader(headerLines, currentBlock, filename, "js/ts"))
				currentBlock = nil
				wordCount = 0
			}
		}
	}

	if len(currentBlock) > 0 {
		chunks = append(chunks, buildChunkWithHeader(headerLines, currentBlock, filename, "js/ts"))
	}

	return chunks
}

// buildChunkWithHeader prepends comment-based Context headers to the block
func buildChunkWithHeader(headers []string, block []string, filename string, lang string) string {
	var sb strings.Builder
	commentChar := "//"
	if lang == "python" {
		commentChar = "#"
	}

	// Limit headers to avoid overloading context
	maxHeaders := 8
	var activeHeaders []string
	for _, h := range headers {
		hTrim := strings.TrimSpace(h)
		if hTrim != "" && len(activeHeaders) < maxHeaders {
			activeHeaders = append(activeHeaders, hTrim)
		}
	}

	sb.WriteString(fmt.Sprintf("%s [cAST Context: %s | File: %s]\n", commentChar, lang, filepath.Base(filename)))
	if len(activeHeaders) > 0 {
		sb.WriteString(fmt.Sprintf("%s Global Headers:\n", commentChar))
		for _, h := range activeHeaders {
			sb.WriteString(fmt.Sprintf("%s   %s\n", commentChar, h))
		}
	}
	sb.WriteString("\n")
	sb.WriteString(strings.Join(block, "\n"))

	return sb.String()
}
