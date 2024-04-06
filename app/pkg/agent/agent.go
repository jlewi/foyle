package agent

import (
	"context"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

// Agent is the agent.
type Agent struct {
	v1alpha1.UnimplementedGenerateServiceServer
}

func (e *Agent) Generate(context.Context, *v1alpha1.GenerateRequest) (*v1alpha1.GenerateResponse, error) {
	resp := &v1alpha1.GenerateResponse{
		Blocks: []*v1alpha1.Block{
			{
				Kind:     v1alpha1.BlockKind_MARKUP,
				Contents: "Hello From The Foyle Server! Your generate request was recieved.",
			},
		},
	}
	return resp, nil
}
