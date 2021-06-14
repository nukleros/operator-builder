package v1

import (
	"errors"
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/utils"
)

func (c ComponentWorkload) Validate() error {

	missingFields := []string{}

	// required fields
	if c.Name == "" {
		missingFields = append(missingFields, "name")
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

func (c ComponentWorkload) GetWorkloadKind() WorkloadKind {
	return c.Kind
}

// methods that implement WorkloadAPIBuilder
func (c ComponentWorkload) GetName() string {
	return c.Name
}

func (c ComponentWorkload) GetPackageName() string {
	return c.PackageName
}

func (c ComponentWorkload) GetAPIGroup() string {
	return c.Spec.APIGroup
}

func (c ComponentWorkload) GetAPIVersion() string {
	return c.Spec.APIVersion
}

func (c ComponentWorkload) GetAPIKind() string {
	return c.Spec.APIKind
}

func (c ComponentWorkload) GetSubcommandName() string {
	return c.Spec.CompanionCliSubcmd.Name
}

func (c ComponentWorkload) GetSubcommandDescr() string {
	return c.Spec.CompanionCliSubcmd.Description
}

func (c ComponentWorkload) GetSubcommandVarName() string {
	return c.Spec.CompanionCliSubcmd.VarName
}

func (c ComponentWorkload) GetSubcommandFileName() string {
	return c.Spec.CompanionCliSubcmd.FileName
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

func (c ComponentWorkload) IsStandalone() bool {
	return false
}

func (c ComponentWorkload) IsComponent() bool {
	return true
}

func (c ComponentWorkload) IsCollection() bool {
	return false
}

func (c *ComponentWorkload) SetSpecFields(workloadPath string) error {

	apiSpecFields, err := processMarkers(workloadPath, c.Spec.Resources)
	if err != nil {
		return err
	}
	c.Spec.APISpecFields = *apiSpecFields

	return nil
}

func (c *ComponentWorkload) SetResources(workloadPath string) error {

	sourceFiles, rbacRules, err := processResources(workloadPath, c.Spec.Resources)
	if err != nil {
		return err
	}
	c.Spec.SourceFiles = *sourceFiles
	c.Spec.RBACRules = *rbacRules

	return nil
}

func (c ComponentWorkload) GetDependencies() []string {
	return []string{}
}

func (c *ComponentWorkload) SetComponents(components *[]ComponentWorkload) error {
	return errors.New("Cannot set component workloads on a component workload - only on collections")
}

func (c ComponentWorkload) HasChildResources() bool {
	if len(c.Spec.Resources) > 0 {
		return true
	}
	return false
}

func (c ComponentWorkload) GetComponents() *[]ComponentWorkload {
	return &[]ComponentWorkload{}
}

func (c ComponentWorkload) GetSourceFiles() *[]SourceFile {
	return &c.Spec.SourceFiles
}

func (c ComponentWorkload) GetAPISpecFields() *[]APISpecField {
	return &c.Spec.APISpecFields
}

func (c ComponentWorkload) GetRBACRules() *[]RBACRule {
	return &c.Spec.RBACRules
}

func (c ComponentWorkload) GetComponentResource(domain, repo string, clusterScoped bool) *resource.Resource {

	var namespaced bool
	if clusterScoped {
		namespaced = false
	} else {
		namespaced = true
	}
	api := resource.API{
		CRDVersion: "v1",
		Namespaced: namespaced,
	}
	return &resource.Resource{
		GVK: resource.GVK{
			Domain:  domain,
			Group:   c.Spec.APIGroup,
			Version: c.Spec.APIVersion,
			Kind:    c.Spec.APIKind,
		},
		Plural: utils.PluralizeKind(c.Spec.APIKind),
		Path: fmt.Sprintf(
			"%s/apis/%s/%s",
			repo,
			c.Spec.APIGroup,
			c.Spec.APIVersion,
		),
		API:        &api,
		Controller: true,
	}
}

func (c ComponentWorkload) HasSubCmdName() bool {
	if c.Spec.CompanionCliSubcmd.Name != "" {
		return true
	} else {
		return false
	}
}

func (c *ComponentWorkload) SetNames() {
	c.PackageName = utils.ToPackageName(c.Name)
	if c.HasSubCmdName() {
		c.Spec.CompanionCliSubcmd.VarName = utils.ToVarName(c.Spec.CompanionCliSubcmd.Name)
		c.Spec.CompanionCliSubcmd.FileName = utils.ToFileName(c.Spec.CompanionCliSubcmd.Name)
	}
}
