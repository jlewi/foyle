package oai

import (
	"context"
	"github.com/jlewi/foyle/app/api/oaiapi"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// VSController is a controller for OpenAI vector store
type VSController struct {
	cfg    config.Config
	client *openai.Client
}

// NewVSController creates a new controller for OpenAI vector store
func NewVSController(cfg config.Config) (*VSController, error) {
	// Delay creation of the client so that we don't create it just to register the controller
	return &VSController{cfg: cfg}, nil
}

// ReconcileNode reconciles the state of the resource.
func (v *VSController) ReconcileNode(ctx context.Context, node *yaml.RNode) error {
	s := &oaiapi.VectorStore{}
	if err := node.YNode().Decode(s); err != nil {
		return errors.Wrap(err, "Failed to decode VectorStore")
	}

	return v.Apply(ctx, s)
}

func (v *VSController) Apply(ctx context.Context, s *oaiapi.VectorStore) error {
	log := logs.FromContext(ctx)
	
	if v.client == nil {
		client, err := NewClient(v.cfg)
		if err != nil {
			return errors.Wrap(err, "Failed to create OpenAI client")
		}
		v.client = client
	}

	request := &openai.VectorStoreRequest{
		Name: s.Metadata.Name,
	}
	response, err := v.client.CreateVectorStore(ctx, *request)
	if err != nil {
		return errors.Wrapf(err, "Failed to create vector store %v", s.Metadata.Name)
	}
	log.Info("Created vector store", "name", s.Metadata.Name, "id", response.ID)
	return nil
}
