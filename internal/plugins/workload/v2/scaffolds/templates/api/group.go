// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package api

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
)

var _ machinery.Template = &Group{}

// Group scaffolds the file that defines the registration methods for a certain group and version.
type Group struct {
	machinery.TemplateMixin
	machinery.MultiGroupMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin
}

func (f *Group) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"apis",
		f.Resource.Group,
		f.Resource.Version,
		"groupversion_info.go",
	)

	f.TemplateBody = groupTemplate

	return nil
}

const groupTemplate = `{{ .Boilerplate }}

// Package {{ .Resource.Version }} contains API Schema definitions for the {{ .Resource.Group }} {{ .Resource.Version }} API group.
//+kubebuilder:object:generate=true
//+groupName={{ .Resource.QualifiedGroup }}
package {{ .Resource.Version }}

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "{{ .Resource.QualifiedGroup }}", Version: "{{ .Resource.Version }}"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
`
