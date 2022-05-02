// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package kinds

import (
	"errors"
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/commands/companion"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/manifests"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/markers"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/rbac"
)

var ErrNoComponentsOnComponent = errors.New("cannot set component workloads on a component workload - only on collections")

// ComponentWorkloadSpec defines the attributes for a workload that is a
// component of a collection.
type ComponentWorkloadSpec struct {
	API                   WorkloadAPISpec      `json:"api" yaml:"api"`
	CompanionCliSubcmd    companion.CLI        `json:"companionCliSubcmd" yaml:"companionCliSubcmd" validate:"omitempty"`
	CompanionCliRootcmd   companion.CLI        `json:"-" yaml:"-" validate:"omitempty"`
	Dependencies          []string             `json:"dependencies" yaml:"dependencies"`
	ConfigPath            string               `json:"-" yaml:"-" validate:"omitempty"`
	ComponentDependencies []*ComponentWorkload `json:"-" yaml:"-" validate:"omitempty"`
	WorkloadSpec          `yaml:",inline"`
}

// ComponentWorkload defines a workload that is a component of a collection.
type ComponentWorkload struct {
	WorkloadShared `yaml:",inline"`
	Spec           ComponentWorkloadSpec `json:"spec" yaml:"spec" validate:"required"`
}

// NewComponentWorkload returns a new component workload object.
func NewComponentWorkload(
	name string,
	spec WorkloadAPISpec,
	manifestFiles, dependencies []string,
) *ComponentWorkload {
	return &ComponentWorkload{
		WorkloadShared: WorkloadShared{
			Kind: WorkloadKindComponent,
			Name: name,
		},
		Spec: ComponentWorkloadSpec{
			API: *NewSampleAPISpec(),
			WorkloadSpec: WorkloadSpec{
				Manifests: manifests.FromFiles(manifestFiles),
			},
			Dependencies: dependencies,
		},
	}
}

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
		return fmt.Errorf("%w: %s", ErrMissingRequiredFields, missingFields)
	}

	return nil
}

func (c *ComponentWorkload) GetWorkloadKind() WorkloadKind {
	return c.Kind
}

// methods that implement WorkloadAPIBuilder.
func (c *ComponentWorkload) GetDomain() string {
	return c.Spec.API.Domain
}

func (c *ComponentWorkload) HasRootCmdName() bool {
	return false
}

func (c *ComponentWorkload) GetName() string {
	return c.Name
}

func (c *ComponentWorkload) GetPackageName() string {
	return c.PackageName
}

func (c *ComponentWorkload) GetAPISpec() WorkloadAPISpec {
	return c.Spec.API
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

func (c *ComponentWorkload) SetRBAC() {
	c.Spec.RBACRules.Add(rbac.ForWorkloads(c, c.Spec.Collection))
}

func (c *ComponentWorkload) SetResources(workloadPath string) error {
	err := c.Spec.processManifests(markers.FieldMarkerType)
	if err != nil {
		return err
	}

	return nil
}

func (c *ComponentWorkload) GetDependencies() []*ComponentWorkload {
	return c.Spec.ComponentDependencies
}

func (*ComponentWorkload) SetComponents(components []*ComponentWorkload) error {
	return ErrNoComponentsOnComponent
}

func (c *ComponentWorkload) HasChildResources() bool {
	return len(*c.Spec.Manifests) > 0
}

func (c *ComponentWorkload) GetCollection() *WorkloadCollection {
	return c.Spec.Collection
}

func (*ComponentWorkload) GetComponents() []*ComponentWorkload {
	return []*ComponentWorkload{}
}

func (c *ComponentWorkload) GetAPISpecFields() *APIFields {
	return c.Spec.APISpecFields
}

func (c *ComponentWorkload) GetManifests() *manifests.Manifests {
	return c.Spec.Manifests
}

func (c *ComponentWorkload) GetRBACRules() *[]rbac.Rule {
	var rules []rbac.Rule = *c.Spec.RBACRules

	return &rules
}

func (c *ComponentWorkload) GetComponentResource(domain, repo string, clusterScoped bool) *resource.Resource {
	api := resource.API{
		CRDVersion: "v1",
		Namespaced: !clusterScoped,
	}

	return &resource.Resource{
		GVK: resource.GVK{
			Domain:  domain,
			Group:   c.Spec.API.Group,
			Version: c.Spec.API.Version,
			Kind:    c.Spec.API.Kind,
		},
		Plural: resource.RegularPlural(c.Spec.API.Kind),
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
	return c.Spec.CompanionCliSubcmd.HasName()
}

func (c *ComponentWorkload) HasSubCmdDescription() bool {
	return c.Spec.CompanionCliSubcmd.HasDescription()
}

func (c *ComponentWorkload) SetNames() {
	c.PackageName = utils.ToPackageName(c.Name)

	c.Spec.CompanionCliSubcmd.SetCommonValues(c, true)
}

func (c *ComponentWorkload) GetRootCommand() *companion.CLI {
	return &c.Spec.Collection.Spec.CompanionCliRootcmd
}

func (c *ComponentWorkload) GetSubCommand() *companion.CLI {
	return &c.Spec.CompanionCliSubcmd
}

func (c *ComponentWorkload) LoadManifests(workloadPath string) error {
	expanded, err := manifests.ExpandManifests(workloadPath, c.Spec.Resources)
	if err != nil {
		return fmt.Errorf("%w; %s for component %s", err, ErrLoadManifests, c.Name)
	}

	c.Spec.Manifests = expanded
	for _, manifest := range *c.Spec.Manifests {
		if err := manifest.LoadContent(c.IsCollection()); err != nil {
			return fmt.Errorf("%w; %s for component %s", err, ErrLoadManifests, c.Name)
		}
	}

	return nil
}
