package api

// N.B. We need to redefine Metadata and not reuse the version in the K8s libraries
// because we want it to have yaml tags so we can serialize with the YAML library.

// Metadata holds an optional name of the project.
type Metadata struct {
	Name        string            `yaml:"name,omitempty"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
	// ResourceVersion is used for optimistic concurrency.
	// Ref: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	// This should be treated as an opaque value by clients.
	ResourceVersion string `yaml:"resourceVersion,omitempty"`
}

// N.B we define our own GVK rather than reusing K8s APIMachinery to minimize dependencies on K8s client libraries.
// Arguably, this is pointless because we are already pulling in K8s client libraries indirectly; probably through
// hydros.

// Gvk holds the group, version, and kind of a resource.
type Gvk struct {
	Group   string `json:"group,omitempty" yaml:"group,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	Kind    string `json:"kind,omitempty" yaml:"kind,omitempty"`
}
