package analyze

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"

	"github.com/jlewi/foyle/app/pkg/logs/matchers"

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

	// If the directory doesn't exit opening the SQLLite database will fail.
	sessionsDBFile := cfg.GetSessionsDB()
	dbDir := filepath.Dir(sessionsDBFile)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return testTuple{}, errors.Wrapf(err, "Failed to create directory: %v", dbDir)
	}

	db, err := sql.Open(SQLLiteDriver, sessionsDBFile)
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

// Process the log entry
func testNotifier(session *logspb.Session) error {
	fmt.Printf("Received session end event for context: %v", session.GetContextId())
	return nil
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

	tuple.p.processLogEvent(entry, testNotifier)

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

func Test_ProcessStreamGenerate(t *testing.T) {
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
	tuple.p.processLogEntry(entry, testNotifier)

	s, err := tuple.sessions.Get(context.Background(), contextId)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	opts := cmpopts.IgnoreUnexported(v1alpha1.FullContext{}, parserv1.Notebook{}, parserv1.Cell{})

	if d := cmp.Diff(fullContext, s.GetFullContext(), opts); d != "" {
		t.Errorf("Unexpected diff in full context:\n%v", d)
	}
}

func Test_processLLMUsage(t *testing.T) {
	type testCase struct {
		name     string
		usage    *api.LLMUsage
		session  *logspb.Session
		expected *logspb.Session
	}

	cases := []testCase{
		{
			name: "empty",
			usage: &api.LLMUsage{
				InputTokens:  10,
				OutputTokens: 20,
			},
			session: &logspb.Session{},
			expected: &logspb.Session{
				TotalInputTokens:  10,
				TotalOutputTokens: 20,
			},
		},
		{
			name: "sum",
			usage: &api.LLMUsage{
				InputTokens:  10,
				OutputTokens: 20,
			},
			session: &logspb.Session{
				TotalInputTokens:  10,
				TotalOutputTokens: 20,
			},
			expected: &logspb.Session{
				TotalInputTokens:  20,
				TotalOutputTokens: 40,
			},
		},
	}

	opts := cmpopts.IgnoreUnexported(logspb.Session{})

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := updateSessionFromUsage(*c.usage, c.session); err != nil {
				t.Fatalf("Failed to update session: %v", err)
			}

			if d := cmp.Diff(c.expected, c.session, opts); d != "" {
				t.Errorf("Unexpected diff in session:\n%v", d)
			}
		})
	}
}
