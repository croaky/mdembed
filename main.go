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
	"unicode"

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
	".csv":   {NoComment: true},                                     // CSV
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
	".json":  {NoComment: true},                                     // JSON
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

// A run stops with an error when it embeds more than maxEmbedFiles
// files or reads more than maxEmbedBytes bytes from them. Markdown
// files that embed other Markdown files more than one time can
// otherwise make a small input expand without limit.
const (
	maxEmbedFiles = 10000
	maxEmbedBytes = 64 << 20
)

type ProcessState struct {
	filesInProcess map[string]bool
	filesEmbedded  int
	bytesEmbedded  int
}

type Style struct {
	LineComment string
	BlockDo     string
	BlockDone   string
	NoComment   bool
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

		if err := embedPattern(pattern, blockName, output, state); err != nil {
			return err
		}

		// Add newline between multiple patterns
		if i < len(lines)-1 {
			fmt.Fprintln(output)
		}
	}
	return nil
}

// embedPattern embeds each file that matches pattern in the working
// directory. An os.Root holds each access in the working directory, so
// a symlink cannot move a read outside it. The glob does not descend
// through a directory symlink. embedPattern refuses a brace, because
// the glob expands n brace groups to k^n patterns before it matches.
func embedPattern(pattern, blockName string, output io.Writer, state *ProcessState) error {
	if strings.ContainsRune(pattern, '{') {
		return fmt.Errorf("pattern %s holds a brace, which mdembed does not support", pattern)
	}

	root, err := os.OpenRoot(".")
	if err != nil {
		return err
	}
	defer root.Close()

	matches, err := doublestar.Glob(root.FS(), pattern, doublestar.WithNoFollow())
	if err != nil {
		return fmt.Errorf("failed to glob pattern %s: %v", pattern, err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("no files match pattern %s", pattern)
	}

	for j, filename := range matches {
		if strings.IndexFunc(filename, unicode.IsControl) >= 0 {
			return fmt.Errorf("file name %q holds a control character", filename)
		}

		state.filesEmbedded++
		if state.filesEmbedded > maxEmbedFiles {
			return fmt.Errorf("embed limit of %d files reached at %s", maxEmbedFiles, filename)
		}

		content, err := readEmbed(root, filename, state)
		if err != nil {
			return err
		}

		if err := processFile(filename, blockName, string(content), output, state); err != nil {
			return err
		}

		// Add newline between multiple code blocks
		if j < len(matches)-1 {
			fmt.Fprintln(output)
		}
	}
	return nil
}

// readEmbed reads a regular file in root. It refuses a symlink, a
// directory, or a device, and it reads no more bytes than the run has
// left of maxEmbedBytes. It also refuses a file in a directory symlink,
// so each file has one name.
func readEmbed(root *os.Root, filename string, state *ProcessState) ([]byte, error) {
	for dir := path.Dir(filename); dir != "." && dir != "/"; dir = path.Dir(dir) {
		info, err := root.Lstat(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", filename, err)
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return nil, fmt.Errorf("%s is in a directory symlink %s", filename, dir)
		}
	}

	info, err := root.Lstat(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filename, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", filename)
	}

	f, err := root.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filename, err)
	}
	defer f.Close()
	if info, err := f.Stat(); err != nil || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", filename)
	}

	left := maxEmbedBytes - state.bytesEmbedded
	content, err := io.ReadAll(io.LimitReader(f, int64(left)+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filename, err)
	}
	if len(content) > left {
		return nil, fmt.Errorf("embed limit of %d bytes reached at %s", maxEmbedBytes, filename)
	}
	state.bytesEmbedded += len(content)
	return content, nil
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

	// Prepare header comment with filename
	var headerComment string
	if style.LineComment != "" {
		headerComment = style.LineComment + " " + filename
	} else if style.BlockDo != "" && style.BlockDone != "" {
		headerComment = fmt.Sprintf("%s %s %s", style.BlockDo, filename, style.BlockDone)
	} else {
		// For file types without comments (like JSON), just use the filename
		headerComment = filename
	}

	// If a block name is specified and the file doesn't support comments, return an error
	if blockName != "" && style.NoComment {
		return fmt.Errorf("block embedding is not supported for files that do not support comments, such as %s", filename)
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

	// Write code block to output. The fence is longer than each
	// backtick run in the block, so no line in it can close the fence.
	fence := strings.Repeat("`", max(3, longestBacktickRun(headerComment+"\n"+fileContent)+1))
	fmt.Fprintf(output, "%s%s\n", fence, lang)
	fmt.Fprintf(output, "%s\n", headerComment)
	fmt.Fprintf(output, "%s", fileContent)
	fmt.Fprintf(output, "\n%s\n", fence)

	return nil
}

// longestBacktickRun returns the length of the longest run of
// backticks in s.
func longestBacktickRun(s string) int {
	longest, run := 0, 0
	for i := 0; i < len(s); i++ {
		if s[i] == '`' {
			run++
			longest = max(longest, run)
		} else {
			run = 0
		}
	}
	return longest
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
		// For files without comments, return empty markers
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
