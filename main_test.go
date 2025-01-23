package main

import (
	"bytes"
	"os"
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

	in, err := os.ReadFile("input.md")
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
		setup      func()
		teardown   func()
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
			wantErrMsg: "failed to read file nonexistent.go",
		},
		{
			name:  "Unsupported file type",
			input: "Some text before.\n\n```embed\nfile.unknown\n```\n",
			setup: func() {
				os.WriteFile("file.unknown", []byte("// content"), 0644)
			},
			teardown: func() {
				os.Remove("file.unknown")
			},
			wantErrMsg: "unsupported file type: .unknown",
		},
		{
			name:  "Do mark not found",
			input: "Some text before.\n\n```embed\nfile.go block1\n```\n",
			setup: func() {
				os.WriteFile("file.go", []byte(`// This is a test file`), 0644)
			},
			teardown: func() {
				os.Remove("file.go")
			},
			wantErrMsg: "do mark",
		},
		{
			name:  "Done mark not found",
			input: "Some text before.\n\n```embed\nfile.go block2\n```\n",
			setup: func() {
				os.WriteFile("file.go", []byte(`// emdo block2                                
    // Code block content`), 0644)
			},
			teardown: func() {
				os.Remove("file.go")
			},
			wantErrMsg: "done mark",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.teardown != nil {
				defer tt.teardown()
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
