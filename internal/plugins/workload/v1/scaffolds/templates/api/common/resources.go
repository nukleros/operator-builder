// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package common

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Resources{}

// Resources scaffolds the common resources for all workloads.
type Resources struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
}

func (f *Resources) SetTemplateDefaults() error {
	f.Path = filepath.Join("apis", "common", "resources.go")

	f.TemplateBody = resourcesTemplate

	return nil
}

const resourcesTemplate = `{{ .Boilerplate }}

package common

// ResourceCommon are the common fields used across multiple resource types.
type ResourceCommon struct {
	// Group defines the API Group of the resource.
	Group string ` + "`" + `json:"group"` + "`" + `

	// Version defines the API Version of the resource.
	Version string ` + "`" + `json:"version"` + "`" + `

	// Kind defines the kind of the resource.
	Kind string ` + "`" + `json:"kind"` + "`" + `

	// Name defines the name of the resource from the metadata.name field.
	Name string ` + "`" + `json:"name"` + "`" + `

	// Namespace defines the namespace in which this resource exists in.
	Namespace string ` + "`" + `json:"namespace"` + "`" + `
}

// Resource is the resource and its condition as stored on the object status field.
type Resource struct {
	ResourceCommon ` + "`" + `json:",omitempty"` + "`" + `

	// ResourceCondition defines the current condition of this resource.
	ResourceCondition ` + "`" + `json:"condition,omitempty"` + "`" + `
}

// GetResourceIndex returns the index of a matching resource.  Any integer which is 0
// or greater indicates that the resource was found.  Anything lower indicates that an
// associated resource is not found.
func (resource *Resource) GetResourceIndex(component Component) int {
	for i, currentResource := range component.GetResources() {
		if currentResource.Group == resource.Group && currentResource.Version == resource.Version && currentResource.Kind == resource.Kind {
			if currentResource.Name == resource.Name && currentResource.Namespace == resource.Namespace {
				return i
			}
		}
	}

	return -1
}
`
