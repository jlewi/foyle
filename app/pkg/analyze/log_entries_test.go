package analyze

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/app/api"
	config "github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/fnames"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"go.uber.org/zap"
	"testing"
	"time"
)

func Test_ProcessLogEvent(t *testing.T) {
	d := zap.NewDevelopmentConfig()
	d.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, err := d.Build()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	zap.ReplaceGlobals(logger)

	cfg, err := config.NewWithTempDir()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	sessions, err := NewSessionsManager(*cfg)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	p := NewLogEntryProcessor(sessions)

	event := &v1alpha1.LogEvent{
		EventId:   "event1",
		ContextId: "context1",
		Type:      v1alpha1.LogEventType_SESSION_START,
	}

	// Define the layout string
	layout := "2006-01-02 15:04:05 MST"

	// Parse the time string
	timeStr := "2024-08-09 13:24:55 PST"
	startTime, err := time.Parse(layout, timeStr)
	if err != nil {
		t.Fatalf("Failed to parse time: %v", err)
	}

	// Create a log entry for the LogEvent
	entry := &api.LogEntry{
		"function":  fnames.LogEvents,
		"message":   "LogEvent",
		"eventId":   event.GetEventId(),
		"contextId": event.GetContextId(),
		"time":      float64(startTime.Unix()),
		"event":     logs.ZapProto("event", event).Interface,
	}

	// Process the log entry
	p.processLogEvent(entry)

	s, err := sessions.Get(context.Background(), event.GetContextId())
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if len(s.LogEvents) != 1 {
		t.Fatalf("Expected 1 log event but got %v", len(s.LogEvents))
	}

	opts := cmpopts.IgnoreUnexported(v1alpha1.LogEvent{})
	if d := cmp.Diff(event, s.LogEvents[0], opts); d != "" {
		t.Fatalf("Unexpected diff in log event:\n%v", d)
	}

	// Verify that session time is set to the start of the session message.
	if s.StartTime.AsTime().UTC() != startTime.UTC() {
		t.Errorf("Expected start time to be %v but got %v", startTime.UTC(), s.StartTime.AsTime().UTC())
	}
}
