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
	Resources           []string   `json:"resources"`
	CompanionCliRootcmd CliCommand `json:"companionCliRootcmd" ` // when no WorkloadCollection
	CompanionCliSubcmd  CliCommand `json:"companionCliSubcmd"`
}

// Workload defines the attributes of a distinct workload
// A Workload will get an API type and a controller to manage the Kubernetes
// resourses that constitute that workload
type Workload struct {
	Spec WorkloadSpec `json:"spec"`
}

// WorkloadCollectionSpec defines the desired state for a WorkloadCollection
type WorkloadCollectionSpec struct {
	ClusterScoped       bool       `json:"clusterScoped"`
	CompanionCliRootcmd CliCommand `json:"companionCliRootcmd"`
}

// A WorkloadCollection represents a set of Workloads that belong to a broader
// collective and may have dependencies on one another
type WorkloadCollection struct {
	Spec WorkloadCollectionSpec `json:"spec"`
}
