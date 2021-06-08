package v1

// WorkloadInitializer defines the interface that must be implemented by a
// workload being used to configure project initialization
type WorkloadInitializer interface {
	Validate() error
	GetDomain() string
	HasRootCmdName() bool
	GetRootCmdName() string
	GetRootCmdDescr() string
}

// WorkloadAPIBuilder defines the interface that must be implemented by a
// workload being used to configure API and controller creation
type WorkloadAPIBuilder interface {
	Validate() error
	GetName() string
	GetGroup() string
	GetVersion() string
	GetKind() string
	GetSubcommandName() string
	GetSubcommandDescr() string
	GetRootcommandName() string
	IsClusterScoped() bool
	IsComponent() bool
	GetSpecFields(workloadPath string) (*[]APISpecField, error)
	GetResources(workloadPath string) (*[]SourceFile, *[]RBACRule, error)
	GetDependencies() []string
}
