package api

// Experiment is a struct that represents an experiment
type Experiment struct {
	Metadata Metadata       `json:"metadata" yaml:"metadata"`
	Spec     ExperimentSpec `json:"spec"    yaml:"spec"`
}

type ExperimentSpec struct {
	// EvalDir is the directory containing the evaluation the evaluation input
	EvalDir string `json:"evalDir" yaml:"evalDir"`

	// DBDir is the directory for the pebble database that will store the results
	DBDir string `json:"dbDir" yaml:"dbDir"`

	// GoogleSheetID is the ID of the Google Sheet to update with the results.
	GoogleSheetID string `json:"googleSheetID" yaml:"googleSheetID"`

	// SheetName is the name of the sheet to update.
	SheetName string `json:"sheetName" yaml:"sheetName"`

	// Agent is the configuration for the agent
	Agent *AgentConfig `json:"agent,omitempty" yaml:"agent,omitempty"`
}
