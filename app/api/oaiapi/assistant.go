package oaiapi

import (
	"github.com/jlewi/foyle/app/api"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	AssistantGVK = schema.FromAPIVersionAndKind(OAIGroup+"/"+api.Version, "Assistant")
)

// Assistant based off https://platform.openai.com/docs/api-reference/assistants/create
type Assistant struct {
	Metadata api.Metadata  `json:"metadata" yaml:"metadata"`
	Spec     AssistantSpec `json:"spec"    yaml:"spec"`
}

type AssistantSpec struct {
	// Model is the name of the model to use
	Model string `json:"model" yaml:"model"`
	// Instructions is the instructions for the assistant
	Instructions string `json:"instructions" yaml:"instructions"`
	// VectorStoreIDs is the IDs of the vector stores to use
	VectorStoreIDs []string `json:"vectorStoreIDs" yaml:"vectorStoreIDs"`
	Description    string   `json:"description" yaml:"description"`
}
