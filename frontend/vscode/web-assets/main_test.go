package main

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_run(t *testing.T) {
	SetupLogging()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting current directory; %v", err)
	}

	vscodeDir := filepath.Join(cwd, "test_data")

	outDir, err := os.MkdirTemp("", "vscodeWebAssetsRunTest")
	if err != nil {
		t.Fatalf("Error creating temp dir; %v", err)
	}

	t.Logf("Running run with vscodeDir: %s, outDir: %s", vscodeDir, outDir)
	if err := run(vscodeDir, outDir); err != nil {
		t.Fatalf("Error running run; %v", err)
	}

	expected := []string{
		"extensions/markdown-basics/package.json",
		"extensions/some-other-ext/package.json",
		"out-vscode-reh-web-min/somefile.txt",
		"resources/someresource.txt",
	}

	missing := make([]string, 0, len(expected))

	for _, e := range expected {
		path := filepath.Join(outDir, e)

		_, err := os.Stat(path)

		if err != nil {
			missing = append(missing, e)
		}
	}

	if len(missing) > 0 {
		t.Fatalf("Missing files: %v", missing)
	}
}
