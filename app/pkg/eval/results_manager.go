package eval

import (
	"context"
	"database/sql"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

// ResultsManager manages the database containing the evaluation results
type ResultsManager struct {
	//queries *fsql.Queries
	db *sql.DB
}

// EvalResultUpdater is a function that updates an evaluation result.
type EvalResultUpdater func(result *v1alpha1.EvalResult) error

// Update updates an evaluation result. Update performs a read-modify-write operation on the results with the given id.
// The updateFunc is called with the example to be updated. The updateFunc should modify the session in place.
// If the updateFunc returns an error then the example is not updated.
// If the given id doesn't exist then an empty Session is passed to updateFunc and the result will be
// inserted if the updateFunc returns nil. If the session result exists then the result is passed to updateFunc
// and the updated value is then written to the database
func (db *ResultsManager) Update(ctx context.Context, resultID string, updateFunc EvalResultUpdater) error {
	//log := logs.FromContext(ctx)
	//if contextID == "" {
	//	return errors.WithStack(errors.New("contextID must be non-empty"))
	//}
	//log = log.WithValues("contextId", contextID)
	//
	//tx, err := db.db.BeginTx(ctx, &sql.TxOptions{})
	//if err != nil {
	//	return errors.Wrapf(err, "Failed to start transaction")
	//}
	//
	//queries := db.queries.WithTx(tx)
	//// Read the record
	//sessRow, err := queries.GetSession(ctx, contextID)
	//
	//// If the session doesn't exist then we do nothing because session is initializeed to empty session
	//session := &logspb.Session{
	//	ContextId: contextID,
	//}
	//if err != nil {
	//	if err != sql.ErrNoRows {
	//		if txErr := tx.Rollback(); txErr != nil {
	//			log.Error(txErr, "Failed to rollback transaction")
	//		}
	//		return errors.Wrapf(err, "Failed to get session with id %v", contextID)
	//	}
	//} else {
	//	// Deserialize the proto
	//	if err := proto.Unmarshal(sessRow.Proto, session); err != nil {
	//		if txErr := tx.Rollback(); txErr != nil {
	//			log.Error(txErr, "Failed to rollback transaction")
	//		}
	//		return errors.Wrapf(err, "Failed to deserialize session")
	//	}
	//}
	//
	//if err := updateFunc(session); err != nil {
	//	if txErr := tx.Rollback(); txErr != nil {
	//		log.Error(txErr, "Failed to rollback transaction")
	//	}
	//	return errors.Wrapf(err, "Failed to update session")
	//}
	//
	//newRow, err := protoToRow(session)
	//if err != nil {
	//	if txErr := tx.Rollback(); txErr != nil {
	//		log.Error(txErr, "Failed to rollback transaction")
	//	}
	//	return errors.Wrapf(err, "Failed to convert session proto to table row")
	//}
	//
	//if newRow.Contextid != contextID {
	//	if txErr := tx.Rollback(); txErr != nil {
	//		log.Error(txErr, "Failed to rollback transaction")
	//	}
	//	return errors.WithStack(errors.Errorf("contextID in session doesn't match contextID. Update was called with contextID: %v but session has contextID: %v", contextID, newRow.Contextid))
	//}
	//
	//update := fsql.UpdateSessionParams{
	//	Contextid:         contextID,
	//	Proto:             newRow.Proto,
	//	Starttime:         newRow.Starttime,
	//	Endtime:           newRow.Endtime,
	//	Selectedid:        newRow.Selectedid,
	//	Selectedkind:      newRow.Selectedkind,
	//	TotalInputTokens:  newRow.TotalInputTokens,
	//	TotalOutputTokens: newRow.TotalOutputTokens,
	//	NumGenerateTraces: newRow.NumGenerateTraces,
	//}
	//
	//if err := queries.UpdateSession(ctx, update); err != nil {
	//	if txErr := tx.Rollback(); txErr != nil {
	//		log.Error(txErr, "Failed to rollback transaction")
	//	}
	//	return errors.Wrapf(err, "Failed to update session")
	//}
	//
	//if err := tx.Commit(); err != nil {
	//	return errors.Wrapf(err, "Failed to commit transaction")
	//}
	//
	//return nil
}
