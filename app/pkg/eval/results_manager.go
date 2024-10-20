package eval

import (
	"context"
	"database/sql"
	_ "embed"
	"os"
	"path/filepath"
	"time"

	"github.com/jlewi/foyle/app/pkg/analyze"
	"github.com/jlewi/foyle/app/pkg/analyze/fsql"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

// ResultsManager manages the database containing the evaluation results
type ResultsManager struct {
	queries *fsql.Queries
	db      *sql.DB
}

// EvalResultUpdater is a function that updates an evaluation result.
type EvalResultUpdater func(result *v1alpha1.EvalResult) error

func openResultsManager(dbFile string) (*ResultsManager, error) {
	stat, err := os.Stat(dbFile)
	if err == nil && stat.IsDir() {
		return nil, errors.Wrapf(err, "Can't open database: %v; it is a directory", dbFile)
	}
	dbDir := filepath.Dir(dbFile)
	if err := os.MkdirAll(dbDir, helpers.UserGroupAllPerm); err != nil {
		return nil, errors.Wrapf(err, "Failed to create directory: %v", dbDir)
	}

	db, err := sql.Open(analyze.SQLLiteDriver, dbFile)

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open database: %v", dbFile)
	}

	manager, err := NewResultsManager(db)
	if err != nil {
		return nil, err
	}
	return manager, nil
}

func NewResultsManager(db *sql.DB) (*ResultsManager, error) {
	// create tables
	// TODO(jeremy): This creates the analyzer and ResultsManager table because we don't separate the DDL statements.
	// We might want to refactor to support that.
	if _, err := db.ExecContext(context.TODO(), analyze.GetDDL()); err != nil {
		return nil, err
	}

	// Create the dbtx from the actual database
	queries := fsql.New(db)

	return &ResultsManager{
		queries: queries,
		db:      db,
	}, nil
}

// Get retrieves an example with the given id
func (m *ResultsManager) Get(ctx context.Context, id string) (*v1alpha1.EvalResult, error) {
	queries := m.queries

	// Read the record
	row, err := queries.GetResult(ctx, id)

	if err != nil {
		return nil, err
	}

	result := &v1alpha1.EvalResult{}
	if err := protojson.Unmarshal([]byte(row.ProtoJson), result); err != nil {
		return nil, errors.Wrapf(err, "Failed to deserialize EvalResult")
	}

	return result, nil
}

// Update updates an evaluation result. Update performs a read-modify-write operation on the results with the given id.
// The updateFunc is called with the example to be updated. The updateFunc should modify the session in place.
// If the updateFunc returns an error then the example is not updated.
// If the given id doesn't exist then an empty Session is passed to updateFunc and the result will be
// inserted if the updateFunc returns nil. If the session result exists then the result is passed to updateFunc
// and the updated Value is then written to the database
//
// TODO(jeremy): How should the update function signal an error that shouldn't block the update and should be reported
// by Update. For example, when processing a result; we might have an error processing an example (e.g. generating
// a completion). We still want to update the database though and signal to caller of the Update that the error failed.
// Should the EvalResultUpdater return a boolean indicating whether to commit or rollback the transaction?
// Should Update wrap the EvalResultUpdater in a error that stores the error returned by updateFunc?
func (m *ResultsManager) Update(ctx context.Context, id string, updateFunc EvalResultUpdater) error {
	log := logs.FromContext(ctx)
	if id == "" {
		return errors.WithStack(errors.New("id must be non-empty"))
	}
	log = log.WithValues("exampleId", id)

	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.Wrapf(err, "Failed to start transaction")
	}

	queries := m.queries.WithTx(tx)
	// Read the record
	row, err := queries.GetResult(ctx, id)

	// If the session doesn't exist then we do nothing because session is initializeed to empty session
	rowPb := &v1alpha1.EvalResult{
		// Initialize the id.
		Example: &v1alpha1.EvalExample{
			Id: id,
		},
	}
	if err != nil {
		if err != sql.ErrNoRows {
			if txErr := tx.Rollback(); txErr != nil {
				log.Error(txErr, "Failed to rollback transaction")
			}
			return errors.Wrapf(err, "Failed to get result with id %v", id)
		}
	} else {
		// Deserialize the proto
		if err := protojson.Unmarshal([]byte(row.ProtoJson), rowPb); err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				log.Error(txErr, "Failed to rollback transaction")
			}
			return errors.Wrapf(err, "Failed to deserialize result")
		}
	}

	if err := updateFunc(rowPb); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Error(txErr, "Failed to rollback transaction")
		}
		return errors.Wrapf(err, "Failed to update result")
	}

	update, err := protoToRowUpdate(rowPb)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Error(txErr, "Failed to rollback transaction")
		}
		return errors.Wrapf(err, "Failed to convert EvalResult proto to table row")
	}

	if update.ID != id {
		if txErr := tx.Rollback(); txErr != nil {
			log.Error(txErr, "Failed to rollback transaction")
		}
		return errors.WithStack(errors.Errorf("id in EvalResult doesn't match id. Update was called with ID: %v but session has ID: %v", id, update.ID))
	}

	if err := queries.UpdateResult(ctx, *update); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Error(txErr, "Failed to rollback transaction")
		}
		return errors.Wrapf(err, "Failed to update session")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrapf(err, "Failed to commit transaction")
	}

	return nil
}

// ListResults lists the results in the database if cursor is nil then the first page is returned.
// If cursor is non-nil then the next page is returned.
// The cursor is the time.
// Returns empty list of results when no more results.
func (m *ResultsManager) ListResults(ctx context.Context, cursor *time.Time, pageSize int) ([]*v1alpha1.EvalResult, *time.Time, error) {
	params := fsql.ListResultsParams{
		PageSize: int64(pageSize),
	}

	if cursor != nil {
		params.Cursor = *cursor
	} else {
		params.Cursor = ""
	}

	rows, err := m.queries.ListResults(ctx, params)

	if err != nil {
		return nil, nil, errors.Wrapf(err, "Failed to list results")
	}

	results := make([]*v1alpha1.EvalResult, 0)

	// ListResults return nil if there are no results
	if rows == nil {
		return results, nil, nil
	}

	for _, row := range rows {
		result := &v1alpha1.EvalResult{}
		if err := protojson.Unmarshal([]byte(row.ProtoJson), result); err != nil {
			return nil, nil, errors.Wrapf(err, "Failed to deserialize EvalResult")
		}
		results = append(results, result)
	}
	lastTime := &time.Time{}
	*lastTime = rows[len(rows)-1].Time
	return results, lastTime, nil
}

func protoToRowUpdate(result *v1alpha1.EvalResult) (*fsql.UpdateResultParams, error) {
	// Emit default values. Otherwise SQL queries become more complex.
	opts := protojson.MarshalOptions{
		EmitDefaultValues: true,
	}
	protoJson, err := opts.Marshal(result)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to serialize EvalResult to JSON")
	}

	return &fsql.UpdateResultParams{
		ID:        result.GetExample().GetId(),
		Time:      result.Example.GetTime().AsTime(),
		ProtoJson: string(protoJson),
	}, nil
}
