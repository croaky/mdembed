package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProcessMD(t *testing.T) {
	dir := "examples"

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd err: %v", err)
	}
	defer os.Chdir(wd)

	err = os.Chdir(dir)
	if err != nil {
		t.Fatalf("os.Chdir err: %v", err)
	}

	in, err := os.ReadFile("input1.md")
	if err != nil {
		t.Fatalf("os.ReadFile err: %v", err)
	}

	want, err := os.ReadFile("outwant.md")
	if err != nil {
		t.Fatalf("osReadFile err: %v", err)
	}

	var outBuf bytes.Buffer

	state := &ProcessState{
		filesInProcess: make(map[string]bool),
	}

	err = processMD(bytes.NewReader(in), &outBuf, state)
	if err != nil {
		t.Fatalf("processMD err: %v", err)
	}

	got := outBuf.Bytes()
	err = os.WriteFile("outgot.md", got, 0644)
	if err != nil {
		t.Fatalf("os.WriteFile err: %v", err)
	}

	if string(got) != string(want) {
		t.Logf("outwant.md:\n%s", string(want))
		t.Logf("outgot.md:\n%s", string(got))

		t.Errorf("got != want\nSee %s/outgot.md", dir)
	}
}

func TestProcessMD_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		file       string
		fileData   []byte
		wantErrMsg string
	}{
		{
			name:       "Unterminated embed block",
			input:      "Some text before.\n\n```embed\nfilename.go\n",
			wantErrMsg: "unterminated ```embed code block",
		},
		{
			name:       "Invalid format in embed code block",
			input:      "Some text before.\n\n```embed\nfilename.go extra words\n```\n",
			wantErrMsg: "invalid format in embed code block: filename.go extra words",
		},
		{
			name:       "File not found",
			input:      "Some text before.\n\n```embed\nnonexistent.go\n```\n",
			wantErrMsg: "no files match pattern nonexistent.go",
		},
		{
			name:       "Unsupported file type",
			input:      "Some text before.\n\n```embed\nfile.unknown\n```\n",
			file:       "file.unknown",
			fileData:   []byte("// content"),
			wantErrMsg: "unsupported file type: .unknown",
		},
		{
			name:       "Do mark not found",
			input:      "Some text before.\n\n```embed\nfile.go block1\n```\n",
			file:       "file.go",
			fileData:   []byte(`// This is a test file`),
			wantErrMsg: "do mark",
		},
		{
			name:  "Done mark not found",
			input: "Some text before.\n\n```embed\nfile.go block2\n```\n",
			file:  "file.go",
			fileData: []byte(`// emdo block2
    // Code block content`),
			wantErrMsg: "done mark",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}

			if tt.file != "" {
				if err := os.WriteFile(filepath.Join(dir, tt.file), tt.fileData, 0644); err != nil {
					t.Fatal(err)
				}
			}

			var outBuf bytes.Buffer

			state := &ProcessState{
				filesInProcess: make(map[string]bool),
			}

			err := processMD(strings.NewReader(tt.input), &outBuf, state)

			if err == nil {
				t.Errorf("Expected error but got none")
			} else if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("Expected error message to contain '%s', but got '%s'",
					tt.wantErrMsg, err.Error())
			}
		})
	}
}

// embedIn makes a temporary working directory with the given files,
// changes to it, and runs processMD on input. It returns the output
// and the error.
func embedIn(t *testing.T, files map[string]string, input string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	for name, data := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)
	var out bytes.Buffer
	err := processMD(strings.NewReader(input), &out, &ProcessState{
		filesInProcess: make(map[string]bool),
	})
	return out.String(), err
}

// outsideFile writes a secret file in a directory outside the working
// directory and returns the path of that directory.
func outsideFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "secret.yml"), []byte("SECRET"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestEmbedRefusesSymlinks(t *testing.T) {
	outside := outsideFile(t)

	tests := []struct {
		name  string
		link  string
		to    string
		input string
	}{
		{"file link out of the tree", "link.yml", filepath.Join(outside, "secret.yml"), "link.yml"},
		{"file link in the tree", "link.yml", "real.txt", "link.yml"},
		{"directory link by name", "linkdir", outside, "linkdir/secret.yml"},
		{"directory link by glob", "linkdir", outside, "linkdir/*.yml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "real.txt"), []byte("SECRET"), 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(tt.to, filepath.Join(dir, tt.link)); err != nil {
				t.Fatal(err)
			}
			t.Chdir(dir)
			var out bytes.Buffer
			err := processMD(strings.NewReader("```embed\n"+tt.input+"\n```\n"), &out, &ProcessState{
				filesInProcess: make(map[string]bool),
			})
			if err == nil {
				t.Errorf("got no error, want an error")
			}
			if strings.Contains(out.String(), "SECRET") {
				t.Errorf("output holds the link target:\n%s", out.String())
			}
		})
	}
}

func TestEmbedGlobSkipsDirectorySymlinks(t *testing.T) {
	outside := outsideFile(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.yml"), []byte("a: 1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "linkdir")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	var out bytes.Buffer
	err := processMD(strings.NewReader("```embed\n**/*.yml\n```\n"), &out, &ProcessState{
		filesInProcess: make(map[string]bool),
	})
	if err != nil {
		t.Fatalf("processMD err: %v", err)
	}
	if strings.Contains(out.String(), "SECRET") {
		t.Errorf("output holds a file under a directory link:\n%s", out.String())
	}
}

func TestEmbedRefusesFileInDirectorySymlink(t *testing.T) {
	for _, input := range []string{"alias/a.yml", "alias/*.yml"} {
		t.Run(input, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.Mkdir(filepath.Join(dir, "real"), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "real", "a.yml"), []byte("a: 1"), 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("real", filepath.Join(dir, "alias")); err != nil {
				t.Fatal(err)
			}
			t.Chdir(dir)
			var out bytes.Buffer
			err := processMD(strings.NewReader("```embed\n"+input+"\n```\n"), &out, &ProcessState{
				filesInProcess: make(map[string]bool),
			})
			if err == nil || !strings.Contains(err.Error(), "directory symlink") {
				t.Errorf("got error %v, want a directory symlink error", err)
			}
		})
	}
}

func TestEmbedRefusesBraces(t *testing.T) {
	// With no limit, the glob expands 40 groups of 2 to 2^40 patterns.
	pattern := strings.Repeat("{a,b}", 40) + ".sh"
	start := time.Now()
	_, err := embedIn(t, map[string]string{"a.sh": "echo hi"}, "```embed\n"+pattern+"\n```\n")
	if err == nil || !strings.Contains(err.Error(), "brace") {
		t.Errorf("got error %v, want a brace error", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("brace pattern ran for %v, want less than 1s", d)
	}
}

func TestEmbedNestedMarkdownLimit(t *testing.T) {
	// Each file embeds the next file two times, so a run with no limit
	// writes the leaf 2^40 times.
	const depth = 40
	files := map[string]string{fmt.Sprintf("l%d.md", depth): "leaf"}
	for i := range depth {
		next := fmt.Sprintf("l%d.md", i+1)
		files[fmt.Sprintf("l%d.md", i)] = "```embed\n" + next + "\n" + next + "\n```\n"
	}

	start := time.Now()
	_, err := embedIn(t, files, "```embed\nl0.md\n```\n")
	if err == nil || !strings.Contains(err.Error(), "limit") {
		t.Errorf("got error %v, want a limit error", err)
	}
	if d := time.Since(start); d > 10*time.Second {
		t.Errorf("nested embeds ran for %v, want less than 10s", d)
	}
}

func TestEmbedByteLimit(t *testing.T) {
	big := strings.Repeat("x", maxEmbedBytes+1)
	_, err := embedIn(t, map[string]string{"big.sh": big}, "```embed\nbig.sh\n```\n")
	if err == nil || !strings.Contains(err.Error(), "limit") {
		t.Errorf("got error %v, want a limit error", err)
	}
}

func TestEmbedFenceLongerThanContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "no backticks",
			content: "echo hi",
			want:    "```sh\n# a.sh\necho hi\n```\n",
		},
		{
			name:    "three backticks",
			content: "cat <<EOF\n```\n<b>live</b>\nEOF",
			want:    "````sh\n# a.sh\ncat <<EOF\n```\n<b>live</b>\nEOF\n````\n",
		},
		{
			name:    "indented five backticks",
			content: "  x\n  `````\r\n  y",
			want:    "``````sh\n# a.sh\nx\n`````\r\ny\n``````\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := embedIn(t, map[string]string{"a.sh": tt.content}, "```embed\na.sh\n```\n")
			if err != nil {
				t.Fatalf("processMD err: %v", err)
			}
			if got != tt.want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestEmbedRefusesControlCharacterInName(t *testing.T) {
	name := "a\n```\n<b>live\n.sh"
	got, err := embedIn(t, map[string]string{name: "echo hi"}, "```embed\n*.sh\n```\n")
	if err == nil {
		t.Errorf("got no error, want an error")
	}
	if strings.Contains(got, "<b>live") {
		t.Errorf("output holds the file name lines:\n%s", got)
	}
}
