package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
