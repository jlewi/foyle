package analyze

import (
	"context"
	"database/sql"
	_ "embed"
	"github.com/jlewi/foyle/app/pkg/analyze/fsql"
	"github.com/jlewi/foyle/app/pkg/config"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite"
)

//go:embed fsql/schema.sql
var ddl string

// SessionUpdater is a function that updates a session.
type SessionUpdater func(session *logspb.Session) error

// SessionsManager manages the database containing sessions.
type SessionsManager struct {
	queries *fsql.Queries
	db      *sql.DB
}

func NewSessionsManager(cfg config.Config) (*SessionsManager, error) {
	db, err := sql.Open("sqlite", cfg.GetSessionsDB())

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open database: %v", cfg.GetSessionsDB())
	}

	// create tables
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
	if contextID == "" {
		return errors.WithStack(errors.New("contextID must be non-empty"))
	}

	tx, err := db.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.Wrapf(err, "Failed to start transaction")
	}

	queries := db.queries.WithTx(tx)
	// Read the record
	sessRow, err := queries.GetSession(ctx, contextID)

	var session *logspb.Session
	if err != nil {
		if err != sql.ErrNoRows {
			tx.Rollback()
			return errors.Wrapf(err, "Failed to get session with id %v", contextID)
		}
		// If the session doesn't exist initialize to empty session
		session = &logspb.Session{
			ContextId: contextID,
		}
	} else {
		// Deserialize the proto
		if err := proto.Unmarshal(sessRow.Proto, session); err != nil {
			tx.Rollback()
			return errors.Wrapf(err, "Failed to deserialize session")
		}
	}

	if err := updateFunc(session); err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "Failed to update session")
	}

	newRow, err := protoToRow(session)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "Failed to convert session proto to table row")
	}

	if newRow.Contextid != contextID {
		tx.Rollback()
		return errors.WithStack(errors.Errorf("contextID in session doesn't match contextID. Update was called with contextID: %v but session has contextID: %v", contextID, newRow.Contextid))
	}

	update := fsql.UpdateSessionParams{
		Contextid: contextID,
		Proto:     newRow.Proto,
		Starttime: newRow.Starttime,
		Endtime:   newRow.Endtime,
	}

	if err := queries.UpdateSession(ctx, update); err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "Failed to update session")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrapf(err, "Failed to commit transaction")
	}

	return nil
}

func (db *SessionsManager) Close() error {
	return db.Close()
}

// protoToRow converts from the proto representation of a session to the database row representation.
func protoToRow(session *logspb.Session) (*fsql.Session, error) {
	protoBytes, err := proto.Marshal(session)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to serialize session")
	}

	// TODO: How do we deal with the end/starttime? In sqlc should we specify the type as timestamp?
	return &fsql.Session{
		Contextid: session.ContextId,
		Starttime: session.StartTime.AsTime(),
		Endtime:   session.EndTime.AsTime(),
		Proto:     protoBytes,
	}, nil
}
