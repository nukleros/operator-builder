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
}

// Workload defines the attributes of a distinct workload
// A Workload will get an API type and a controller to manage the Kubernetes
// resourses that constitute that workload
type Workload struct {
	Name string       `json:"name"`
	Spec WorkloadSpec `json:"spec"`
}
