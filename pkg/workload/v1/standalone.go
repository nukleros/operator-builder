package v1

// methods that implement WorkloadInitializer
func (s StandaloneWorkload) GetDomain() string {
	return s.Spec.Domain
}

func (s StandaloneWorkload) HasRootCmdName() bool {

	if s.Spec.CompanionCliRootcmd.Name != "" {
		return true
	} else {
		return false
	}
}

func (s StandaloneWorkload) GetRootCmdName() string {
	return s.Spec.CompanionCliRootcmd.Name
}

func (s StandaloneWorkload) GetRootCmdDescr() string {
	return s.Spec.CompanionCliRootcmd.Description
}

// methods that implement WorkloadAPIBuilder
func (s StandaloneWorkload) GetName() string {
	return s.Name
}

func (s StandaloneWorkload) GetGroup() string {
	return s.Spec.Group
}

func (s StandaloneWorkload) GetVersion() string {
	return s.Spec.Version
}

func (s StandaloneWorkload) GetKind() string {
	return s.Spec.Kind
}

func (s StandaloneWorkload) GetSubcommandName() string {
	// no subcommands for standalone workloads
	return ""
}

func (s StandaloneWorkload) GetSubcommandDescr() string {
	// no subcommands for standalone workloads
	return ""
}

func (s StandaloneWorkload) GetRootcommandName() string {
	return s.Spec.CompanionCliRootcmd.Name
}

func (s StandaloneWorkload) IsClusterScoped() bool {
	if s.Spec.ClusterScoped {
		return true
	} else {
		return false
	}
}

func (s StandaloneWorkload) IsComponent() bool {
	return false
}

func (s StandaloneWorkload) GetSpecFields(workloadPath string) (*[]APISpecField, error) {

	return processMarkers(workloadPath, s.Spec.Resources)
}

func (s StandaloneWorkload) GetResources(workloadPath string) (*[]SourceFile, error) {

	// each sourceFile is a source code file that contains one or more child
	// resource definition
	var sourceFiles []SourceFile

	for _, manifestFile := range s.Spec.Resources {
		sourceFile, err := processResources(manifestFile, workloadPath)
		if err != nil {
			return &[]SourceFile{}, err
		}

		sourceFiles = append(sourceFiles, sourceFile)

	}

	return &sourceFiles, nil
}

func (s StandaloneWorkload) GetDependencies() []string {
	return []string{}
}
