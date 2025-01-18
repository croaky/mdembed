package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestProcessMarkdown(t *testing.T) {
	// Set up test inputs and expected outputs
	exampleDir := "example"

	// Read the example.md file
	exampleMDPath := filepath.Join(exampleDir, "example.md")
	inputData, err := os.ReadFile(exampleMDPath)
	if err != nil {
		t.Fatalf("Failed to read input file: %v", err)
	}

	// Expected output, define the expected output string
	expectedOutputPath := filepath.Join(exampleDir, "expected_output.md")
	expectedOutputData, err := os.ReadFile(expectedOutputPath)
	if err != nil {
		t.Fatalf("Failed to read expected output file: %v", err)
	}

	// Save the current working directory
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalWD)

	// Change directory to the example directory to ensure files are found
	if err := os.Chdir(exampleDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	var outputBuffer bytes.Buffer
	inputReader := bytes.NewReader(inputData)

	// Call processMarkdown with input and output
	err = processMarkdown(inputReader, &outputBuffer)
	if err != nil {
		t.Fatalf("processMarkdown returned an error: %v", err)
	}

	// Get the output data
	outputData := outputBuffer.Bytes()

	// Compare the output with the expected output
	if !bytes.Equal(outputData, expectedOutputData) {
		// Write the actual output to a file for debugging
		err := ioutil.WriteFile("actual_output.md", outputData, 0644)
		if err != nil {
			t.Fatalf("Failed to write actual output to file: %v", err)
		}

		t.Errorf("Output does not match expected output.\nSee actual_output.md in %s for details.", exampleDir)
	}
}
