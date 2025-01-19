package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var styles = map[string]Style{
	".bash": {LineComment: "#"},
	".css":  {BlockDo: "/*", BlockDone: "*/"},
	".go":   {LineComment: "//"},
	".haml": {LineComment: "-#"},
	".html": {BlockDo: "<!--", BlockDone: "-->"},
	".js":   {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},
	".lua":  {LineComment: "--"},
	".rb":   {LineComment: "#"},
	".scss": {BlockDo: "/*", BlockDone: "*/"},
	".sh":   {LineComment: "#"},
	".ts":   {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},
}

type Style struct {
	LineComment string
	BlockDo     string
	BlockDone   string
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
	if len(lines) == 0 {
		return nil
	}

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
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

		style, ok := styles[ext]
		if !ok {
			return fmt.Errorf("unsupported file type: %s", ext)
		}

		var fileName string
		if style.LineComment != "" {
			fileName = style.LineComment + " " + filename
		} else if style.BlockDo != "" && style.BlockDone != "" {
			fileName = fmt.Sprintf("%s %s %s", style.BlockDo, filename, style.BlockDone)
		}

		if blockName != "" {
			blockName = strings.TrimSpace(blockName)
			var doMark, doneMark string

			if style.LineComment != "" {
				doMark = strings.TrimSpace(fmt.Sprintf("%s emdo %s", style.LineComment, blockName))
				doneMark = strings.TrimSpace(fmt.Sprintf("%s emdone %s", style.LineComment, blockName))
			} else if style.BlockDo != "" && style.BlockDone != "" {
				beginContent := strings.TrimSpace(fmt.Sprintf("emdo %s", blockName))
				endContent := strings.TrimSpace(fmt.Sprintf("emdone %s", blockName))

				doMark = fmt.Sprintf("%s %s %s", style.BlockDo, beginContent, style.BlockDone)
				doneMark = fmt.Sprintf("%s %s %s", style.BlockDo, endContent, style.BlockDone)
			}

			beginIndex := strings.Index(fileContent, doMark)
			if beginIndex == -1 {
				return fmt.Errorf("do mark '%s' not found in file %s", doMark, filename)
			}
			beginIndex += len(doMark)

			endIndex := strings.Index(fileContent[beginIndex:], doneMark)
			if endIndex == -1 {
				return fmt.Errorf("done mark '%s' not found in file %s", doneMark, filename)
			}
			endIndex += beginIndex

			fileContent = fileContent[beginIndex:endIndex]
		}

		fileContent = strings.Trim(fileContent, "\n")
		fileContent = dedent(fileContent)

		fmt.Fprintf(output, "```%s\n", lang)
		fmt.Fprintln(output, fileName)

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
