package analyze

import (
	"context"
	"database/sql"
	"github.com/jlewi/foyle/app/pkg/logs/matchers"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/app/api"
	config "github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	"go.uber.org/zap"
)

type testTuple struct {
	p        *sessionBuilder
	sessions *SessionsManager
}

func setup() (testTuple, error) {
	d := zap.NewDevelopmentConfig()
	d.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, err := d.Build()
	if err != nil {
		return testTuple{}, errors.Wrapf(err, "Failed to create logger")
	}
	zap.ReplaceGlobals(logger)

	cfg, err := config.NewWithTempDir()
	if err != nil {
		return testTuple{}, errors.Wrapf(err, "failed to create config")
	}

	db, err := sql.Open(SQLLiteDriver, cfg.GetSessionsDB())
	if err != nil {
		return testTuple{}, errors.Wrapf(err, "Failed to open database")
	}
	sessions, err := NewSessionsManager(db)
	if err != nil {
		return testTuple{}, errors.Wrapf(err, "Failed to create session manager")
	}

	p, err := NewSessionBuilder(sessions)

	if err != nil {
		return testTuple{}, err
	}

	return testTuple{
		p:        p,
		sessions: sessions,
	}, nil
}

func Test_ProcessLogEvent(t *testing.T) {
	tuple, err := setup()
	if err != nil {
		t.Fatalf("Setup failed: %+v", err)
	}
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
		"function":  matchers.LogEvents,
		"message":   "LogEvent",
		"eventId":   event.GetEventId(),
		"contextId": event.GetContextId(),
		"time":      float64(startTime.Unix()),
		"event":     logs.ZapProto("event", event).Interface,
	}

	// Process the log entry
	tuple.p.processLogEvent(entry)

	s, err := tuple.sessions.Get(context.Background(), event.GetContextId())
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

func Test_ProcessStreamGeneerate(t *testing.T) {
	tuple, err := setup()
	if err != nil {
		t.Fatalf("Setup failed: %+v", err)
	}

	if err != nil {
		t.Fatalf("Setup failed: %+v", err)
	}
	contextId := "context1"

	fullContext := &v1alpha1.FullContext{
		Selected: 5,
		Notebook: &parserv1.Notebook{
			Cells: []*parserv1.Cell{
				{
					Kind:  parserv1.CellKind_CELL_KIND_MARKUP,
					Value: "cellcontents",
				},
			},
		},
	}
	// Create a log entry for the LogEvent
	entry := &api.LogEntry{
		"function":  matchers.StreamGenerate,
		"context":   logs.ZapProto("context", fullContext).Interface,
		"contextId": contextId,
	}

	// Process the log entry
	tuple.p.processLogEntry(entry)

	s, err := tuple.sessions.Get(context.Background(), contextId)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	opts := cmpopts.IgnoreUnexported(v1alpha1.FullContext{}, parserv1.Notebook{}, parserv1.Cell{})

	if d := cmp.Diff(fullContext, s.GetFullContext(), opts); d != "" {
		t.Errorf("Unexpected diff in full context:\n%v", d)
	}
}
