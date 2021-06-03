package v1

// WorkloadSharedSpec contains fields shared by all workload specs
type WorkloadSharedSpec struct {
	Group         string `json:"group"`
	Version       string `json:"version"`
	Kind          string `json:"kind"`
	ClusterScoped bool   `json:"clusterScoped"`
}

// WorkloadShared contains fields shared by all workloads
type WorkloadShared struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

// CliCommand defines the command name and description for the root command or
// subcommand of a companion CLI
type CliCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// StandaloneWorkloadSpec defines the attributes for a standalone workload
type StandaloneWorkloadSpec struct {
	WorkloadSharedSpec
	Domain              string     `json:"domain"`
	CompanionCliRootcmd CliCommand `json:"companionCliRootcmd" `
	Resources           []string   `json:"resources"`
}

// StandaloneWorkload defines a standalone workload
type StandaloneWorkload struct {
	WorkloadShared
	Spec StandaloneWorkloadSpec `json:"spec"`
}

// ComponentWorkloadSpec defines the attributes for a workload that is a
// component of a collection
type ComponentWorkloadSpec struct {
	WorkloadSharedSpec
	CompanionCliSubcmd CliCommand `json:"companionCliRootcmd" `
	Resources          []string   `json:"resources"`
}

// ComponentWorkload defines a workload that is a component of a collection
type ComponentWorkload struct {
	WorkloadShared
	Spec ComponentWorkloadSpec `json:"spec"`
}

// WorkloadCollectionSpec defines the attributes for a workload collection
type WorkloadCollectionSpec struct {
	WorkloadSharedSpec
	Domain              string     `json:"domain"`
	CompanionCliRootcmd CliCommand `json:"companionCliRootcmd" `
	Components          []string   `json:"components"`
	Dependencies        []string   `yaml:"dependencies"`
}

// WorkloadCollection defines a workload collection
type WorkloadCollection struct {
	WorkloadShared
	Spec WorkloadCollectionSpec `json:"spec"`
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

// SourceFile represents a golang source code file that contains one or more
// child resource objects
type SourceFile struct {
	Filename string
	Children []ChildResource
	Legacy   bool
}

// ChildResource contains attributes for resources created by the custom resource.
// These definitions are inferred from the resource manfiests.
type ChildResource struct {
	Name                 string
	UniqueName           string
	Group                string
	Version              string
	Kind                 string
	StaticContent        string
	SourceCode           string
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

// Project contains the project config saved to the WORKLOAD file to allow
// access to config values shared across different operator-builder commands
type Project struct {
	CliRootCommandName string `json:"cliRootCommandName"`
}

const ConfigTaxiKey = "configTaxi"

type ConfigTaxi struct {
	StandaloneConfigPath string
	CollectionConfigPath string
	ComponentConfigPath  string
}
