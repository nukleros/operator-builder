// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/rbac"
)

// WorkloadIdentifier defines an interface for identifying any workload.
type WorkloadIdentifier interface {
	GetName() string
	GetWorkloadKind() WorkloadKind
}

// WorkloadInitializer defines the interface that must be implemented by a
// workload being used to configure project initialization.
type WorkloadInitializer interface {
	Validate() error

	HasRootCmdName() bool

	GetDomain() string
	GetRootCommand() *CliCommand
	GetWorkloadKind() WorkloadKind

	SetNames()

	IsCollection() bool
}

// WorkloadAPIBuilder defines the interface that must be implemented by a
// workload being used to configure API and controller creation.
type WorkloadAPIBuilder interface {
	Validate() error

	IsClusterScoped() bool
	IsStandalone() bool
	IsComponent() bool
	IsCollection() bool

	HasSubCmdName() bool
	HasChildResources() bool

	GetName() string
	GetPackageName() string
	GetDomain() string
	GetAPIGroup() string
	GetAPIVersion() string
	GetAPIKind() string
	GetDependencies() []*ComponentWorkload
	GetCollection() *WorkloadCollection
	GetComponents() []*ComponentWorkload
	GetSourceFiles() *[]SourceFile
	GetAPISpecFields() *APIFields
	GetRBACRules() *[]rbac.Rule
	GetComponentResource(domain, repo string, clusterScoped bool) *resource.Resource
	GetFuncNames() (createFuncNames, initFuncNames []string)
	GetRootCommand() *CliCommand
	GetSubCommand() *CliCommand

	SetNames()
	SetRBAC()
	SetResources(workloadPath string) error
	SetComponents(components []*ComponentWorkload) error

	LoadManifests(workloadPath string) error
}
