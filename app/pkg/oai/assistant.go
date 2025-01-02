package oai

import (
	"context"
	"github.com/jlewi/foyle/app/api/oaiapi"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// AssistantController is a controller for OpenAI assistant
type AssistantController struct {
	cfg    config.Config
	client *openai.Client
}

// NewAssistantController creates a new controller for OpenAI assistant
func NewAssistantController(cfg config.Config) (*AssistantController, error) {
	return &AssistantController{cfg: cfg}, nil
}

// ReconcileNode reconciles the state of the resource.
func (a *AssistantController) ReconcileNode(ctx context.Context, node *yaml.RNode) error {
	s := &oaiapi.Assistant{}
	if err := node.YNode().Decode(s); err != nil {
		return errors.Wrap(err, "Failed to decode Assistant")
	}

	return a.Apply(ctx, s)
}

func (a *AssistantController) Apply(ctx context.Context, s *oaiapi.Assistant) error {
	log := logs.FromContext(ctx)
	if a.client == nil {
		client, err := NewClient(a.cfg)
		if err != nil {
			return errors.Wrap(err, "Failed to create OpenAI client")
		}
		a.client = client
	}

	client := a.client
	tools := []openai.AssistantTool{
		{
			Type: openai.AssistantToolTypeFileSearch,
		},
	}
	req := &openai.AssistantRequest{
		Model:        s.Spec.Model,
		Name:         proto.String(s.Metadata.Name),
		Description:  proto.String(s.Spec.Description),
		Instructions: proto.String(s.Spec.Instructions),
		Tools:        tools,
		ToolResources: &openai.AssistantToolResource{
			FileSearch: &openai.AssistantToolFileSearch{VectorStoreIDs: s.Spec.VectorStoreIDs},
		},
	}
	resp, err := client.CreateAssistant(ctx, *req)

	if err != nil {
		return errors.Wrapf(err, "Failed to create assistant %v", s.Metadata.Name)
	}

	log.Info("Created assistant", "name", s.Metadata.Name, "id", resp.ID)
	return nil
}
