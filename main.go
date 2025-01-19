package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var commentStyles = map[string]CommentStyle{
	".bash": {LineComment: "#"},
	".css":  {BlockStart: "/*", BlockEnd: "*/"},
	".go":   {LineComment: "//"},
	".haml": {LineComment: "-#"},
	".html": {BlockStart: "<!--", BlockEnd: "-->"},
	".js":   {LineComment: "//", BlockStart: "/*", BlockEnd: "*/"},
	".lua":  {LineComment: "--"},
	".rb":   {LineComment: "#"},
	".scss": {BlockStart: "/*", BlockEnd: "*/"},
	".sh":   {LineComment: "#"},
	".ts":   {LineComment: "//", BlockStart: "/*", BlockEnd: "*/"},
}

type CommentStyle struct {
	LineComment string
	BlockStart  string
	BlockEnd    string
}

func main() {
	if err := processMD(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func processMD(input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
	inEmbedBlock := false
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()

		if !inEmbedBlock {
			if line == "```embed" {
				inEmbedBlock = true
				lines = []string{}
			} else {
				fmt.Fprintln(output, line)
			}
		} else {
			if line == "```" {
				if err := processEmbed(lines, output); err != nil {
					return err
				}
				inEmbedBlock = false
			} else {
				lines = append(lines, line)
			}
		}
	}

	if inEmbedBlock {
		return fmt.Errorf("unterminated ```embed code block")
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func processEmbed(lines []string, output io.Writer) error {
	// Skip empty embed blocks
	if len(lines) == 0 {
		return nil
	}

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue // Skip empty lines
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		filename := parts[0]
		blockName := ""

		if len(parts) == 2 {
			blockName = parts[1]
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

		style, ok := commentStyles[ext]
		if !ok {
			return fmt.Errorf("unsupported file type: %s", ext)
		}

		var fileNameComment string
		if style.LineComment != "" {
			fileNameComment = style.LineComment + " " + filename
		} else if style.BlockStart != "" && style.BlockEnd != "" {
			fileNameComment = fmt.Sprintf("%s %s %s", style.BlockStart, filename, style.BlockEnd)
		} else {
			// Should not reach here since unsupported styles should have been caught
			return fmt.Errorf("unsupported comment style for file type: %s", ext)
		}

		if blockName != "" {
			blockName = strings.TrimSpace(blockName)
			var beginMarker, endMarker string

			if style.LineComment != "" {
				beginMarker = strings.TrimSpace(fmt.Sprintf("%s emdo %s", style.LineComment, blockName))
				endMarker = strings.TrimSpace(fmt.Sprintf("%s emdone %s", style.LineComment, blockName))
			} else if style.BlockStart != "" && style.BlockEnd != "" {
				beginContent := strings.TrimSpace(fmt.Sprintf("emdo %s", blockName))
				endContent := strings.TrimSpace(fmt.Sprintf("emdone %s", blockName))

				beginMarker = fmt.Sprintf("%s %s %s", style.BlockStart, beginContent, style.BlockEnd)
				endMarker = fmt.Sprintf("%s %s %s", style.BlockStart, endContent, style.BlockEnd)
			} else {
				// Should not reach here since unsupported styles should have been caught
				return fmt.Errorf("unsupported comment style for file type: %s", ext)
			}

			beginIndex := strings.Index(fileContent, beginMarker)
			if beginIndex == -1 {
				return fmt.Errorf("begin marker '%s' not found in file %s", beginMarker, filename)
			}
			beginIndex += len(beginMarker)

			endIndex := strings.Index(fileContent[beginIndex:], endMarker)
			if endIndex == -1 {
				return fmt.Errorf("end marker '%s' not found in file %s", endMarker, filename)
			}
			endIndex += beginIndex

			fileContent = fileContent[beginIndex:endIndex]
		}

		// Trim leading and trailing blank lines
		fileContent = strings.Trim(fileContent, "\n")

		// Dedent the file content
		fileContent = dedent(fileContent)

		fmt.Fprintf(output, "```%s\n", lang)
		fmt.Fprintln(output, fileNameComment)

		// Trim any trailing newlines to prevent extra blank lines
		fileContent = strings.TrimRight(fileContent, "\n")

		// Print the file content without adding an extra newline
		fmt.Fprint(output, fileContent)
		fmt.Fprintln(output) // Add a single newline after the content

		fmt.Fprintln(output, "```")

		// Add a newline between code blocks except after the last one
		if i < len(lines)-1 {
			fmt.Fprintln(output)
		}
	}

	return nil
}

func dedent(s string) string {
	lines := strings.Split(s, "\n")
	minIndent := -1

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			continue
		}
		indent := len(line) - len(trimmed)
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent > 0 {
		for i, line := range lines {
			if len(line) >= minIndent {
				lines[i] = line[minIndent:]
			}
		}
	}

	return strings.Join(lines, "\n")
}
