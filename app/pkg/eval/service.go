package eval

import (
	"context"

	"connectrpc.com/connect"
	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// EvalServer is the server that implements the Eval service interface.
// This is used to make results available to the frontend.
type EvalServer struct{}

func (s *EvalServer) List(
	ctx context.Context,
	req *connect.Request[v1alpha1.EvalResultListRequest],
) (*connect.Response[v1alpha1.EvalResultListResponse], error) {
	log := logs.FromContext(ctx)

	if req.Msg.GetDatabase() == "" {
		err := connect.NewError(connect.CodeInvalidArgument, errors.New("Request is missing database"))
		log.Error(err, "Invalid EvalResultListRequest")
		return nil, err
	}

	db, err := pebble.Open(req.Msg.GetDatabase(), &pebble.Options{})
	if err != nil {
		log.Error(err, "Failed to open database")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer helpers.DeferIgnoreError(db.Close)

	iter, err := db.NewIterWithContext(ctx, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer iter.Close()

	results := &v1alpha1.EvalResultListResponse{
		Items: make([]*v1alpha1.EvalResult, 0, 100),
	}

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if key == nil {
			break
		}

		value, err := iter.ValueAndErr()
		if err != nil {
			log.Error(err, "Failed to read value for key", "key", string(key))
			continue
		}

		result := &v1alpha1.EvalResult{}
		if err := proto.Unmarshal(value, result); err != nil {
			log.Error(err, "Failed to unmarshal value for", "key", string(key))
			continue
		}
		results.Items = append(results.Items, result)
	}

	res := connect.NewResponse(results)
	res.Header().Set("Eval-Version", "v1alpha1")
	return res, nil
}
