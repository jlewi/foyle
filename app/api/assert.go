package api

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	AssertJobGVK = schema.FromAPIVersionAndKind(Group+"/"+Version, "AssertJob")
)

// AssertJob is a struct that represents an assert job. This is a job that runs level one evaluations.
type AssertJob struct {
	Metadata Metadata      `json:"metadata" yaml:"metadata"`
	Spec     AssertJobSpec `json:"spec"    yaml:"spec"`
}

type AssertJobSpec struct {
	// Sources is a list of sources to get the data from
	Sources []EvalSource `json:"sources" yaml:"sources"`

	// AgentAddress is the address of the agent to use to generate completions
	AgentAddress string `json:"agentAddress" yaml:"agentAddress"`

	// DBDir is the directory for the pebble database that will store the results
	DBDir string `json:"dbDir" yaml:"dbDir"`

	// SheetID is the ID of the Google Sheet to update with the results.
	SheetID string `json:"sheetID" yaml:"sheetID"`

	// SheetName is the name of the sheet to update.
	SheetName string `json:"sheetName" yaml:"sheetName"`
}

type EvalSource {
	MarkdownSource *MarkdownSource `json:"markdownSource,omitempty" yaml:"markdownSource,omitempty"`
}

type MarkdownSource struct {
	// Path to the markdown files to use as evaluation data.
	Path string `json:"path" yaml:"path"`
}