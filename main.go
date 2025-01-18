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
)

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

	// Determine the language based on the extension of the first filename
	firstLine := strings.TrimSpace(embedLines[0])
	if firstLine == "" {
		return fmt.Errorf("embed code block is empty")
	}
	firstParts := strings.Fields(firstLine)
	if len(firstParts) == 0 {
		return fmt.Errorf("embed code block is empty")
	}
	firstFilename := firstParts[0]
	ext := strings.TrimPrefix(filepath.Ext(firstFilename), ".")
	lang := ext

	var codeLines []string

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

		if useSubset {
			beginMarker := "# beginembed"
			endMarker := "# endembed"
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
			fileContent = strings.TrimSpace(fileContent[beginIndex:endIndex])
		}

		// Trim trailing newlines
		fileContent = strings.TrimRight(fileContent, "\n")

		// Dedent content
		dedentedContent := dedentContent(fileContent)

		// Collect the dedented content
		codeLines = append(codeLines, dedentedContent)
	}

	// Output the code block with appropriate language tag
	fmt.Fprintf(output, "```%s\n", lang)
	for _, code := range codeLines {
		fmt.Fprintln(output, code)
	}
	fmt.Fprintln(output, "```")

	return nil
}

func dedentContent(content string) string {
	rawLines := strings.Split(content, "\n")
	var margin string
	var lines []string

	for i, l := range rawLines {
		if i == 0 {
			matches := leadingWhitespaceRe.FindStringSubmatch(l)
			if len(matches) > 1 {
				margin = matches[1]
			} else {
				margin = ""
			}
		}
		if margin != "" {
			dedented := strings.TrimPrefix(l, margin)
			lines = append(lines, dedented)
		} else {
			lines = append(lines, l)
		}
	}
	return strings.Join(lines, "\n")
}
