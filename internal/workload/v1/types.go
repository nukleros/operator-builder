// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

// WorkloadKind indicates which of the supported workload kinds are being used.
type WorkloadKind string

const (
	WorkloadKindStandalone WorkloadKind = "StandaloneWorkload"
	WorkloadKindCollection WorkloadKind = "WorkloadCollection"
	WorkloadKindComponent  WorkloadKind = "ComponentWorkload"
)

// APISpec contains fields shared by all workload specs.
type APISpec struct {
	Domain        string `json:"domain" yaml:"domain"`
	Group         string `json:"group" yaml:"group"`
	Version       string `json:"version" yaml:"version"`
	Kind          string `json:"kind" yaml:"kind"`
	ClusterScoped bool   `json:"clusterScoped" yaml:"clusterScoped"`
}

// WorkloadShared contains fields shared by all workloads.
type WorkloadShared struct {
	Name        string       `json:"name"  yaml:"name" validate:"required"`
	Kind        WorkloadKind `json:"kind"  yaml:"kind" validate:"required"`
	PackageName string
}

// CliCommand defines the command name and description for the root command or
// subcommand of a companion CLI.
type CliCommand struct {
	Name        string
	Description string
	VarName     string
	FileName    string
}

// StandaloneWorkloadSpec defines the attributes for a standalone workload.
type StandaloneWorkloadSpec struct {
	API                 APISpec    `json:"api" yaml:"api"`
	CompanionCliRootcmd CliCommand `json:"companionCliRootcmd" yaml:"companionCliRootcmd" validate:"omitempty"`
	Resources           []string   `json:"resources" yaml:"resources"`
	APISpecFields       []*APISpecField
	SourceFiles         []SourceFile
	RBACRules           []RBACRule
	OwnershipRules      []OwnershipRule
}

// StandaloneWorkload defines a standalone workload.
type StandaloneWorkload struct {
	WorkloadShared `yaml:",inline"`
	Spec           StandaloneWorkloadSpec `json:"spec" yaml:"spec" validate:"required"`
}

// ComponentWorkloadSpec defines the attributes for a workload that is a
// component of a collection.
type ComponentWorkloadSpec struct {
	API                   APISpec    `json:"api" yaml:"api"`
	CompanionCliSubcmd    CliCommand `json:"companionCliSubcmd" yaml:"companionCliSubcmd" validate:"omitempty"`
	Resources             []string   `json:"resources" yaml:"resources"`
	Dependencies          []string   `json:"dependencies" yaml:"dependencies"`
	ConfigPath            string
	ComponentDependencies []*ComponentWorkload
	APISpecFields         []*APISpecField
	SourceFiles           []SourceFile
	RBACRules             []RBACRule
	OwnershipRules        []OwnershipRule
}

// ComponentWorkload defines a workload that is a component of a collection.
type ComponentWorkload struct {
	WorkloadShared `yaml:",inline"`
	Spec           ComponentWorkloadSpec `json:"spec" yaml:"spec" validate:"required"`
}

// WorkloadCollectionSpec defines the attributes for a workload collection.
type WorkloadCollectionSpec struct {
	API                 APISpec    `json:"api" yaml:"api"`
	CompanionCliRootcmd CliCommand `json:"companionCliRootcmd" yaml:"companionCliRootcmd" validate:"omitempty"`
	CompanionCliSubcmd  CliCommand `json:"companionCliSubcmd" yaml:"companionCliSubcmd" validate:"omitempty"`
	Resources           []string   `json:"resources" yaml:"resources"`
	ComponentFiles      []string   `json:"componentFiles" yaml:"componentFiles"`
	Components          []*ComponentWorkload
	APISpecFields       []*APISpecField
	SourceFiles         []SourceFile
	RBACRules           []RBACRule
	OwnershipRules      []OwnershipRule
}

// WorkloadCollection defines a workload collection.
type WorkloadCollection struct {
	WorkloadShared `yaml:",inline"`
	Spec           WorkloadCollectionSpec `json:"spec" yaml:"spec" validate:"required"`
}

// APISpecField represents a single field in a custom API type.
type APISpecField struct {
	FieldName          string
	ManifestFieldName  string
	DataType           string
	DefaultVal         string
	ZeroVal            string
	APISpecContent     string
	SampleField        string
	DocumentationLines []string
}

// SourceFile represents a golang source code file that contains one or more
// child resource objects.
type SourceFile struct {
	Filename  string
	Children  []ChildResource
	HasStatic bool
}

// ChildResource contains attributes for resources created by the custom resource.
// These definitions are inferred from the resource manifests.
type ChildResource struct {
	Name          string
	UniqueName    string
	Group         string
	Version       string
	Kind          string
	StaticContent string
	SourceCode    string
}

// SourceCodeTemplateData is a collection of variables used to generate source code.
type SourceCodeTemplateData struct {
	SpecFields     []*APISpecField
	SourceFiles    *[]SourceFile
	RBACRules      *[]RBACRule
	OwnershipRules *[]OwnershipRule
}

// RBACRule contains the info needed to create the kubebuilder:rbac markers in
// the controller.
type RBACRule struct {
	Group      string
	Resource   string
	Verbs      []string
	VerbString string
}

// OwnershipRule contains the info needed to create the controller ownership
// functionality when setting up the controller with the manager.  This allows
// the controller to reconcile the state of a deleted resource that it manages.
type OwnershipRule struct {
	Version string
	Kind    string
	CoreAPI bool
}

const PluginConfigKey = "operatorBuilder"

// PluginConfig contains the project config values which are stored in the
// PROJECT file under plugins.operatorBuilder.
type PluginConfig struct {
	WorkloadConfigPath string `json:"workloadConfigPath" yaml:"workloadConfigPath"`
	CliRootCommandName string `json:"cliRootCommandName" yaml:"cliRootCommandName"`
}
