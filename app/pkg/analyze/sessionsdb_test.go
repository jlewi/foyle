package analyze

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/app/pkg/config"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/utils/temp"
	"testing"
	"time"
)

func Test_SessionsCRUD(t *testing.T) {
	dir, err := temp.CreateTempDir("sessionsDBTest")
	if err != nil {
		t.Fatalf("Error creating temp dir: %v", err)
	}

	cfg := config.Config{}
	// TO control the location of the sqllite database we have to set LogDir
	cfg.Logging = config.Logging{
		LogDir: dir.Name,
	}

	m, err := NewSessionsManager(cfg)
	if err != nil {
		t.Fatalf("Error creating SessionsManager: %v", err)
	}

	cid := "1"

	expected := &logspb.Session{
		ContextId: cid,
		StartTime: timestamppb.New(time.Now()),
		EndTime:   timestamppb.New(time.Now().Add(time.Hour)),
		LogEvents: []*v1alpha1.LogEvent{
			{
				Type: v1alpha1.LogEventType_ACCEPTED,
			},
		},
	}

	if err := m.Update(context.Background(), cid, func(s *logspb.Session) error {
		s.ContextId = cid
		s.StartTime = expected.StartTime
		s.EndTime = expected.EndTime
		s.LogEvents = expected.LogEvents
		return nil
	}); err != nil {
		t.Fatalf("Error updating session: %v", err)
	}

	// TODO(jeremy): Try to read the session back and verify it was written correctly.
	actual, err := m.Get(context.Background(), cid)
	if err != nil {
		t.Fatalf("Error getting session: %v", err)
	}

	comparer := cmpopts.IgnoreUnexported(logspb.Session{}, v1alpha1.LogEvent{}, timestamppb.Timestamp{})
	if d := cmp.Diff(actual, expected, comparer); d != "" {
		t.Fatalf("Unexpected diff between expected and actual session:\n%v", d)
	}
}
