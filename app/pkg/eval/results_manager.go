package eval

import (
	"context"
	"database/sql"
	"github.com/jlewi/foyle/app/pkg/analyze/fsql"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
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

// Update updates an evaluation result. Update performs a read-modify-write operation on the results with the given id.
// The updateFunc is called with the example to be updated. The updateFunc should modify the session in place.
// If the updateFunc returns an error then the example is not updated.
// If the given id doesn't exist then an empty Session is passed to updateFunc and the result will be
// inserted if the updateFunc returns nil. If the session result exists then the result is passed to updateFunc
// and the updated value is then written to the database
func (db *ResultsManager) Update(ctx context.Context, id string, updateFunc EvalResultUpdater) error {
	log := logs.FromContext(ctx)
	if id == "" {
		return errors.WithStack(errors.New("id must be non-empty"))
	}
	log = log.WithValues("exampleId", id)

	tx, err := db.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.Wrapf(err, "Failed to start transaction")
	}

	queries := db.queries.WithTx(tx)
	// Read the record
	row, err := queries.GetResult(ctx, id)

	// If the session doesn't exist then we do nothing because session is initializeed to empty session
	rowPb := &v1alpha1.EvalResult{}
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

func protoToRowUpdate(result *v1alpha1.EvalResult) (*fsql.UpdateResultParams, error) {
	protoJson, err := protojson.Marshal(result)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to serialize EvalResult to JSON")
	}

	return &fsql.UpdateResultParams{
		ID:        result.GetExample().GetId(),
		Time:      result.Example.GetTime().AsTime(),
		ProtoJson: string(protoJson),
	}, nil
}
