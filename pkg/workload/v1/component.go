package v1

import (
	"errors"
	"fmt"
)

func (c ComponentWorkload) Validate() error {

	missingFields := []string{}

	// required fields
	if c.Name == "" {
		missingFields = append(missingFields, "name")
	}
	if c.Spec.Group == "" {
		missingFields = append(missingFields, "spec.group")
	}
	if c.Spec.Version == "" {
		missingFields = append(missingFields, "spec.version")
	}
	if c.Spec.Kind == "" {
		missingFields = append(missingFields, "spec.kind")
	}
	if len(missingFields) > 0 {
		msg := fmt.Sprintf("Missing required fields: %s", missingFields)
		return errors.New(msg)
	}

	return nil
}

func (c ComponentWorkload) GetName() string {
	return c.Name
}

func (c ComponentWorkload) GetGroup() string {
	return c.Spec.Group
}

func (c ComponentWorkload) GetVersion() string {
	return c.Spec.Version
}

func (c ComponentWorkload) GetKind() string {
	return c.Spec.Kind
}

func (c ComponentWorkload) GetSubcommandName() string {
	return c.Spec.CompanionCliSubcmd.Name
}

func (c ComponentWorkload) GetSubcommandDescr() string {
	return c.Spec.CompanionCliSubcmd.Description
}

func (c ComponentWorkload) GetRootcommandName() string {
	// no root commands for component workloads
	return ""
}

func (c ComponentWorkload) IsClusterScoped() bool {
	if c.Spec.ClusterScoped {
		return true
	} else {
		return false
	}
}

func (c ComponentWorkload) IsComponent() bool {
	return true
}

func (c ComponentWorkload) GetSpecFields(workloadPath string) (*[]APISpecField, error) {

	return processMarkers(workloadPath, c.Spec.Resources)

}

func (c ComponentWorkload) GetResources(workloadPath string) (*[]SourceFile, error) {

	// each sourceFile is a source code file that contains one or more child
	// resource definition
	var sourceFiles []SourceFile

	for _, manifestFile := range c.Spec.Resources {
		sourceFile, err := processResources(manifestFile, workloadPath)
		if err != nil {
			return nil, err
		}

		sourceFiles = append(sourceFiles, sourceFile)
	}

	return &sourceFiles, nil
}

func (c ComponentWorkload) GetDependencies() []string {
	return []string{}
}
