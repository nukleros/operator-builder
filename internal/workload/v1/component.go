// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"errors"
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

func (c *ComponentWorkload) Validate() error {
	missingFields := []string{}

	// required fields
	if c.Name == "" {
		missingFields = append(missingFields, "name")
	}

	if c.Spec.API.Group == "" {
		missingFields = append(missingFields, "spec.api.group")
	}

	if c.Spec.API.Version == "" {
		missingFields = append(missingFields, "spec.api.version")
	}

	if c.Spec.API.Kind == "" {
		missingFields = append(missingFields, "spec.api.kind")
	}

	if len(missingFields) > 0 {
		msg := fmt.Sprintf("Missing required fields: %s", missingFields)

		return errors.New(msg)
	}

	return nil
}

func (c *ComponentWorkload) GetWorkloadKind() WorkloadKind {
	return c.Kind
}

// methods that implement WorkloadAPIBuilder.
func (c *ComponentWorkload) GetName() string {
	return c.Name
}

func (c *ComponentWorkload) GetPackageName() string {
	return c.PackageName
}

func (c *ComponentWorkload) GetAPIGroup() string {
	return c.Spec.API.Group
}

func (c *ComponentWorkload) GetAPIVersion() string {
	return c.Spec.API.Version
}

func (c *ComponentWorkload) GetAPIKind() string {
	return c.Spec.API.Kind
}

func (c *ComponentWorkload) GetSubcommandName() string {
	return c.Spec.CompanionCliSubcmd.Name
}

func (c *ComponentWorkload) GetSubcommandDescr() string {
	return c.Spec.CompanionCliSubcmd.Description
}

func (c *ComponentWorkload) GetSubcommandVarName() string {
	return c.Spec.CompanionCliSubcmd.VarName
}

func (c *ComponentWorkload) GetSubcommandFileName() string {
	return c.Spec.CompanionCliSubcmd.FileName
}

func (*ComponentWorkload) GetRootcommandName() string {
	// no root commands for component workloads
	return ""
}

func (c *ComponentWorkload) IsClusterScoped() bool {
	return c.Spec.API.ClusterScoped
}

func (*ComponentWorkload) IsStandalone() bool {
	return false
}

func (*ComponentWorkload) IsComponent() bool {
	return true
}

func (*ComponentWorkload) IsCollection() bool {
	return false
}

func (c *ComponentWorkload) SetResources(workloadPath string) error {
	resources, err := processMarkers(workloadPath, c.Spec.Resources, false, false)
	if err != nil {
		return err
	}

	c.Spec.APISpecFields = resources.SpecFields
	c.Spec.SourceFiles = *resources.SourceFiles
	c.Spec.RBACRules = *resources.RBACRules
	c.Spec.OwnershipRules = *resources.OwnershipRules

	return nil
}

func (c *ComponentWorkload) GetDependencies() []*ComponentWorkload {
	return c.Spec.ComponentDependencies
}

func (*ComponentWorkload) SetComponents(components []*ComponentWorkload) error {
	return errors.New("cannot set component workloads on a component workload - only on collections")
}

func (c *ComponentWorkload) HasChildResources() bool {
	return len(c.Spec.Resources) > 0
}

func (*ComponentWorkload) GetComponents() []*ComponentWorkload {
	return []*ComponentWorkload{}
}

func (c *ComponentWorkload) GetSourceFiles() *[]SourceFile {
	return &c.Spec.SourceFiles
}

func (c *ComponentWorkload) GetFuncNames() (createFuncNames, initFuncNames []string) {
	return getFuncNames(*c.GetSourceFiles())
}

func (c *ComponentWorkload) GetAPISpecFields() []*APISpecField {
	return c.Spec.APISpecFields
}

func (c *ComponentWorkload) GetRBACRules() *[]RBACRule {
	return &c.Spec.RBACRules
}

func (c *ComponentWorkload) GetOwnershipRules() *[]OwnershipRule {
	return &c.Spec.OwnershipRules
}

func (c *ComponentWorkload) GetComponentResource(domain, repo string, clusterScoped bool) *resource.Resource {
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
			Group:   c.Spec.API.Group,
			Version: c.Spec.API.Version,
			Kind:    c.Spec.API.Kind,
		},
		Plural: utils.PluralizeKind(c.Spec.API.Kind),
		Path: fmt.Sprintf(
			"%s/apis/%s/%s",
			repo,
			c.Spec.API.Group,
			c.Spec.API.Version,
		),
		API:        &api,
		Controller: true,
	}
}

func (c *ComponentWorkload) HasSubCmdName() bool {
	return c.Spec.CompanionCliSubcmd.Name != ""
}

func (c *ComponentWorkload) SetNames() {
	c.PackageName = utils.ToPackageName(c.Name)
	if c.HasSubCmdName() {
		c.Spec.CompanionCliSubcmd.VarName = utils.ToPascalCase(c.Spec.CompanionCliSubcmd.Name)
		c.Spec.CompanionCliSubcmd.FileName = utils.ToFileName(c.Spec.CompanionCliSubcmd.Name)
	}
}

func (c *ComponentWorkload) GetSubcommands() *[]CliCommand {
	commands := []CliCommand{}

	if c.Spec.CompanionCliSubcmd.Name != "" {
		commands = append(commands, c.Spec.CompanionCliSubcmd)
	}
	return &commands
}
