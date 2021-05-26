package v1

// CliCommand defines the command name and description for the root command or
// subcommand of a companion CLI
type CliCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// WorkloadSpec defines the desired state for a Workload
type WorkloadSpec struct {
	Group               string     `json:"group"`
	Version             string     `json:"version"`
	Kind                string     `json:"kind"`
	ClusterScoped       bool       `json:"clusterScoped"`
	CompanionCliSubcmd  CliCommand `json:"companionCliSubcmd"`
	Resources           []string   `json:"resources"`
	Collection          bool       `json:"collection"`
	CompanionCliRootcmd CliCommand `json:"companionCliRootcmd" ` // when collection: true
	Children            []string   `json:"children"`             // when collection: true
	Dependencies        []string   `yaml:"dependencies"`
}

// Workload defines the attributes of a distinct workload
// A Workload will get an API type and a controller to manage the Kubernetes
// resourses that constitute that workload
type Workload struct {
	Name string       `json:"name"`
	Spec WorkloadSpec `json:"spec"`
}

// APISpecField represents a single field in a custom API type
type APISpecField struct {
	FieldName         string
	ManifestFieldName string
	DataType          string
	DefaultVal        string
	ZeroVal           string
	ApiSpecContent    string
	SampleField       string
}

// SourceFile is a golang source code file that contains one or more child
// resource objects
type SourceFile struct {
	Filename       string
	ChildResources []ChildResource
	Legacy         bool
}

// ChildResource contains attributes for resources created by the custom resource.
// These definitions are inferred from the resource manfiests.
type ChildResource struct {
	Name                 string
	Group                string
	Version              string
	Kind                 string
	Content              string
	CreateResourceName   string
	LegacyCreateStrategy bool
}

// Marker contains the attributes of a workload marker from a static manifest
type Marker struct {
	Key           string
	Value         string
	FieldName     string
	DataType      string
	Default       string
	LeadingSpaces int
}
