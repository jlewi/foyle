package analyze

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"os"
	"testing"
)

func Test_readFromOffset(t *testing.T) {
	logFile, err := os.CreateTemp("", "test.log")
	if err != nil {
		t.Fatal(err)
	}

	// Write some data to the file
	if _, err := logFile.WriteString("line 1\nline 2\n"); err != nil {
		t.Fatal(err)
	}

	// Read the data from the file
	lines, offset, err := readLinesFromOffset(context.Background(), logFile.Name(), 0)

	if d := cmp.Diff(lines, []string{"line 1", "line 2"}); d != "" {
		t.Errorf("unexpected lines:\n%v", d)
	}

	if offset != 14 {
		t.Errorf("unexpected offset: %d", offset)
	}

	// Write some more data
	if _, err := logFile.WriteString("line 3\nline 4\n"); err != nil {
		t.Fatal(err)
	}

	// Read the data from the file and see that we properly carry on reading.
	newLines, offset, err := readLinesFromOffset(context.Background(), logFile.Name(), offset)
	if err != nil {
		t.Fatal(err)
	}

	if offset != 28 {
		t.Errorf("unexpected offset: %d", offset)
	}
	if d := cmp.Diff(newLines, []string{"line 3", "line 4"}); d != "" {
		t.Errorf("unexpected lines:\n%v", d)
	}

	// Now write an incomplete line and verify we don't read anything
	if _, err := logFile.WriteString("lin"); err != nil {
		t.Fatalf("failed to write to file: %v", err)
	}

	partialLines, offset, err := readLinesFromOffset(context.Background(), logFile.Name(), offset)
	if err != nil {
		t.Fatalf("failed to read from file: %v", err)
	}

	if len(partialLines) != 0 {
		t.Errorf("unexpected lines: %v", partialLines)
	}

	// Offset should be unchanged because this was a partial read so it doesn't get read
	if offset != 28 {
		t.Errorf("unexpected offset: %d", offset)
	}

	// Write the remainder of the line
	if _, err := logFile.WriteString("e 5\n"); err != nil {
		t.Fatalf("failed to write to file: %v", err)
	}

	lastLine, offset, err := readLinesFromOffset(context.Background(), logFile.Name(), offset)
	if err != nil {
		t.Fatal(err)
	}

	if offset != 35 {
		t.Errorf("unexpected offset: %d", offset)
	}
	if d := cmp.Diff(lastLine, []string{"line 5"}); d != "" {
		t.Errorf("unexpected lines:\n%v", d)
	}
}
