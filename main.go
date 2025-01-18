package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	leadingWhitespaceRe = regexp.MustCompile(`(?m)(^[ \t]*)(?:[^ \t])`)

	// Map of file extensions to their comment styles
	commentStyles = map[string]CommentStyle{
		".rb":   {LineComment: "#"},
		".haml": {LineComment: "-#"},
		".js":   {LineComment: "//", BlockStart: "/*", BlockEnd: "*/"},
		".ts":   {LineComment: "//", BlockStart: "/*", BlockEnd: "*/"},
		".go":   {LineComment: "//"},
		".css":  {BlockStart: "/*", BlockEnd: "*/"},
		".scss": {BlockStart: "/*", BlockEnd: "*/"},
		".html": {BlockStart: "<!--", BlockEnd: "-->"},
	}
)

type CommentStyle struct {
	LineComment string // e.g., "//" or "#"
	BlockStart  string // e.g., "/*" or "<!--"
	BlockEnd    string // e.g., "*/" or "-->"
}

func main() {
	if err := processMarkdown(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func processMarkdown(input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
	state := "NORMAL"
	var embedLines []string

	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case "NORMAL":
			if line == "```embed" {
				state = "EMBED_CODE_BLOCK"
				embedLines = []string{}
			} else {
				fmt.Fprintln(output, line)
			}
		case "EMBED_CODE_BLOCK":
			if line == "```" {
				if err := processEmbedCodeBlock(embedLines, output); err != nil {
					return err
				}
				state = "NORMAL"
			} else {
				embedLines = append(embedLines, line)
			}
		}
	}

	if state == "EMBED_CODE_BLOCK" {
		return fmt.Errorf("unterminated ```embed code block")
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func processEmbedCodeBlock(embedLines []string, output io.Writer) error {
	// Skip empty embed blocks
	if len(embedLines) == 0 {
		return nil
	}

	for _, line := range embedLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue // Skip empty lines
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		filename := parts[0]
		useSubset := false

		if len(parts) == 2 {
			if parts[1] == "subset" {
				useSubset = true
			} else {
				return fmt.Errorf("invalid option in embed code block: %s", line)
			}
		} else if len(parts) > 2 {
			return fmt.Errorf("invalid format in embed code block: %s", line)
		}

		// Read the file content
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", filename, err)
		}
		fileContent := string(content)

		ext := filepath.Ext(filename)
		lang := strings.TrimPrefix(ext, ".")

		commentStyle := getCommentStyle(ext)
		var beginMarker, endMarker string

		if useSubset {
			beginMarker, endMarker = buildMarkers(commentStyle)

			beginIndex := strings.Index(fileContent, beginMarker)
			if beginIndex == -1 {
				return fmt.Errorf("beginembed marker not found in file %s", filename)
			}
			beginIndex += len(beginMarker)
			endIndex := strings.Index(fileContent[beginIndex:], endMarker)
			if endIndex == -1 {
				return fmt.Errorf("endembed marker not found in file %s", filename)
			}
			endIndex += beginIndex
			fileContent = fileContent[beginIndex:endIndex]

			// Strip all leading spaces
			fileContent = stripAllLeadingSpaces(fileContent)
		}

		// Trim leading and trailing whitespace
		fileContent = strings.TrimSpace(fileContent)

		// Include file name as a comment at the top
		fileNameComment := buildFileNameComment(filename, commentStyle)

		// Output the code block
		fmt.Fprintf(output, "```%s\n", lang)
		fmt.Fprintln(output, fileNameComment)
		fmt.Fprintln(output, fileContent)
		fmt.Fprintln(output, "```")
	}

	return nil
}

func stripAllLeadingSpaces(content string) string {
	rawLines := strings.Split(content, "\n")
	var lines []string

	for _, l := range rawLines {
		lines = append(lines, strings.TrimLeft(l, " \t"))
	}
	return strings.Join(lines, "\n")
}

func getCommentStyle(ext string) CommentStyle {
	if style, ok := commentStyles[ext]; ok {
		return style
	}
	// Default to line comment "#"
	return CommentStyle{LineComment: "#"}
}

func buildMarkers(style CommentStyle) (beginMarker, endMarker string) {
	// Use LineComment if available
	if style.LineComment != "" {
		beginMarker = style.LineComment + " beginembed"
		endMarker = style.LineComment + " endembed"
		return
	}
	// Use BlockStart and BlockEnd
	if style.BlockStart != "" && style.BlockEnd != "" {
		beginMarker = style.BlockStart + " beginembed " + style.BlockEnd
		endMarker = style.BlockStart + " endembed " + style.BlockEnd
		return
	}
	// Fallback to default markers
	beginMarker = "/* beginembed */"
	endMarker = "/* endembed */"
	return
}

func buildFileNameComment(filename string, style CommentStyle) string {
	if style.LineComment != "" {
		return style.LineComment + " " + filename
	}
	if style.BlockStart != "" && style.BlockEnd != "" {
		return style.BlockStart + " " + filename + " " + style.BlockEnd
	}
	// Default to line comment "#"
	return "# " + filename
}
