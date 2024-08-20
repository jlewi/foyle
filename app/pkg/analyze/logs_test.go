package analyze

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadAnthropicLog(t *testing.T) {
	type testCase struct {
		name    string
		logFile string
		traceId string
	}
	cases := []testCase{
		{
			name:    "basic",
			logFile: "anthropic_logs.jsonl",
			traceId: "8d63ff91499d2c09be43be8a94d80d43",
		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed to get current working directory")
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fullPath := filepath.Join(cwd, "test_data", c.logFile)
			result, err := readAnthropicLog(context.Background(), c.traceId, fullPath)
			if err != nil {
				t.Errorf("Failed to read Anthropic request: %v", err)
			}
			if result == nil {
				t.Fatalf("Request should not be nil")
			}
			if result.Request == nil {
				t.Errorf("Request should not be nil")
			}
			if result.Response == nil {
				t.Errorf("Response should not be nil")
			} else {
				if result.Response.Model == "" {
					t.Errorf("Model should not be empty")
				}
			}
		})
	}
}

func TestGetLogFilesSorted(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "logtest")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after the test

	// Create test log files
	testFiles := []string{
		"foyle.logs.2024-08-19T15:44:42.json",
		"foyle.logs.2024-08-16T16:43:16.json",
		"foyle.logs.2024-08-14T10:46:47.json",
		"not_a_log_file.txt",
	}

	for _, file := range testFiles {
		_, err := os.Create(filepath.Join(tempDir, file))
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Run the function
	result, err := getLogFilesSorted(tempDir)

	// Check for errors
	if err != nil {
		t.Fatalf("getLogFilesSorted returned an error: %v", err)
	}

	// Expected result (sorted in descending order)
	expected := []string{
		"foyle.logs.2024-08-19T15:44:42.json",
		"foyle.logs.2024-08-16T16:43:16.json",
		"foyle.logs.2024-08-14T10:46:47.json",
	}

	// Compare results
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("getLogFilesSorted returned %v, want %v", result, expected)
	}
}
