package executor

import (
	"context"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

// Executor is responsible for executing the commands
type Executor struct {
	v1alpha1.UnimplementedExecuteServiceServer
}

func (e *Executor) Execute(context.Context, *v1alpha1.ExecuteRequest) (*v1alpha1.ExecuteResponse, error) {
	resp := &v1alpha1.ExecuteResponse{
		Outputs: []*v1alpha1.BlockOutput{
			{
				Items: []*v1alpha1.BlockOutputItem{
					{
						TextData: "Hello From The Foyle Server!",
					},
				},
			},
		},
	}
	return resp, nil
}
