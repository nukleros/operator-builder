// Copyright 2023 Nukleros
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

var ErrNoComponentsOnStandalone = errors.New("cannot set component workloads on a component workload - only on collections")

// StandaloneWorkloadSpec defines the attributes for a standalone workload.
type StandaloneWorkloadSpec struct {
	API                 WorkloadAPISpec `json:"api" yaml:"api"`
	CompanionCliRootcmd companion.CLI   `json:"companionCliRootcmd" yaml:"companionCliRootcmd" validate:"omitempty"`
	WorkloadSpec        `yaml:",inline"`
}

// StandaloneWorkload defines a standalone workload.
type StandaloneWorkload struct {
	WorkloadShared `yaml:",inline"`
	Spec           StandaloneWorkloadSpec `json:"spec" yaml:"spec" validate:"required"`
}

// NewStandaloneWorkload returns a new standalone workload object.
func NewStandaloneWorkload(
	name string,
	spec WorkloadAPISpec,
	manifestFiles []string,
) *StandaloneWorkload {
	return &StandaloneWorkload{
		WorkloadShared: WorkloadShared{
			Kind: WorkloadKindStandalone,
			Name: name,
		},
		Spec: StandaloneWorkloadSpec{
			API: spec,
			WorkloadSpec: WorkloadSpec{
				Manifests: manifests.FromFiles(manifestFiles),
			},
		},
	}
}

func (s *StandaloneWorkload) Validate() error {
	missingFields := []string{}

	// required fields
	if s.Name == "" {
		missingFields = append(missingFields, "name")
	}

	if s.Spec.API.Domain == "" {
		missingFields = append(missingFields, "spec.domain")
	}

	if s.Spec.API.Group == "" {
		missingFields = append(missingFields, "spec.api.group")
	}

	if s.Spec.API.Version == "" {
		missingFields = append(missingFields, "spec.api.version")
	}

	if s.Spec.API.Kind == "" {
		missingFields = append(missingFields, "spec.api.kind")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("%w: %s", ErrMissingRequiredFields, missingFields)
	}

	return nil
}

func (s *StandaloneWorkload) GetWorkloadKind() WorkloadKind {
	return s.Kind
}

// methods that implement WorkloadInitializer.
func (s *StandaloneWorkload) GetDomain() string {
	return s.Spec.API.Domain
}

func (s *StandaloneWorkload) HasRootCmdName() bool {
	return s.Spec.CompanionCliRootcmd.HasName()
}

func (s *StandaloneWorkload) HasRootCmdDescription() bool {
	return s.Spec.CompanionCliRootcmd.HasDescription()
}

func (*StandaloneWorkload) HasSubCmdName() bool {
	// standalone workloads never have subcommands
	return false
}

// methods that implement WorkloadAPIBuilder.
func (s *StandaloneWorkload) GetName() string {
	return s.Name
}

func (s *StandaloneWorkload) GetPackageName() string {
	return s.PackageName
}

func (s *StandaloneWorkload) GetAPISpec() WorkloadAPISpec {
	return s.Spec.API
}

func (s *StandaloneWorkload) GetAPIGroup() string {
	return s.Spec.API.Group
}

func (s *StandaloneWorkload) GetAPIVersion() string {
	return s.Spec.API.Version
}

func (s *StandaloneWorkload) GetAPIKind() string {
	return s.Spec.API.Kind
}

func (s *StandaloneWorkload) IsClusterScoped() bool {
	return s.Spec.API.ClusterScoped
}

func (*StandaloneWorkload) IsStandalone() bool {
	return true
}

func (*StandaloneWorkload) IsComponent() bool {
	return false
}

func (*StandaloneWorkload) IsCollection() bool {
	return false
}

func (s *StandaloneWorkload) SetRBAC() {
	s.Spec.RBACRules.Add(rbac.ForWorkloads(s))
}

func (s *StandaloneWorkload) SetResources(workloadPath string) error {
	err := s.Spec.processManifests(markers.FieldMarkerType)
	if err != nil {
		return err
	}

	return nil
}

func (*StandaloneWorkload) GetDependencies() []*ComponentWorkload {
	return []*ComponentWorkload{}
}

func (*StandaloneWorkload) SetComponents(components []*ComponentWorkload) error {
	return ErrNoComponentsOnStandalone
}

func (s *StandaloneWorkload) HasChildResources() bool {
	return len(*s.Spec.Manifests) > 0
}

func (s *StandaloneWorkload) GetCollection() *WorkloadCollection {
	// no collection for standalone workloads
	return nil
}

func (s *StandaloneWorkload) GetComponents() []*ComponentWorkload {
	return []*ComponentWorkload{}
}

func (s *StandaloneWorkload) GetAPISpecFields() *APIFields {
	return s.Spec.APISpecFields
}

func (s *StandaloneWorkload) GetManifests() *manifests.Manifests {
	return s.Spec.Manifests
}

func (s *StandaloneWorkload) GetRBACRules() *[]rbac.Rule {
	var rules []rbac.Rule = *s.Spec.RBACRules

	return &rules
}

func (*StandaloneWorkload) GetComponentResource(domain, repo string, clusterScoped bool) *resource.Resource {
	return &resource.Resource{}
}

func (s *StandaloneWorkload) SetNames() {
	s.PackageName = utils.ToPackageName(s.Name)

	// only set the names if we have specified the root command name else none
	// of the following values will matter as the code for the cli will not be
	// generated
	if s.HasRootCmdName() {
		// set the root command values
		s.Spec.CompanionCliRootcmd.SetCommonValues(s, false)
	}
}

func (s *StandaloneWorkload) GetRootCommand() *companion.CLI {
	return &s.Spec.CompanionCliRootcmd
}

func (s *StandaloneWorkload) GetSubCommand() *companion.CLI {
	// no subcommands for a standalone workload
	return &companion.CLI{}
}

func (s *StandaloneWorkload) LoadManifests(workloadPath string) error {
	expanded, err := manifests.ExpandManifests(workloadPath, s.Spec.Resources)
	if err != nil {
		return fmt.Errorf("%w; %s for standalone workload %s", err, ErrLoadManifests.Error(), s.Name)
	}

	s.Spec.Manifests = expanded
	for _, manifest := range *s.Spec.Manifests {
		if err := manifest.LoadContent(s.IsCollection()); err != nil {
			return fmt.Errorf("%w; %s for standalone workload %s", err, ErrLoadManifests.Error(), s.Name)
		}
	}

	return nil
}
