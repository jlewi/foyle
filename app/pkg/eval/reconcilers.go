package eval

import (
	"context"

	"connectrpc.com/connect"
	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1/v1alpha1connect"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// reconcilePredictions reconciles predictions for examples in the database.
func reconcilePredictions(ctx context.Context, db *pebble.DB, client v1alpha1connect.GenerateServiceClient) error {
	olog := logs.FromContext(ctx)
	iter, err := db.NewIterWithContext(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if key == nil {
			break
		}

		log := olog.WithValues("id", string(key))
		value, err := iter.ValueAndErr()
		if err != nil {
			return errors.Wrapf(err, "Failed to read value for key %s", string(key))
		}

		result := &v1alpha1.EvalResult{}
		if err := proto.Unmarshal(value, result); err != nil {
			return errors.Wrapf(err, "Failed to unmarshal value for key %s", string(key))
		}

		if len(result.GetActual()) > 0 {
			log.V(logs.Debug).Info("not generating a completion; already have answer", "path", result.ExampleFile)
			// We have the answer so we don't need to generate it.
			continue
		}

		if len(result.Actual) == 0 {
			// Initialize a trace
			resp, err := func() (*connect.Response[v1alpha1.GenerateResponse], error) {
				newCtx, span := tracer().Start(ctx, "(*Evaluator).reconcilePredictions")
				defer span.End()

				req := connect.NewRequest(&v1alpha1.GenerateRequest{
					Doc: result.Example.Query,
				})
				// We need to generate the answer.
				return client.Generate(newCtx, req)
			}()

			if err != nil {
				connectErr, ok := err.(*connect.Error)
				if ok {
					// If this is a permanent error we want to abort with an error
					if connectErr.Code() == connect.CodeUnavailable || connectErr.Code() == connect.CodeUnimplemented {
						return errors.Wrap(err, "Unable to connect to the agent.")
					}
				}
				result.Error = err.Error()
				result.Status = v1alpha1.EvalResultStatus_ERROR
				continue
			}

			result.Actual = resp.Msg.GetBlocks()
			result.GenTraceId = resp.Msg.GetTraceId()

			log.Info("Writing result to DB")
			if err := updateResult(ctx, string(key), result, db); err != nil {
				return errors.Wrapf(err, "Failed to write result to DB")
			}
		}
	}
	return nil
}
