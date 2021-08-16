package v1

import (
	"errors"
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/utils"
)

func (s StandaloneWorkload) Validate() error {
	missingFields := []string{}

	// required fields
	if s.Name == "" {
		missingFields = append(missingFields, "name")
	}

	if s.Spec.Domain == "" {
		missingFields = append(missingFields, "spec.domain")
	}

	if s.Spec.APIGroup == "" {
		missingFields = append(missingFields, "spec.apiGroup")
	}

	if s.Spec.APIVersion == "" {
		missingFields = append(missingFields, "spec.apiVersion")
	}

	if s.Spec.APIKind == "" {
		missingFields = append(missingFields, "spec.apiKind")
	}

	if len(missingFields) > 0 {
		msg := fmt.Sprintf("Missing required fields: %s", missingFields)
		return errors.New(msg)
	}

	return nil
}

func (s StandaloneWorkload) GetWorkloadKind() WorkloadKind {
	return s.Kind
}

// methods that implement WorkloadInitializer.
func (s StandaloneWorkload) GetDomain() string {
	return s.Spec.Domain
}

func (s StandaloneWorkload) HasRootCmdName() bool {
	return s.Spec.CompanionCliRootcmd.Name != ""
}

func (s StandaloneWorkload) HasSubCmdName() bool {
	// standalone workloads never have subcommands
	return false
}

func (s StandaloneWorkload) GetRootCmdName() string {
	return s.Spec.CompanionCliRootcmd.Name
}

func (s StandaloneWorkload) GetRootCmdDescr() string {
	return s.Spec.CompanionCliRootcmd.Description
}

// methods that implement WorkloadAPIBuilder.
func (s StandaloneWorkload) GetName() string {
	return s.Name
}

func (s StandaloneWorkload) GetPackageName() string {
	return s.PackageName
}

func (s StandaloneWorkload) GetAPIGroup() string {
	return s.Spec.APIGroup
}

func (s StandaloneWorkload) GetAPIVersion() string {
	return s.Spec.APIVersion
}

func (s StandaloneWorkload) GetAPIKind() string {
	return s.Spec.APIKind
}

func (s StandaloneWorkload) GetSubcommandName() string {
	// no subcommands for standalone workloads
	return ""
}

func (s StandaloneWorkload) GetSubcommandDescr() string {
	// no subcommands for standalone workloads
	return ""
}

func (s StandaloneWorkload) GetSubcommandVarName() string {
	// no subcommands for standalone workloads
	return ""
}

func (s StandaloneWorkload) GetSubcommandFileName() string {
	// no subcommands for standalone workloads
	return ""
}

func (s StandaloneWorkload) GetRootcommandName() string {
	return s.Spec.CompanionCliRootcmd.Name
}

func (s StandaloneWorkload) IsClusterScoped() bool {
	return s.Spec.ClusterScoped
}

func (s StandaloneWorkload) IsStandalone() bool {
	return true
}

func (s StandaloneWorkload) IsComponent() bool {
	return false
}

func (s StandaloneWorkload) IsCollection() bool {
	return false
}

func (s *StandaloneWorkload) SetSpecFields(workloadPath string) error {
	apiSpecFields, err := processMarkers(workloadPath, s.Spec.Resources, false)
	if err != nil {
		return err
	}

	s.Spec.APISpecFields = *apiSpecFields

	return nil
}

func (s *StandaloneWorkload) SetResources(workloadPath string) error {
	sourceFiles, rbacRules, ownershipRules, err := processResources(workloadPath, s.Spec.Resources)
	if err != nil {
		return err
	}

	s.Spec.SourceFiles = *sourceFiles
	s.Spec.RBACRules = *rbacRules
	s.Spec.OwnershipRules = *ownershipRules

	return nil
}

func (s StandaloneWorkload) GetDependencies() *[]ComponentWorkload {
	return &[]ComponentWorkload{}
}

func (s *StandaloneWorkload) SetComponents(components *[]ComponentWorkload) error {
	return errors.New("Cannot set component workloads on a standalone workload - only on collections")
}

func (s StandaloneWorkload) HasChildResources() bool {
	return len(s.Spec.Resources) > 0
}

func (s StandaloneWorkload) GetComponents() *[]ComponentWorkload {
	return &[]ComponentWorkload{}
}

func (s StandaloneWorkload) GetSourceFiles() *[]SourceFile {
	return &s.Spec.SourceFiles
}

func (s StandaloneWorkload) GetFuncNames() (createFuncNames, initFuncNames []string) {
	return getFuncNames(*s.GetSourceFiles())
}

func (s StandaloneWorkload) GetAPISpecFields() *[]APISpecField {
	return &s.Spec.APISpecFields
}

func (s StandaloneWorkload) GetRBACRules() *[]RBACRule {
	return &s.Spec.RBACRules
}

func (s StandaloneWorkload) GetOwnershipRules() *[]OwnershipRule {
	return &s.Spec.OwnershipRules
}

func (s StandaloneWorkload) GetComponentResource(domain, repo string, clusterScoped bool) *resource.Resource {
	return &resource.Resource{}
}

func (s *StandaloneWorkload) SetNames() {
	s.PackageName = utils.ToPackageName(s.Name)
	if s.HasRootCmdName() {
		s.Spec.CompanionCliRootcmd.VarName = utils.ToVarName(s.Spec.CompanionCliRootcmd.Name)
		s.Spec.CompanionCliRootcmd.FileName = utils.ToFileName(s.Spec.CompanionCliRootcmd.Name)
	}
}
