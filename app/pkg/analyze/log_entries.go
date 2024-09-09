package analyze

import (
	"context"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/fnames"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// logEntryProcessor processes log entries.
// Processing a log entry has the following steps
//  1. Identifying the type of log entry based on LogEntry properties such as callstack location
//  2. Delegating to the appropriate handler for that specific call site.
//  3. Updating the DB where the processed logs are stored.
//
// The 3rd step usually involves updating a database via a ReadModifyWrite operation.
// To make it easy to unittest we may use interfaces to abstract the ReadModifyWrite operation. This way in the
// unittest we don't need to construct a real database.
type logEntryProcessor struct {
	sessions *SessionsManager
}

func NewLogEntryProcessor(sessions *SessionsManager) *logEntryProcessor {
	return &logEntryProcessor{
		sessions: sessions,
	}
}

// n.b. processLogEntry doesn't return a error because we expect errors to be ignored and for processing to continue.
func (p *logEntryProcessor) processLogEntry(entry *api.LogEntry) {
	if entry.Function() == fnames.LogEvents {
		// TODO(Jeremy): There is also Analyzer.processLogEvent
		p.processLogEvent(entry)
	}

	if entry.Function() == fnames.StreamGenerate {
		p.processStreamGenerate(entry)
	}
}

func (p *logEntryProcessor) processLogEvent(entry *api.LogEntry) {
	log := zapr.NewLogger(zap.L())
	event := &v1alpha1.LogEvent{}

	if !entry.GetProto("event", event) {
		log.Error(errors.New("Failed to decode event"), "Failed to decode LogEvent", "entry", entry)
		return
	}

	// Update the session with the log event
	if event.GetContextId() == "" {
		log.Error(errors.New("LogEvent missing ContextId"), "LogEvent missing ContextId", "event", event)
		return
	}

	updateFunc := func(s *logspb.Session) error {
		for _, e := range s.LogEvents {
			if e.GetEventId() == event.GetEventId() {
				return nil
			}
		}
		if s.LogEvents == nil {
			s.LogEvents = make([]*v1alpha1.LogEvent, 0, 5)
		}

		if event.Type == v1alpha1.LogEventType_SESSION_START {
			s.StartTime = timestamppb.New(entry.Time())
		} else if event.Type == v1alpha1.LogEventType_SESSION_END {
			s.EndTime = timestamppb.New(entry.Time())
		}
		s.LogEvents = append(s.LogEvents, event)
		return nil
	}

	if err := p.sessions.Update(context.Background(), event.GetContextId(), updateFunc); err != nil {
		log.Error(err, "Failed to update session", "event", event)
	}
}

func (p *logEntryProcessor) processStreamGenerate(entry *api.LogEntry) {
	log := zapr.NewLogger(zap.L())
	contextId, ok := entry.GetString("contextId")
	if !ok {
		return
	}
	fullContext := &v1alpha1.FullContext{}
	if ok := entry.GetProto("context", fullContext); !ok {
		return
	}

	updateFunc := func(s *logspb.Session) error {
		s.FullContext = fullContext
		return nil
	}

	if err := p.sessions.Update(context.Background(), contextId, updateFunc); err != nil {
		log.Error(err, "Failed to update session", "contextId", contextId)
	}
}
