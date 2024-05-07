package eval

import (
	"connectrpc.com/connect"
	"context"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"log"
)

// EvalServer is the server that implements the Eval service interface.
// This is used to make results available to the frontend.
type EvalServer struct{}

func (s *EvalServer) List(
	ctx context.Context,
	req *connect.Request[v1alpha1.EvalResultListRequest],
) (*connect.Response[v1alpha1.EvalResultListResponse], error) {
	log.Println("Request headers: ", req.Header())
	res := connect.NewResponse(&v1alpha1.EvalResultListResponse{
		Items: []*v1alpha1.EvalResult{
			{
				Example: &v1alpha1.Example{
					Id: "hello-world",
				},
			},
		},
	})
	res.Header().Set("Eval-Version", "v1alpha1")
	return res, nil
}
