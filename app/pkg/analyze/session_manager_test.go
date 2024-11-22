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
		StartTime: timeMustParse(time.RFC3339, "2021-01-01T00:01:00Z"),
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
				Id:   "1",
				Time: timeMustParse(time.RFC3339, "2021-01-01T00:01:00Z"),
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

			comparer := cmpopts.IgnoreUnexported(v1alpha1.EvalExample{}, v1alpha1.FullContext{}, parserv1.Notebook{}, parserv1.Cell{}, timestamppb.Timestamp{})
			if d := cmp.Diff(actual, c.expected, comparer); d != "" {
				t.Fatalf("Unexpected diff between expected and actual example:\n%v", d)
			}
		})
	}
}

func Test_RetriesOnLocked(t *testing.T) {
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

	db2, err := sql.Open(SQLLiteDriver, filepath.Join(dir.Name, "sessions.db"))
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}
	manager, err := NewSessionsManager(db2)

	if err != nil {
		t.Fatalf("Error creating SessionsManager: %v", err)
	}

	// Insert a session into the database
	//if err := manager.Update(context.Background(), session1.ContextId, func(s *logspb.Session) error {
	//	// Signal main thread that write has started
	//	s.FullContext = session1.FullContext
	//	s.LogEvents = session1.LogEvents
	//	s.ContextId = session1.ContextId
	//
	//	return nil
	//}); err != nil {
	//	t.Fatalf("Error writing sessions to the DB: %v", err)
	//}

	blockUpdate := make(chan bool, 1)

	blockRead := make(chan bool, 1)

	contextID := "1234"
	// To lock the database we need start a write transaction and then block
	go func() {
		tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
		if err != nil {
			t.Fatalf("Error starting transaction: %v", err)
		}
		queries := manager.queries.WithTx(tx)

		update := fsql.UpdateSessionParams{
			Contextid:         contextID,
			Proto:             []byte{},
			Starttime:         time.Now(),
			Endtime:           time.Now(),
			Selectedid:        "5234",
			Selectedkind:      "kind",
			TotalInputTokens:  100,
			TotalOutputTokens: 100,
			NumGenerateTraces: 0,
		}

		if err := queries.UpdateSession(context.Background(), update); err != nil {
			t.Fatalf("Error writing sessions to the DB: %v", err)
		}

		// Now that we've issued the write the database should be locked
		// Signal main thread that write has started
		blockRead <- true

		// Block until we are told to continue
		t.Log("Blocking until told to continue")
		<-blockUpdate
		t.Log("Continuing transaction")

		if err := tx.Commit(); err != nil {
			t.Logf("Failed to commit transaction: %+v", err)
		}
		//if err := manager.Update(context.Background(), session1.ContextId, func(s *logspb.Session) error {
		//	// Signal main thread that write has started
		//	blockRead <- true
		//	s.FullContext = session1.FullContext
		//	s.LogEvents = session1.LogEvents
		//	s.ContextId = session1.ContextId
		//
		//	// Block until we are told to continue
		//	t.Log("Blocking until told to continue")
		//	<-blockUpdate
		//	t.Log("Continuing transaction")
		//	return nil
		//}); err != nil {
		//	t.Fatalf("Error writing sessions to the DB: %v", err)
		//}
	}()
	// Block until the write transaction has started
	<-blockRead

	// Try to read the session
	if _, err := manager.Get(context.Background(), contextID); err != nil {
		t.Fatalf("Error getting session: %v", err)
	}

	//request := &logspb.DumpExamplesRequest{
	//	Output: filepath.Join(dir.Name, "examples"),
	//}
	//resp, err := manager.DumpExamples(context.Background(), connect.NewRequest(request))
	//
	//if err != nil {
	//	t.Fatalf("Error dumping examples: %v", err)
	//}
	//if resp.Msg.GetNumExamples() != 1 {
	//	t.Fatalf("Expected 1 example but got %v", resp.Msg.GetNumExamples())
	//}
	//t.Logf("Dumped examples to: %v", request.GetOutput())
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

	// Test all fields other than full notebook
	sess2 := &logspb.Session{
		ContextId:         "2",
		TotalOutputTokens: 10,
		TotalInputTokens:  11,
		GenerateTraceIds:  []string{"1", "2"},
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
		{
			name:    "tokens",
			session: sess2,
			expected: &fsql.Session{
				Contextid:         "2",
				Starttime:         sess2.GetStartTime().AsTime(),
				Endtime:           sess2.GetEndTime().AsTime(),
				TotalInputTokens:  11,
				TotalOutputTokens: 10,
				NumGenerateTraces: 2,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := protoToRow(c.session)
			if err != nil {
				t.Fatalf("Error converting session to row: %v", err)
			}
			// Compute the expected serialized proto
			if c.expected.Proto == nil {
				b, err := proto.Marshal(c.session)
				if err != nil {
					t.Fatalf("Error marshalling session: %v", err)
				}
				c.expected.Proto = b
			}
			comparer := cmpopts.IgnoreUnexported(fsql.Session{}, time.Time{})
			if d := cmp.Diff(actual, c.expected, comparer); d != "" {
				t.Fatalf("Unexpected diff between expected and actual session:\n%v", d)
			}
		})
	}
}
