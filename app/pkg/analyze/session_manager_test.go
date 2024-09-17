package analyze

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jlewi/foyle/app/pkg/analyze/fsql"
	"github.com/jlewi/foyle/app/pkg/runme/converters"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	"google.golang.org/protobuf/proto"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/utils/temp"
)

func Test_SessionsCRUD(t *testing.T) {
	dir, err := temp.CreateTempDir("sessionsDBTest")
	if err != nil {
		t.Fatalf("Error creating temp dir: %v", err)
	}

	dbPath := filepath.Join(dir.Name, "sessions.db")
	db, err := sql.Open(SQLLiteDriver, dbPath)
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}
	m, err := NewSessionsManager(db)
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

var (
	session1 = &logspb.Session{
		ContextId: "1",
		FullContext: &v1alpha1.FullContext{
			Notebook: &parserv1.Notebook{
				Cells: []*parserv1.Cell{
					{
						Kind:  parserv1.CellKind_CELL_KIND_MARKUP,
						Value: "This is cell 1",
					},
					{
						Kind:  parserv1.CellKind_CELL_KIND_CODE,
						Value: "This should not be the answer",
						Metadata: map[string]string{
							converters.RunmeIdField: "0123345-id",
						},
					},
					{
						Kind:  parserv1.CellKind_CELL_KIND_MARKUP,
						Value: "This cell should not be in the response because it is after the executed cell",
					},
				},
			},
			Selected: 1,
		},
		LogEvents: []*v1alpha1.LogEvent{
			{
				Type: v1alpha1.LogEventType_EXECUTE,
				Cells: []*parserv1.Cell{
					{
						Kind:  parserv1.CellKind_CELL_KIND_CODE,
						Value: "Actual cell executed",
					},
				},
				SelectedIndex: 1,
			},
		},
	}
)

func Test_getExampleFromSession(t *testing.T) {
	type testCase struct {
		name     string
		session  *logspb.Session
		expected *v1alpha1.EvalExample
	}

	cases := []testCase{
		{
			name:    "Basic",
			session: session1,
			expected: &v1alpha1.EvalExample{
				Id: "1",
				FullContext: &v1alpha1.FullContext{
					Notebook: &parserv1.Notebook{
						Cells: []*parserv1.Cell{
							{
								Kind:  parserv1.CellKind_CELL_KIND_MARKUP,
								Value: "This is cell 1",
							},
						},
					},
					Selected: 0,
				},
				ExpectedCells: []*parserv1.Cell{
					{
						Kind:  parserv1.CellKind_CELL_KIND_CODE,
						Value: "Actual cell executed",
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := getExampleFromSession(c.session)
			if err != nil {
				t.Fatalf("Error getting example from session: %v", err)
			}

			comparer := cmpopts.IgnoreUnexported(v1alpha1.EvalExample{}, v1alpha1.FullContext{}, parserv1.Notebook{}, parserv1.Cell{})
			if d := cmp.Diff(actual, c.expected, comparer); d != "" {
				t.Fatalf("Unexpected diff between expected and actual example:\n%v", d)
			}
		})
	}
}

func Test_DumpExamples(t *testing.T) {
	// This test verifies the logic to dump the examples from the database.
	// It doesn't test the logic to convert sessions to examples because that is tested via the Test_getExampleFromSession
	// unittest
	dir, err := temp.CreateTempDir("sessionsDBTest")
	if err != nil {
		t.Fatalf("Error creating temp dir: %v", err)
	}

	db, err := sql.Open(SQLLiteDriver, filepath.Join(dir.Name, "sessions.db"))
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}
	manager, err := NewSessionsManager(db)

	if err != nil {
		t.Fatalf("Error creating SessionsManager: %v", err)
	}
	// Write some sessions to the database
	if err := manager.Update(context.Background(), session1.ContextId, func(s *logspb.Session) error {
		s.FullContext = session1.FullContext
		s.LogEvents = session1.LogEvents
		s.ContextId = session1.ContextId
		return nil
	}); err != nil {
		t.Fatalf("Error writing sessions to the DB: %v", err)
	}

	request := &logspb.DumpExamplesRequest{
		Output: filepath.Join(dir.Name, "examples"),
	}
	resp, err := manager.DumpExamples(context.Background(), connect.NewRequest(request))

	if err != nil {
		t.Fatalf("Error dumping examples: %v", err)
	}
	if resp.Msg.GetNumExamples() != 1 {
		t.Fatalf("Expected 1 example but got %v", resp.Msg.GetNumExamples())
	}
	t.Logf("Dumped examples to: %v", request.GetOutput())
}

func Test_protoToRow(t *testing.T) {

	type testCase struct {
		name     string
		session  *logspb.Session
		expected *fsql.Session
	}

	sess1Bytes, err := proto.Marshal(session1)
	if err != nil {
		t.Fatalf("Error marshalling session1: %v", err)
	}
	cases := []testCase{
		{
			name:    "Basic",
			session: session1,
			expected: &fsql.Session{
				Contextid:    "1",
				Starttime:    session1.GetStartTime().AsTime(),
				Endtime:      session1.GetEndTime().AsTime(),
				Selectedid:   "0123345-id",
				Selectedkind: "CELL_KIND_CODE",
				Proto:        sess1Bytes,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := protoToRow(c.session)
			if err != nil {
				t.Fatalf("Error converting session to row: %v", err)
			}
			comparer := cmpopts.IgnoreUnexported(fsql.Session{}, time.Time{})
			if d := cmp.Diff(actual, c.expected, comparer); d != "" {
				t.Fatalf("Unexpected diff between expected and actual session:\n%v", d)
			}
		})
	}
}
