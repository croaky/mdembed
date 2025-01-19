package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	commentStyles = map[string]CommentStyle{
		".css":  {BlockStart: "/*", BlockEnd: "*/"},
		".go":   {LineComment: "//"},
		".haml": {LineComment: "-#"},
		".html": {BlockStart: "<!--", BlockEnd: "-->"},
		".js":   {LineComment: "//", BlockStart: "/*", BlockEnd: "*/"},
		".rb":   {LineComment: "#"},
		".scss": {BlockStart: "/*", BlockEnd: "*/"},
		".ts":   {LineComment: "//", BlockStart: "/*", BlockEnd: "*/"},
	}
)

type CommentStyle struct {
	LineComment string
	BlockStart  string
	BlockEnd    string
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
				return fmt.Errorf("emdo marker not found in file %s", filename)
			}
			beginIndex += len(beginMarker)
			endIndex := strings.Index(fileContent[beginIndex:], endMarker)
			if endIndex == -1 {
				return fmt.Errorf("emdone marker not found in file %s", filename)
			}
			endIndex += beginIndex
			fileContent = fileContent[beginIndex:endIndex]
		}

		fileContent = strings.TrimSpace(fileContent)

		fileNameComment := buildFileNameComment(filename, commentStyle)

		fmt.Fprintf(output, "```%s\n", lang)
		fmt.Fprintln(output, fileNameComment)
		fmt.Fprintln(output, fileContent)
		fmt.Fprintln(output, "```")
	}

	return nil
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
		beginMarker = style.LineComment + " emdo"
		endMarker = style.LineComment + " emdone"
		return
	}
	// Use BlockStart and BlockEnd
	if style.BlockStart != "" && style.BlockEnd != "" {
		beginMarker = style.BlockStart + " emdo " + style.BlockEnd
		endMarker = style.BlockStart + " emdone " + style.BlockEnd
		return
	}
	// Fallback to default markers
	beginMarker = "/* emdo */"
	endMarker = "/* emdone */"
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
