package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

var styles = map[string]Style{
	".ada":   {LineComment: "--"},                                   // Ada
	".asm":   {LineComment: ";"},                                    // Assembly
	".awk":   {LineComment: "#"},                                    // Awk
	".bash":  {LineComment: "#"},                                    // Bash
	".c":     {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // C
	".clj":   {LineComment: ";"},                                    // Clojure
	".cob":   {LineComment: "*>"},                                   // COBOL
	".cpp":   {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // C++
	".cs":    {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // C#
	".css":   {BlockDo: "/*", BlockDone: "*/"},                      // CSS
	".d":     {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // D
	".dart":  {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // Dart
	".elm":   {LineComment: "--", BlockDo: "{-", BlockDone: "-}"},   // Elm
	".erl":   {LineComment: "%"},                                    // Erlang
	".ex":    {LineComment: "#"},                                    // Elixir
	".f90":   {LineComment: "!"},                                    // Fortran
	".fs":    {LineComment: "//", BlockDo: "(*", BlockDone: "*)"},   // F#
	".gleam": {LineComment: "//"},                                   // Gleam
	".go":    {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // Go
	".haml":  {LineComment: "-#"},                                   // Haml
	".hs":    {LineComment: "--", BlockDo: "{-", BlockDone: "-}"},   // Haskell
	".html":  {BlockDo: "<!--", BlockDone: "-->"},                   // HTML
	".java":  {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // Java
	".jl":    {LineComment: "#", BlockDo: "#=", BlockDone: "=#"},    // Julia
	".js":    {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // JavaScript
	".jsx":   {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // JSX
	".kt":    {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // Kotlin
	".lisp":  {LineComment: ";", BlockDo: "#|", BlockDone: "|#"},    // Lisp
	".logo":  {LineComment: ";"},                                    // Logo
	".lua":   {LineComment: "--", BlockDo: "--[[", BlockDone: "]]"}, // Lua
	".m":     {LineComment: "%", BlockDo: "%{", BlockDone: "%}"},    // MATLAB
	".ml":    {BlockDo: "(*", BlockDone: "*)"},                      // OCaml
	".mm":    {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // Objective-C
	".mojo":  {LineComment: "#"},                                    // Mojo
	".nim":   {LineComment: "#", BlockDo: "#[", BlockDone: "]#"},    // Nim
	".pas":   {LineComment: "//", BlockDo: "{", BlockDone: "}"},     // Pascal
	".php":   {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // PHP
	".pl":    {LineComment: "#"},                                    // Perl
	".pro":   {LineComment: "%", BlockDo: "/*", BlockDone: "*/"},    // Prolog
	".py":    {LineComment: "#"},                                    // Python
	".r":     {LineComment: "#"},                                    // R
	".rb":    {LineComment: "#"},                                    // Ruby
	".rs":    {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // Rust
	".scala": {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // Scala
	".scm":   {LineComment: ";", BlockDo: "#|", BlockDone: "|#"},    // Scheme
	".scss":  {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // Sass
	".sh":    {LineComment: "#"},                                    // Shell
	".sol":   {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // Solidity
	".sql":   {LineComment: "--", BlockDo: "/*", BlockDone: "*/"},   // SQL
	".swift": {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // Swift
	".tcl":   {LineComment: "#"},                                    // Tcl
	".ts":    {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // TypeScript
	".tsx":   {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // TSX
	".vb":    {LineComment: "'"},                                    // VBScript
	".vbs":   {LineComment: "'"},                                    // Visual Basic
	".wl":    {BlockDo: "(*", BlockDone: "*)"},                      // Wolfram
	".yml":   {LineComment: "#"},                                    // YAML
	".zig":   {LineComment: "//", BlockDo: "/*", BlockDone: "*/"},   // Zig
}

type ProcessState struct {
	filesInProcess map[string]bool
}

type Style struct {
	LineComment string
	BlockDo     string
	BlockDone   string
}

func main() {
	state := &ProcessState{
		filesInProcess: make(map[string]bool),
	}
	if err := processMD(os.Stdin, os.Stdout, state); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// processMD reads Markdown from input and writes to output, processing embed blocks
func processMD(input io.Reader, output io.Writer, state *ProcessState) error {
	scanner := bufio.NewScanner(input)
	inEmbedBlock := false // Flag to track if we're inside an embed block
	var lines []string    // Collects lines within an embed block

	for scanner.Scan() {
		line := scanner.Text()

		if !inEmbedBlock {
			if line == "```embed" {
				// Start of embed block
				inEmbedBlock = true
				lines = []string{}
			} else {
				// Write line directly to output
				fmt.Fprintln(output, line)
			}
		} else {
			if line == "```" {
				// End of an embed block
				if err := processEmbed(lines, output, state); err != nil {
					return err
				}
				inEmbedBlock = false
			} else {
				// Collect lines in embed block
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

// processEmbed processes lines collected within an embed block
func processEmbed(lines []string, output io.Writer, state *ProcessState) error {
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		pattern := parts[0] // Can be a filename or a glob pattern
		blockName := ""     // Optional block name

		if len(parts) == 2 {
			blockName = parts[1]
		} else if len(parts) > 2 {
			return fmt.Errorf("invalid format in embed code block: %s", line)
		}

		// ensure pattern uses forward slashes
		pattern = filepath.ToSlash(pattern)
		// clean up the pattern to remove any ./ or ../
		pattern = path.Clean(pattern)

		// use doublestar.Glob with fs.FS
		fsys := os.DirFS(".")

		// support recursive glob patterns using the new doublestar v4 API
		matches, err := doublestar.Glob(fsys, pattern)
		if err != nil {
			return fmt.Errorf("failed to glob pattern %s: %v", pattern, err)
		}

		if len(matches) == 0 {
			return fmt.Errorf("no files match pattern %s", pattern)
		}

		for j, match := range matches {
			filename := match // relative to fs.FS root

			// Read the file content from fsys
			content, err := fs.ReadFile(fsys, filename)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %v", filename, err)
			}
			fileContent := string(content)

			// Adjust filename to include OS-specific path separators for display
			displayFilename := filepath.FromSlash(filename)

			if err := processFile(displayFilename, blockName, fileContent, output, state); err != nil {
				return err
			}

			// Add newline between multiple code blocks
			if j < len(matches)-1 {
				fmt.Fprintln(output)
			}
		}

		// Add newline between multiple patterns
		if i < len(lines)-1 {
			fmt.Fprintln(output)
		}
	}
	return nil
}

// processFile processes an individual file, handling circular embeddings
func processFile(filename, blockName, fileContent string, output io.Writer, state *ProcessState) error {
	if state.filesInProcess[filename] {
		return fmt.Errorf("circular embedding detected for file %s", filename)
	}

	// Mark the file as being processed
	state.filesInProcess[filename] = true
	defer delete(state.filesInProcess, filename)

	ext := filepath.Ext(filename)
	if ext == ".md" {
		// Process Markdown files recursively
		reader := strings.NewReader(fileContent)
		if err := processMD(reader, output, state); err != nil {
			return fmt.Errorf("processing markdown file %s failed: %v", filename, err)
		}
	} else {
		// Process other files as code blocks
		if err := processCodeFile(filename, blockName, fileContent, output); err != nil {
			return err
		}
	}
	return nil
}

// processCodeFile processes non-Markdown files and embeds their content in code fences
func processCodeFile(filename, blockName, fileContent string, output io.Writer) error {
	ext := filepath.Ext(filename)
	lang := strings.TrimPrefix(ext, ".")

	// Get comment style based on file extension
	style, ok := styles[ext]
	if !ok {
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	// Prepare filename comment
	var fileName string
	if style.LineComment != "" {
		fileName = style.LineComment + " " + filename
	} else if style.BlockDo != "" && style.BlockDone != "" {
		fileName = fmt.Sprintf("%s %s %s", style.BlockDo, filename, style.BlockDone)
	}

	// If a block name is specified, extract block between marks
	if blockName != "" {
		doMark, doneMark := getBlockMarkers(style, blockName)

		extractedContent, err := extractBlock(fileContent, doMark, doneMark)
		if err != nil {
			return fmt.Errorf("%v in file %s", err, filename)
		}

		fileContent = extractedContent
	}

	// Clean up content
	fileContent = strings.Trim(fileContent, "\n")
	fileContent = dedent(fileContent)

	// Write code block to output
	fmt.Fprintf(output, "```%s\n", lang)
	fmt.Fprintf(output, "%s\n", fileName)
	fmt.Fprintf(output, "%s", fileContent)
	fmt.Fprintf(output, "\n```\n")

	return nil
}

// getBlockMarkers generates the start and end markers for a block in a file
func getBlockMarkers(style Style, blockName string) (string, string) {
	blockName = strings.TrimSpace(blockName)
	var doMark, doneMark string

	if style.LineComment != "" {
		// Line comment marks
		doMark = strings.TrimSpace(fmt.Sprintf("%s emdo %s", style.LineComment, blockName))
		doneMark = strings.TrimSpace(fmt.Sprintf("%s emdone %s", style.LineComment, blockName))
	} else if style.BlockDo != "" && style.BlockDone != "" {
		// Block comment marks
		beginContent := strings.TrimSpace(fmt.Sprintf("emdo %s", blockName))
		endContent := strings.TrimSpace(fmt.Sprintf("emdone %s", blockName))

		doMark = fmt.Sprintf("%s %s %s", style.BlockDo, beginContent, style.BlockDone)
		doneMark = fmt.Sprintf("%s %s %s", style.BlockDo, endContent, style.BlockDone)
	} else {
		return "", ""
	}

	return doMark, doneMark
}

// extractBlock extracts the content between doMark and doneMark, ignoring leading and trailing whitespace
func extractBlock(fileContent, doMark, doneMark string) (string, error) {
	lines := strings.Split(fileContent, "\n")
	var inBlock bool
	var blockLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if !inBlock {
			if trimmedLine == doMark {
				inBlock = true
			}
		} else {
			if trimmedLine == doneMark {
				inBlock = false
				break
			}
			blockLines = append(blockLines, line)
		}
	}

	if inBlock {
		return "", fmt.Errorf("done mark '%s' not found", doneMark)
	}

	if len(blockLines) == 0 {
		return "", fmt.Errorf("no content found between do mark '%s' and done mark '%s'", doMark, doneMark)
	}

	return strings.Join(blockLines, "\n"), nil
}

// dedent removes common leading whitespace from all lines
func dedent(s string) string {
	lines := strings.Split(s, "\n")
	minIndent := -1

	// Find minimum indentation level
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			continue // Skip empty or whitespace-only lines
		}
		indent := len(line) - len(trimmed)
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	// Remove minimum indentation from each line
	if minIndent > 0 {
		for i, line := range lines {
			if len(line) >= minIndent {
				lines[i] = line[minIndent:]
			}
		}
	}

	return strings.Join(lines, "\n")
}
