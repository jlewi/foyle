package runme

import (
	"context"
	"github.com/jlewi/foyle/app/pkg/agent"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	aiv1alpha1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/ai/v1alpha1"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
)

// Proxy is a proxy to the agent that converts the agent responses into the runme protocol.
type Proxy struct {
	agent *agent.Agent
}

func (p *Proxy) Generate(ctx context.Context, req *aiv1alpha1.GenerateCellsRequest) (*aiv1alpha1.GenerateCellsResponse, error) {
	span := trace.SpanFromContext(ctx)
	log := logs.FromContext(ctx)
	// We don't update the logger in the context because that will happen in the agent.Generate method and we
	// would end up duplicating the traceId key
	log = log.WithValues("traceId", span.SpanContext().TraceID())

	log.Info("Runme.Generate")

	// Convert the request to the agent format
	doc, err := NotebookToDoc(req.Notebook)
	if err != nil {
		reqJson, err := protojson.Marshal(req)
		if err != nil {
			log.Error(err, "Failed to marshal request")
		}
		log.Error(err, "Failed to convert runme notebook to doc", "request", reqJson)
		return nil, err
	}
	agentReq := &v1alpha1.GenerateRequest{
		Doc: doc,
	}

	// Call the agent
	agentResp, err := p.agent.Generate(ctx, agentReq)
	if err != nil {
		log.Error(err, "Agent.Generate failed")
		return nil, err
	}

	// Convert the agent response to the runme format
	cells, err := BlocksToCells(agentResp.GetBlocks())
	if err != nil {
		log.Error(err, "Failed to convert agent blocks to cells")
		return nil, err
	}
	resp := &aiv1alpha1.GenerateCellsResponse{
		Cells: cells,
	}

	// TODO(jeremy): Set the cell ids.

	return resp, nil
}
