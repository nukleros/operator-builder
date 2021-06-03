package v1

type WorkloadInitializer interface {
	HasRootCmdName() bool
	GetRootCmdName() string
	GetRootCmdDescr() string
}

type WorkloadAPIBuilder interface {
	GetName() string
	GetSubcommandName() string
	GetSubcommandDescr() string
	GetRootcommandName() string
	IsClusterScoped() bool
	IsComponent() bool
	GetSpecFields(workloadPath string) (*[]APISpecField, error)
	GetResources(workloadPath string) (*[]SourceFile, error)
	GetDependencies() []string
}
