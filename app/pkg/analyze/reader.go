package analyze

import (
	"bufio"
	"context"
	"os"

	"github.com/pkg/errors"
)

// readLinesFromOffset reads lines from a file starting at the given offset.
// It will read until the end of the file.
func readLinesFromOffset(ctx context.Context, path string, offset int64) ([]string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "failed to open file %s", path)
	}
	defer f.Close()

	offset, err = f.Seek(offset, 0)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "failed to seek to offset %d in file %s", offset, path)
	}

	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Split(ScanLinesNoPartial)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		offset += int64(len(line) + 1) // +1 for newline
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, errors.Wrapf(err, "failed to scan file %s", path)
	}
	return lines, offset, nil
}

// ScanLinesNoPartial is a split function for a [Scanner] that returns each line of
// text, stripped of any trailing end-of-line marker. The returned line may
// be empty. The end-of-line marker is one optional carriage return followed
// by one mandatory newline. In regular expression notation, it is `\r?\n`.
// The last non-empty line of input will not be returned if it has no newline.
// We do this because we assume if there is no newline that it is a partial record.
func ScanLinesNoPartial(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// If we're at EOF, we request more data because we treat the current line as a partial record
	if atEOF {
		return 0, nil, nil
	}
	return bufio.ScanLines(data, atEOF)
}
