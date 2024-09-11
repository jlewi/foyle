package analyze

import (
	"context"
	"strings"

	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/fnames"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// sessionBuilder build sessions by processing log entries.
type sessionBuilder struct {
	sessions *SessionsManager
}

func NewSessionBuilder(sessions *SessionsManager) (*sessionBuilder, error) {
	return &sessionBuilder{
		sessions: sessions,
	}, nil
}

// n.b. processLogEntry doesn't return a error because we expect errors to be ignored and for processing to continue.
// I'm not sure that's a good idea but we'll see.
func (p *sessionBuilder) processLogEntry(entry *api.LogEntry) {
	// We need to use HasPrefix because the logging statement is nested inside an anonymous function so there
	// will be a suffix like "func1"
	if strings.HasPrefix(entry.Function(), fnames.LogEvents) {
		// TODO(Jeremy): There is also Analyzer.processLogEvent
		p.processLogEvent(entry)
	}

	if entry.Function() == fnames.StreamGenerate {
		p.processStreamGenerate(entry)
	}
}

func (p *sessionBuilder) processLogEvent(entry *api.LogEntry) {
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

func (p *sessionBuilder) processStreamGenerate(entry *api.LogEntry) {
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
