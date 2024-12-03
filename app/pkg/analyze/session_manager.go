package analyze

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"connectrpc.com/connect"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/app/pkg/runme/converters"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"

	"github.com/jlewi/foyle/app/pkg/analyze/fsql"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite"
)

//go:embed fsql/schema.sql
var ddl string

const (
	SQLLiteDriver = "sqlite"
)

var (
	sessCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "session_updates",
			Help: "Number of sessions updated",
		},
		[]string{"status"},
	)

	sqlLiteBusyErrs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sqlite_busy",
		Help: "Number of operations that failed because sqlite was busy",
	})
)

// GetDDL return the DDL for the database.
// This is a hack because the DDL statements for the sessions and eval results tables are in the same file and package.
// The Evaluator needs to be able to get the DDL in order to create the eval results table. We should clean this up
func GetDDL() string {
	return ddl
}

// SessionUpdater is a function that updates a session.
type SessionUpdater func(session *logspb.Session) error

// SessionsManager manages the database containing sessions.
type SessionsManager struct {
	queries *fsql.Queries
	db      *sql.DB
}

func NewSessionsManager(db *sql.DB) (*SessionsManager, error) {
	// create tables
	// TODO(jeremy): I think this creates the evalresults table as well because we don't separate the DDL statements.
	if _, err := db.ExecContext(context.TODO(), ddl); err != nil {
		return nil, err
	}

	// Create the dbtx from the actual database
	queries := fsql.New(db)

	return &SessionsManager{
		queries: queries,
		db:      db,
	}, nil
}

// Get retrieves a session with the given contextID.
func (db *SessionsManager) Get(ctx context.Context, contextID string) (*logspb.Session, error) {
	queries := db.queries

	// Read the record
	sessRow, err := queries.GetSession(ctx, contextID)

	if err != nil {
		logDBErrors(ctx, err)
		return nil, err
	}

	session := &logspb.Session{}
	if err := proto.Unmarshal(sessRow.Proto, session); err != nil {
		return nil, errors.Wrapf(err, "Failed to deserialize session")
	}

	return session, nil
}

// Update updates a session. Update performs a read-modify-write operation on the session with the given contextID.
// The updateFunc is called with the session to be updated. The updateFunc should modify the session in place.
// If the updateFunc returns an error then the session is not updated.
// If the given contextID doesn't exist then an empty Session is passed to updateFunc and the session will be
// inserted if the updateFunc returns nil. If the session already exists then the session is passed to updateFunc
// and the updated value is then written to the database
func (db *SessionsManager) Update(ctx context.Context, contextID string, updateFunc SessionUpdater) error {
	log := logs.FromContext(ctx)
	if contextID == "" {
		return errors.WithStack(errors.New("contextID must be non-empty"))
	}
	log = log.WithValues("contextId", contextID)

	sessCounter.WithLabelValues("start").Inc()

	tx, err := db.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		sessCounter.WithLabelValues("failedstart").Inc()
		return errors.Wrapf(err, "Failed to start transaction")
	}

	err = func() error {
		queries := db.queries.WithTx(tx)
		// Read the record
		sessRow, err := queries.GetSession(ctx, contextID)

		// If the session doesn't exist then we do nothing because session is initializeed to empty session
		session := &logspb.Session{
			ContextId: contextID,
		}
		if err != nil {
			logDBErrors(ctx, err)
			if err != sql.ErrNoRows {
				sessCounter.WithLabelValues("failedget").Inc()
				return errors.Wrapf(err, "Failed to get session with id %v", contextID)
			}
			// ErrNoRows means the session doesn't exist so we just continue with the empty session
		} else {
			// Deserialize the proto
			if err := proto.Unmarshal(sessRow.Proto, session); err != nil {
				return errors.Wrapf(err, "Failed to deserialize session")
			}
		}

		sessCounter.WithLabelValues("callupdatefunc").Inc()

		if err := updateFunc(session); err != nil {
			return errors.Wrapf(err, "Failed to update session")
		}

		newRow, err := protoToRow(session)
		if err != nil {
			return errors.Wrapf(err, "Failed to convert session proto to table row")
		}

		if newRow.Contextid != contextID {
			return errors.WithStack(errors.Errorf("contextID in session doesn't match contextID. Update was called with contextID: %v but session has contextID: %v", contextID, newRow.Contextid))
		}

		update := fsql.UpdateSessionParams{
			Contextid:         contextID,
			Proto:             newRow.Proto,
			Starttime:         newRow.Starttime,
			Endtime:           newRow.Endtime,
			Selectedid:        newRow.Selectedid,
			Selectedkind:      newRow.Selectedkind,
			TotalInputTokens:  newRow.TotalInputTokens,
			TotalOutputTokens: newRow.TotalOutputTokens,
			NumGenerateTraces: newRow.NumGenerateTraces,
		}

		sessCounter.WithLabelValues("callupdatesession").Inc()
		if err := queries.UpdateSession(ctx, update); err != nil {
			logDBErrors(ctx, err)
			return errors.Wrapf(err, "Failed to update session")
		}
		return nil
	}()

	if err == nil {
		if err := tx.Commit(); err != nil {
			logDBErrors(ctx, err)
			log.Error(err, "Failed to commit transaction")
			sessCounter.WithLabelValues("commitfail").Inc()
			return errors.Wrapf(err, "Failed to commit transaction")
		}
		sessCounter.WithLabelValues("success").Inc()
	} else {
		logDBErrors(ctx, err)
		sessCounter.WithLabelValues("fail").Inc()
		log.Error(err, "Failed to update session")
		if txErr := tx.Rollback(); txErr != nil {
			log.Error(txErr, "Failed to rollback transaction")
		}
		return err
	}

	sessCounter.WithLabelValues("done").Inc()
	return nil
}

func (m *SessionsManager) GetSession(ctx context.Context, request *connect.Request[logspb.GetSessionRequest]) (*connect.Response[logspb.GetSessionResponse], error) {
	log := logs.FromContext(ctx)

	if request.Msg.GetContextId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("contextId must be non-empty"))
	}

	session, err := m.Get(ctx, request.Msg.GetContextId())
	if err != nil {
		log.Error(err, "Failed to get session", "contextId", request.Msg.GetContextId())
		return nil, connect.NewError(connect.CodeInternal, errors.Wrapf(err, "Failed to get session"))
	}

	return connect.NewResponse(&logspb.GetSessionResponse{
		Session: session,
	}), nil
}

func (m *SessionsManager) ListSessions(ctx context.Context, request *connect.Request[logspb.ListSessionsRequest]) (*connect.Response[logspb.ListSessionsResponse], error) {
	log := logs.FromContext(ctx)
	queries := m.queries
	dbSessions, err := queries.ListSessions(ctx)
	if err != nil {
		log.Error(err, "Failed to list sessions")
		return nil, connect.NewError(connect.CodeInternal, errors.Wrapf(err, "Failed to  list sessions"))
	}

	resp := &logspb.ListSessionsResponse{
		Sessions: make([]*logspb.Session, 0, len(dbSessions)),
	}
	for _, s := range dbSessions {
		sess := &logspb.Session{}
		if err := proto.Unmarshal(s.Proto, sess); err != nil {
			log.Error(err, "Failed to deserialize session", "contextId", s.Contextid)
			continue
		}
		resp.Sessions = append(resp.Sessions, sess)
	}

	return connect.NewResponse(resp), nil
}

func (m *SessionsManager) DumpExamples(ctx context.Context, request *connect.Request[logspb.DumpExamplesRequest]) (*connect.Response[logspb.DumpExamplesResponse], error) {
	log := logs.FromContext(ctx)
	output := request.Msg.Output
	if output == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("output must be non-empty"))
	}

	_, err := os.Stat(output)
	if os.IsNotExist(err) {
		log.Info("Creating directory", "output", output)

		if err := os.MkdirAll(output, 0755); err != nil {
			log.Error(err, "Failed to create directory", "output", output)
			return nil, connect.NewError(connect.CodeInternal, errors.Wrapf(err, "Failed to create directory %v", output))
		}
	} else if err != nil {
		log.Error(err, "Failed to stat directory", "output", output)
		return nil, connect.NewError(connect.CodeInternal, errors.Wrapf(err, "Failed to stat directory %v", output))
	}

	// List all the sessions
	params := fsql.ListSessionsForExamplesParams{
		Cursor:   "",
		PageSize: 100,
	}
	numExamples := 0
	numSessions := 0
	for {
		sessions, err := m.queries.ListSessionsForExamples(ctx, params)
		if err != nil {
			log.Error(err, "Failed to list sessions")
			return nil, connect.NewError(connect.CodeInternal, errors.Wrapf(err, "Failed to list sessions"))
		}
		if len(sessions) == 0 {
			// No more results
			resp := connect.NewResponse(&logspb.DumpExamplesResponse{
				NumExamples: int32(numExamples),
				NumSessions: int32(numSessions),
			})
			return resp, nil

		}

		numSessions += len(sessions)
		for _, s := range sessions {
			session := &logspb.Session{}
			if err := proto.Unmarshal(s.Proto, session); err != nil {
				log.Error(err, "Failed to deserialize session", "contextId", s.Contextid)
				continue
			}
			example, err := getExampleFromSession(session)
			if err != nil {
				log.Error(err, "Failed to get example from session", "contextId", s.Contextid)
				continue
			}
			if example == nil {
				continue
			}

			b, err := proto.Marshal(example)
			if err != nil {
				log.Error(err, "Failed to marshal example", "contextId", s.Contextid)
				continue
			}

			filename := filepath.Join(output, fmt.Sprintf("%v.evalexample.binpb", example.GetId()))

			if err := os.WriteFile(filename, b, 0644); err != nil {
				log.Error(err, "Failed to write example to file", "filename", filename, "exampleId", example.GetId())
			}
			numExamples += 1
		}

		// Update params
		params.Cursor = sessions[len(sessions)-1].Contextid
	}
}

// protoToRow converts from the proto representation of a session to the database row representation.
//
// TODO(jeremy): I think it would be better to make the return type fsql.UpdateSessionParams. Right now the only
// place this function gets called is in the Update method and the returned value is immediately converted to
// fsql.UpdateSessionParams.
func protoToRow(session *logspb.Session) (*fsql.Session, error) {
	log := logs.NewLogger()
	protoBytes, err := proto.Marshal(session)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to serialize session")
	}

	selectedId := ""
	selectedKind := ""
	if session.GetFullContext().GetNotebook() != nil {
		cells := session.GetFullContext().GetNotebook().GetCells()
		if session.GetFullContext().GetSelected() >= int32(len(cells)) {
			log.Error(errors.New("Selected cell index is out of bounds"), "Selected cell index is out of bounds", "contextId", session.GetContextId(), "selected", session.GetFullContext().GetSelected(), "numCells", len(cells))
		} else {
			cell := cells[session.GetFullContext().GetSelected()]
			selectedId = converters.GetCellID(cell)
			selectedKind = cell.Kind.String()
		}
	}

	return &fsql.Session{
		Contextid:         session.ContextId,
		Starttime:         session.StartTime.AsTime(),
		Endtime:           session.EndTime.AsTime(),
		Proto:             protoBytes,
		Selectedid:        selectedId,
		Selectedkind:      selectedKind,
		TotalInputTokens:  int64(session.TotalInputTokens),
		TotalOutputTokens: int64(session.TotalOutputTokens),
		NumGenerateTraces: int64(len(session.GenerateTraceIds)),
	}, nil
}

// getExampleFromSession turns a session into an example used for evaluation
// returns nil for the Example if the Session doesn't contain data suitable for an example
func getExampleFromSession(s *logspb.Session) (*v1alpha1.EvalExample, error) {
	if s.ContextId == "" {
		return nil, nil
	}

	var executeEvent *v1alpha1.LogEvent
	for _, e := range s.LogEvents {
		if e.GetType() == v1alpha1.LogEventType_EXECUTE {
			executeEvent = e
			break
		}
	}

	// Only sessions with execute events are turned into examples
	if executeEvent == nil {
		return nil, nil
	}

	if s.GetFullContext().GetNotebook() == nil {
		return nil, errors.Errorf("Session doesn't contain a notebook for the full context")
	}

	// If its the first cell in the notebook there's no context from which to do the prediction.
	if s.GetFullContext().GetSelected() <= 0 {
		return nil, nil
	}

	// Check that the selected cell in the full context matches the selected cell in the execute event.
	// This as an attempt to catch data integrity / logging issues. LogEvents should include the SelectedIndex.
	if s.GetFullContext().GetSelected() != executeEvent.SelectedIndex {
		return nil, errors.Errorf("Selected cell in full context %v doesn't match selected cell in execute event; %v", s.GetFullContext().GetSelected(), executeEvent.SelectedIndex)
	}

	// Rebuild the context. For a LogExecuteEvent the FullContext will contain the entire notebook
	// and the selectedID and selected cell should be the cell that is being executed. But we only want to use
	// All the cells before the current cell

	newContext := proto.Clone(s.GetFullContext()).(*v1alpha1.FullContext)
	// Only the cells up to the executed cell should be included
	newContext.Notebook.Cells = newContext.Notebook.Cells[:executeEvent.SelectedIndex]
	// Set the selected cell to the last one in the notebook.
	newContext.Selected = int32(len(newContext.Notebook.Cells) - 1)

	// We need to get the actual cell that was executed from the execute event because the context won't be up todate.
	// The executedCell should be the last one in the event. Some additional context might be sent
	executedCell := executeEvent.Cells[len(executeEvent.Cells)-1]
	expectedCells := []*parserv1.Cell{executedCell}

	// Ensure data integrity by checking the
	if converters.GetCellID(executedCell) != executeEvent.SelectedId {
		return nil, errors.Errorf("The execute event is for cell id %s; but the last cell in the event has cell id %s", executeEvent.SelectedId, converters.GetCellID(executedCell))
	}

	example := &v1alpha1.EvalExample{
		Id:            s.ContextId,
		ExpectedCells: expectedCells,
		FullContext:   newContext,
		Time:          s.GetStartTime(),
	}

	return example, nil
}
