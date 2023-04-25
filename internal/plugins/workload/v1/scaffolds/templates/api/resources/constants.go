// Copyright 2023 Nukleros
// SPDX-License-Identifier: MIT

package resources

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

var _ machinery.Template = &Constants{}

// Types scaffolds the child resource Constants files.
type Constants struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	// input fields
	Builder kinds.WorkloadBuilder

	// template fields
	ConstantStrings []string
}

func (f *Constants) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"apis",
		f.Resource.Group,
		f.Resource.Version,
		f.Builder.GetPackageName(),
		"constants",
		"names.go",
	)

	children := kinds.GetWorkloadChildren(f.Builder)

	for i := range children {
		child := children[i]

		if child.NameConstant() != "" {
			f.ConstantStrings = append(
				f.ConstantStrings,
				fmt.Sprintf("%s = %q", child.UniqueName, child.NameConstant()),
			)
		}
	}

	f.TemplateBody = NamesTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const NamesTemplate = `{{ .Boilerplate }}

package constants

{{ if .Builder.HasChildResources }}
// this package includes the constants which include the resource names.  it is a standalone
// package to prevent import cycle errors when attempting to reference the names from other
// packages (e.g. mutate).
const (
	{{ range .ConstantStrings }}
	{{- . }}
	{{ end }}
)
{{ end }}
`
