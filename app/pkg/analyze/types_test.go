package analyze

import (
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func Test_Deserialize(t *testing.T) {
	type testCase struct {
		name     string
		input    string
		expected LogEntry
	}

	cases := []testCase{
		{
			name:  "simple",
			input: `{"severity":"INFO","time":1713208269.374795,"message":"hello","traceId":"1234","blockIds":["1","2"]}`,
			expected: LogEntry{
				Severity: "INFO",
				Time:     1713208269.374795,
				Message:  "hello",
				TraceID:  "1234",
				BlockIds: []string{"1", "2"},
			},
		},
		{
			name:  "extra_fields",
			input: `{"severity":"INFO","field1": "a", "field2":"b"}`,
			expected: LogEntry{
				Severity: "INFO",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var actual LogEntry
			if err := json.Unmarshal([]byte(c.input), &actual); err != nil {
				t.Fatalf("failed to deserialize: %v", err)
			}
			if d := cmp.Diff(c.expected, actual); d != "" {
				t.Errorf("Unexpected diff:\n%s", d)
			}
		})
	}
}
