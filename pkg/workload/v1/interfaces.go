package v1

type WorkloadInitializer interface {
	GetDomain() string
	HasRootCmdName() bool
	GetRootCmdName() string
	GetRootCmdDescr() string
}

type WorkloadAPIBuilder interface {
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
	GetResources(workloadPath string) (*[]SourceFile, error)
	GetDependencies() []string
}
