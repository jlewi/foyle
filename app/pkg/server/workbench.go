package server

// IWorkbenchConstructionOptions contains the configuration for the vscode workbench
// Keep it in sync with IWorkbenchConstructionOptions
// https://github.com/microsoft/vscode/blob/d5d14242969257ffff1815ef3bec45d1f2eb0e81/src/vs/workbench/browser/web.api.ts#L134
//
// This struct contains the options passed to the workbench.html template.
type IWorkbenchConstructionOptions struct {
	AdditionalBuiltinExtensions []VSCodeUriComponents `json:"additionalBuiltinExtensions,omitempty"`
}

type VSCodeUriComponents struct {
	Scheme    string `json:"scheme,omitempty"`
	Authority string `json:"authority,omitempty"`
	Path      string `json:"path,omitempty"`
	Query     string `json:"query,omitempty"`
	Fragment  string `json:"fragment,omitempty"`
}
