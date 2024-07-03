// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package kinds

import (
	"errors"
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/nukleros/operator-builder/internal/utils"
	"github.com/nukleros/operator-builder/internal/workload/v1/commands/companion"
	"github.com/nukleros/operator-builder/internal/workload/v1/manifests"
	"github.com/nukleros/operator-builder/internal/workload/v1/markers"
	"github.com/nukleros/operator-builder/internal/workload/v1/rbac"
)

var ErrMissingRequiredFields = errors.New("missing required fields")

// WorkloadCollectionSpec defines the attributes for a workload collection.
type WorkloadCollectionSpec struct {
	API                 WorkloadAPISpec      `json:"api" yaml:"api"`
	CompanionCliRootcmd companion.CLI        `json:"companionCliRootcmd,omitempty" yaml:"companionCliRootcmd,omitempty" validate:"omitempty"`
	CompanionCliSubcmd  companion.CLI        `json:"companionCliSubcmd,omitempty" yaml:"companionCliSubcmd,omitempty" validate:"omitempty"`
	ComponentFiles      []string             `json:"componentFiles" yaml:"componentFiles"`
	Components          []*ComponentWorkload `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
	WorkloadSpec        `yaml:",inline"`
}

// WorkloadCollection defines a workload collection.
type WorkloadCollection struct {
	WorkloadShared `yaml:",inline"`
	Spec           WorkloadCollectionSpec `json:"spec" yaml:"spec" validate:"required"`
}

// NewWorkloadCollection returns a new workload collection object.
func NewWorkloadCollection(
	name string,
	spec WorkloadAPISpec,
	componentFiles []string,
) *WorkloadCollection {
	return &WorkloadCollection{
		WorkloadShared: WorkloadShared{
			Kind: WorkloadKindCollection,
			Name: name,
		},
		Spec: WorkloadCollectionSpec{
			API:            spec,
			ComponentFiles: componentFiles,
		},
	}
}

func (c *WorkloadCollection) Validate() error {
	missingFields := []string{}

	// required fields
	if c.Name == "" {
		missingFields = append(missingFields, "name")
	}

	if c.Spec.API.Domain == "" {
		missingFields = append(missingFields, "spec.api.domain")
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

func (c *WorkloadCollection) GetWorkloadKind() WorkloadKind {
	return c.Kind
}

// methods that implement WorkloadInitializer.
func (c *WorkloadCollection) GetDomain() string {
	return c.Spec.API.Domain
}

func (c *WorkloadCollection) HasRootCmdName() bool {
	return c.Spec.CompanionCliRootcmd.HasName()
}

func (c *WorkloadCollection) HasRootCmdDescription() bool {
	return c.Spec.CompanionCliRootcmd.HasDescription()
}

func (c *WorkloadCollection) HasSubCmdName() bool {
	return c.Spec.CompanionCliSubcmd.HasName()
}

func (c *WorkloadCollection) HasSubCmdDescription() bool {
	return c.Spec.CompanionCliSubcmd.HasDescription()
}

// methods that implement WorkloadAPIBuilder.
func (c *WorkloadCollection) GetName() string {
	return c.Name
}

func (c *WorkloadCollection) GetPackageName() string {
	return c.PackageName
}

func (c *WorkloadCollection) GetAPISpec() WorkloadAPISpec {
	return c.Spec.API
}

func (c *WorkloadCollection) GetAPIGroup() string {
	return c.Spec.API.Group
}

func (c *WorkloadCollection) GetAPIVersion() string {
	return c.Spec.API.Version
}

func (c *WorkloadCollection) GetAPIKind() string {
	return c.Spec.API.Kind
}

func (c *WorkloadCollection) IsClusterScoped() bool {
	return c.Spec.API.ClusterScoped
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

func (c *WorkloadCollection) SetRBAC() {
	c.Spec.RBACRules.Add(rbac.ForWorkloads(c))
}

func (c *WorkloadCollection) SetResources(workloadPath string) error {
	err := c.Spec.processManifests(markers.FieldMarkerType, markers.CollectionMarkerType)
	if err != nil {
		return err
	}

	for _, cpt := range c.Spec.Components {
		for _, csr := range *cpt.Spec.Manifests {
			// add to spec fields if not present
			err := c.Spec.processMarkers(csr, markers.CollectionMarkerType)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *WorkloadCollection) GetDependencies() []*ComponentWorkload {
	return []*ComponentWorkload{}
}

func (c *WorkloadCollection) SetComponents(components []*ComponentWorkload) error {
	c.Spec.Components = components

	return nil
}

func (c *WorkloadCollection) HasChildResources() bool {
	return len(*c.Spec.Manifests) > 0
}

func (c *WorkloadCollection) GetCollection() *WorkloadCollection {
	return c.Spec.Collection
}

func (c *WorkloadCollection) GetComponents() []*ComponentWorkload {
	return c.Spec.Components
}

func (c *WorkloadCollection) GetAPISpecFields() *APIFields {
	return c.Spec.APISpecFields
}

func (c *WorkloadCollection) GetManifests() *manifests.Manifests {
	return c.Spec.Manifests
}

func (c *WorkloadCollection) GetRBACRules() *[]rbac.Rule {
	var rules []rbac.Rule = *c.Spec.RBACRules

	return &rules
}

func (c *WorkloadCollection) GetComponentResource(domain, repo string, clusterScoped bool) *resource.Resource {
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

func (c *WorkloadCollection) SetNames() {
	c.PackageName = utils.ToPackageName(c.Name)

	// only set the names if we have specified the root command name else none
	// of the following values will matter as the code for the cli will not be
	// generated
	if c.HasRootCmdName() {
		// set the root command values
		c.Spec.CompanionCliRootcmd.SetCommonValues(c, false)

		// set the subcommand values
		c.Spec.CompanionCliSubcmd.SetCommonValues(c, true)
	}
}

func (c *WorkloadCollection) GetRootCommand() *companion.CLI {
	return &c.Spec.CompanionCliRootcmd
}

func (c *WorkloadCollection) GetSubCommand() *companion.CLI {
	return &c.Spec.CompanionCliSubcmd
}

func (c *WorkloadCollection) LoadManifests(workloadPath string) error {
	expanded, err := manifests.ExpandManifests(workloadPath, c.Spec.Resources)
	if err != nil {
		return fmt.Errorf("%w; %s for collection %s", err, ErrLoadManifests.Error(), c.Name)
	}

	c.Spec.Manifests = expanded
	for _, manifest := range *c.Spec.Manifests {
		if err := manifest.LoadContent(c.IsCollection()); err != nil {
			return fmt.Errorf("%w; %s for collection %s", err, ErrLoadManifests.Error(), c.Name)
		}
	}

	return nil
}
