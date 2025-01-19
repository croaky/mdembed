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
		".bash": {LineComment: "#"},
		".sh":   {LineComment: "#"},
		".lua":  {LineComment: "--"},
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
				if err := processEmbedCodeBlock(embedLines, output); err !=
					nil {
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
		subsetName := ""

		if len(parts) == 2 {
			subsetName = parts[1]
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
		fileNameComment := buildFileNameComment(filename, commentStyle)

		if subsetName != "" {
			beginMarker, endMarker := buildMarkers(commentStyle, subsetName)

			beginIndex := strings.Index(fileContent, beginMarker)
			if beginIndex == -1 {
				return fmt.Errorf("begin marker '%s' not found in file %s",
					beginMarker, filename)
			}
			beginIndex += len(beginMarker)

			endIndex := strings.Index(fileContent[beginIndex:], endMarker)
			if endIndex == -1 {
				return fmt.Errorf("end marker '%s' not found in file %s",
					endMarker, filename)
			}
			endIndex += beginIndex

			fileContent = fileContent[beginIndex:endIndex]
		}

		fileContent = strings.TrimSpace(fileContent)

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

func buildMarkers(style CommentStyle, subsetName string) (string, string) {
	subsetName = strings.TrimSpace(subsetName) // Ensure no extra spaces

	if style.LineComment != "" {
		beginMarker := strings.TrimSpace(fmt.Sprintf("%s emdo %s", style.
			LineComment, subsetName))
		endMarker := strings.TrimSpace(fmt.Sprintf("%s emdone %s", style.
			LineComment, subsetName))
		return beginMarker, endMarker
	} else if style.BlockStart != "" && style.BlockEnd != "" {
		beginContent := strings.TrimSpace(fmt.Sprintf("emdo %s", subsetName))
		endContent := strings.TrimSpace(fmt.Sprintf("emdone %s", subsetName))

		beginMarker := fmt.Sprintf("%s %s %s", style.BlockStart, beginContent,
			style.BlockEnd)
		endMarker := fmt.Sprintf("%s %s %s", style.BlockStart, endContent,
			style.BlockEnd)
		return beginMarker, endMarker
	}

	// Default markers
	beginMarker := fmt.Sprintf("/* emdo %s */", subsetName)
	endMarker := fmt.Sprintf("/* emdone %s */", subsetName)
	return beginMarker, endMarker
}

func buildFileNameComment(filename string, style CommentStyle) string {
	if style.LineComment != "" {
		return style.LineComment + " " + filename
	}
	if style.BlockStart != "" && style.BlockEnd != "" {
		return fmt.Sprintf("%s %s %s", style.BlockStart, filename, style.
			BlockEnd)
	}
	// Default to line comment "#"
	return "# " + filename
}
