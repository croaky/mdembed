package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestProcessMarkdown(t *testing.T) {
	dir := "example"

	mdPath := filepath.Join(dir, "example.md")
	inData, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	wantPath := filepath.Join(dir, "want_output.md")
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWD)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	var outBuffer bytes.Buffer
	inReader := bytes.NewReader(inData)

	err = processMarkdown(inReader, &outBuffer)
	if err != nil {
		t.Fatalf("processMarkdown error: %v", err)
	}

	got := outBuffer.Bytes()

	if string(got) != string(want) {
		err := os.WriteFile("got_output.md", got, 0644)
		if err != nil {
			t.Fatalf("Failed to write: %v", err)
		}

		t.Errorf("Got does not match want.\nSee %s/got_output.md", dir)
	}
}
