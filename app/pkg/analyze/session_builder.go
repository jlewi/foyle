package analyze

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/logs/matchers"
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
	if matchers.IsLogEvent(entry.Function()) {
		// TODO(Jeremy): There is also Analyzer.processLogEvent
		p.processLogEvent(entry)
	}

	if matchers.IsLLMUsage(entry.Function()) {
		p.processLLMUsage(entry)
	}

	if matchers.IsGenerate(entry.Function()) {
		p.processGenerate(entry)
	}

	if matchers.IsStreamGenerate(entry.Function()) {
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
		return updateSessionFromEvent(event, entry.Time(), s)
	}

	if err := p.sessions.Update(context.Background(), event.GetContextId(), updateFunc); err != nil {
		log.Error(err, "Failed to update session", "event", event)
	}
}

func (p *sessionBuilder) processLLMUsage(entry *api.LogEntry) {
	log := zapr.NewLogger(zap.L())
	usage := &api.LLMUsage{}
	if !entry.GetStruct("usage", usage) {
		log.Error(errors.New("Failed to decode usage"), "Failed to decode LLMUsage", "entry", entry)
		return
	}
	contextId, ok := entry.GetString("contextId")
	if !ok {
		log.Error(errors.New("Failed to handle LLMUsage log entry"), "LLMUsage is missing contextId", "entry", entry)
		return
	}

	updateFunc := func(s *logspb.Session) error {
		return updateSessionFromUsage(*usage, s)
	}

	if err := p.sessions.Update(context.Background(), contextId, updateFunc); err != nil {
		log.Error(err, "Failed to update session", "usage", usage)
	}
}

func (p *sessionBuilder) processGenerate(entry *api.LogEntry) {
	log := zapr.NewLogger(zap.L())
	contextId, ok := entry.GetString("contextId")
	if !ok {
		return
	}

	traceId := entry.TraceID()
	if traceId == "" {
		log.Error(errors.New("Failed to handle Agent.Generate log entry"), "Agent.Generate is missing traceId", "entry", entry)
		return

	}

	updateFunc := func(s *logspb.Session) error {
		if s.GenerateTraceIds == nil {
			s.GenerateTraceIds = make([]string, 0, 5)
		}
		s.GenerateTraceIds = append(s.GenerateTraceIds, traceId)
		return nil
	}

	if err := p.sessions.Update(context.Background(), contextId, updateFunc); err != nil {
		log.Error(err, "Failed to update session", "contextId", contextId)
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

// updateSessionFromUsage updates the session with the usage
func updateSessionFromUsage(usage api.LLMUsage, s *logspb.Session) error {
	s.TotalInputTokens = s.GetTotalInputTokens() + int32(usage.InputTokens)
	s.TotalOutputTokens = s.GetTotalOutputTokens() + int32(usage.OutputTokens)
	return nil
}

// updateSessionFromEvent updates the session with the log event.
func updateSessionFromEvent(event *v1alpha1.LogEvent, eventTime time.Time, s *logspb.Session) error {
	if event.Type == v1alpha1.LogEventType_SESSION_START {
		// N.B. We want to support backfilling/reprocessing logs. To do this, everytime we encounter
		// A session start event we want to zero out the proto. This is important for fields that represent
		// accumulations e.g. total_input_tokens, total_output_tokens.
		proto.Reset(s)
		s.ContextId = event.GetContextId()
	}
	if s.LogEvents == nil {
		s.LogEvents = make([]*v1alpha1.LogEvent, 0, 5)
	}

	for _, e := range s.LogEvents {
		if e.GetEventId() == event.GetEventId() {
			return nil
		}
	}

	if event.Type == v1alpha1.LogEventType_SESSION_START {
		s.StartTime = timestamppb.New(eventTime)
	} else if event.Type == v1alpha1.LogEventType_SESSION_END {
		s.EndTime = timestamppb.New(eventTime)
	}
	s.LogEvents = append(s.LogEvents, event)
	return nil
}
