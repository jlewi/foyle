package oaiapi

import (
	"github.com/jlewi/foyle/app/api"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	FileSyncGVK = schema.FromAPIVersionAndKind(OAIGroup+"/"+api.Version, "FileSync")
)

// FileSync based off https://platform.openai.com/docs/api-reference/vector-stores/create
type FileSync struct {
	Metadata api.Metadata `json:"metadata" yaml:"metadata"`
	Spec     FileSyncSpec `json:"spec"    yaml:"spec"`
}

type FileSyncSpec struct {
	// Source is the source glob to match
	Source string `json:"source" yaml:"source"`
	// VectorStoreID is the ID of the vector store to sync the files to
	VectorStoreID   string `json:"vectorStoreID" yaml:"vectorStoreID"`
	VectorStoreName string `json:"vectorStoreName" yaml:"vectorStoreName"`
}
