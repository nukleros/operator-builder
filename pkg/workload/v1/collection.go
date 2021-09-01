package v1

import (
	"errors"
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/utils"
)

func (c *WorkloadCollection) Validate() error {
	missingFields := []string{}

	// required fields
	if c.Name == "" {
		missingFields = append(missingFields, "name")
	}

	if c.Spec.Domain == "" {
		missingFields = append(missingFields, "spec.domain")
	}

	if c.Spec.APIGroup == "" {
		missingFields = append(missingFields, "spec.apiGroup")
	}

	if c.Spec.APIVersion == "" {
		missingFields = append(missingFields, "spec.apiVersion")
	}

	if c.Spec.APIKind == "" {
		missingFields = append(missingFields, "spec.apiKind")
	}

	if len(missingFields) > 0 {
		msg := fmt.Sprintf("Missing required fields: %s", missingFields)
		return errors.New(msg)
	}

	return nil
}

func (c *WorkloadCollection) GetWorkloadKind() WorkloadKind {
	return c.Kind
}

// methods that implement WorkloadInitializer.
func (c *WorkloadCollection) GetDomain() string {
	return c.Spec.Domain
}

func (c *WorkloadCollection) HasRootCmdName() bool {
	return c.Spec.CompanionCliRootcmd.Name != ""
}

func (*WorkloadCollection) HasSubCmdName() bool {
	// workload collections never have subcommands
	return false
}

func (c *WorkloadCollection) GetRootCmdName() string {
	return c.Spec.CompanionCliRootcmd.Name
}

func (c *WorkloadCollection) GetRootCmdDescr() string {
	return c.Spec.CompanionCliRootcmd.Description
}

// methods that implement WorkloadAPIBuilder.
func (c *WorkloadCollection) GetName() string {
	return c.Name
}

func (c *WorkloadCollection) GetPackageName() string {
	return c.PackageName
}

func (c *WorkloadCollection) GetAPIGroup() string {
	return c.Spec.APIGroup
}

func (c *WorkloadCollection) GetAPIVersion() string {
	return c.Spec.APIVersion
}

func (c *WorkloadCollection) GetAPIKind() string {
	return c.Spec.APIKind
}

func (*WorkloadCollection) GetSubcommandName() string {
	// no subcommands for workload collections
	return ""
}

func (*WorkloadCollection) GetSubcommandDescr() string {
	// no subcommands for workload collections
	return ""
}

func (*WorkloadCollection) GetSubcommandVarName() string {
	// no subcommands for workload collections
	return ""
}

func (*WorkloadCollection) GetSubcommandFileName() string {
	// no subcommands for workload collections
	return ""
}

func (c *WorkloadCollection) GetRootcommandName() string {
	return c.Spec.CompanionCliRootcmd.Name
}

func (c *WorkloadCollection) IsClusterScoped() bool {
	return c.Spec.ClusterScoped
}

func (c *WorkloadCollection) IsStandalone() bool {
	return false
}

func (c *WorkloadCollection) IsComponent() bool {
	return false
}

func (c *WorkloadCollection) IsCollection() bool {
	return true
}

func (c *WorkloadCollection) SetResources(workloadPath string) error {
	var specFields []*APISpecField

	for _, component := range c.Spec.Components {
		componentResources, err := processMarkers(
			component.Spec.ConfigPath,
			component.Spec.Resources,
			true,
		)
		if err != nil {
			return err
		}

		// add to spec fields if not present
		for _, csf := range componentResources.SpecField {
			fieldPresent := false

			for i, sf := range specFields {
				if sf.FieldName == csf.FieldName {
					if len(csf.DocumentationLines) > 0 {
						specFields[i].DocumentationLines = csf.DocumentationLines
					}

					fieldPresent = true
				}
			}

			if !fieldPresent {
				specFields = append(specFields, csf)
			}
		}
	}

	c.Spec.APISpecFields = specFields

	return nil
}

func (c *WorkloadCollection) GetDependencies() []*ComponentWorkload {
	return []*ComponentWorkload{}
}

func (c *WorkloadCollection) SetComponents(components []*ComponentWorkload) error {
	c.Spec.Components = components

	return nil
}

func (*WorkloadCollection) HasChildResources() bool {
	// collection never has child resources, only components
	return false
}

func (c *WorkloadCollection) GetComponents() []*ComponentWorkload {
	return c.Spec.Components
}

func (c *WorkloadCollection) GetSourceFiles() *[]SourceFile {
	return &[]SourceFile{}
}

func (c *WorkloadCollection) GetFuncNames() (createFuncNames, initFuncNames []string) {
	return getFuncNames(*c.GetSourceFiles())
}

func (c *WorkloadCollection) GetAPISpecFields() []*APISpecField {
	return c.Spec.APISpecFields
}

func (*WorkloadCollection) GetRBACRules() *[]RBACRule {
	return &[]RBACRule{}
}

func (*WorkloadCollection) GetOwnershipRules() *[]OwnershipRule {
	return &[]OwnershipRule{}
}

func (*WorkloadCollection) GetComponentResource(domain, repo string, clusterScoped bool) *resource.Resource {
	return &resource.Resource{}
}

func (c *WorkloadCollection) SetNames() {
	c.PackageName = utils.ToPackageName(c.Name)
	if c.HasRootCmdName() {
		c.Spec.CompanionCliRootcmd.VarName = utils.ToPascalCase(c.Spec.CompanionCliRootcmd.Name)
		c.Spec.CompanionCliRootcmd.FileName = utils.ToFileName(c.Spec.CompanionCliRootcmd.Name)
	}
}
