package oaiapi

import (
	"github.com/jlewi/foyle/app/api"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	VectorStoreGVK = schema.FromAPIVersionAndKind(OAIGroup+"/"+api.Version, "VectorStore")
)

// VectorStore based off https://platform.openai.com/docs/api-reference/vector-stores/create
type VectorStore struct {
	Metadata api.Metadata    `json:"metadata" yaml:"metadata"`
	Spec     VectorStoreSpec `json:"spec"    yaml:"spec"`
}

type VectorStoreSpec struct {
	// TODO(jeremy): Should add actual fields.
}
