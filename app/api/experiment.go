package api

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	ExperimentGVK = schema.FromAPIVersionAndKind(Group+"/"+Version, "Experiment")
)

// Experiment is a struct that represents an experiment
type Experiment struct {
	Metadata Metadata       `json:"metadata" yaml:"metadata"`
	Spec     ExperimentSpec `json:"spec"    yaml:"spec"`
}

type ExperimentSpec struct {
	// AgentAddress is the address of the agent to use to generate completions
	AgentAddress string `json:"agentAddress" yaml:"agentAddress"`

	// EvalDir is the directory containing the evaluation examples.
	// These should be EvalExample protos.
	EvalDir string `json:"evalDir" yaml:"evalDir"`

	// OutputDB is the path to the file to store the results in.
	OutputDB string `json:"outputDB" yaml:"outputDB"`
}
