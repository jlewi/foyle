package api

import (
	"encoding/json"
	"testing"
	"time"
)

func Test_Deserialize(t *testing.T) {
	type testCase struct {
		name    string
		input   string
		message string
		time    time.Time
		traceID string
	}

	cases := []testCase{
		{
			name:    "simple",
			input:   `{"severity":"INFO","time":1713208269,"message":"hello","traceId":"1234","blockIds":["1","2"]}`,
			message: "hello",
			time:    time.Unix(1713208269, 0),
			traceID: "1234",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var actual LogEntry
			if err := json.Unmarshal([]byte(c.input), &actual); err != nil {
				t.Fatalf("failed to deserialize: %v", err)
			}

			if actual.Message() != c.message {
				t.Errorf("Expected message %v but got %v", c.message, actual.Message())

			}
			if actual.Time() != c.time {
				t.Errorf("Expected time %v but got %v", c.time, actual.Time())
			}
			if actual.TraceID() != c.traceID {
				t.Errorf("Expected traceID %v but got %v", c.traceID, actual.TraceID())
			}
		})
	}
}
