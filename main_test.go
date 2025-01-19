package main

import (
	"bytes"
	"os"
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

	err = processMD(bytes.NewReader(in), &outBuf)
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
