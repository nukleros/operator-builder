// Copyright 2024 Nukleros
// SPDX-License-Identifier: MIT

package resources

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"

	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
	"github.com/nukleros/operator-builder/internal/workload/v1/manifests"
)

var (
	_ machinery.Template = &Mutate{}
)

// Mutate scaffolds the root command file for the companion CLI.
type Mutate struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	// input variables
	Builder       kinds.WorkloadBuilder
	ChildResource manifests.ChildResource
}

func (f *Mutate) SetTemplateDefaults() error {
	// set interface variables
	f.Path = filepath.Join(
		"apis",
		f.Resource.Group,
		f.Resource.Version,
		f.Builder.GetPackageName(),
		"mutate",
		f.ChildResource.MutateFileName(),
	)

	f.TemplateBody = MutateTemplate

	return nil
}

// GetIfExistsAction implements file.Builder interface.
func (*Mutate) GetIfExistsAction() machinery.IfExistsAction {
	return machinery.SkipFile
}

//nolint:lll
const MutateTemplate = `{{ .Boilerplate }}

package mutate

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"

	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	{{- if .Builder.IsComponent }}
	{{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }} "{{ .Repo }}/apis/{{ .Builder.GetCollection.Spec.API.Group }}/{{ .Builder.GetCollection.Spec.API.Version }}"
	{{ end -}}
)

// {{ .ChildResource.MutateFuncName }} mutates the {{ .ChildResource.Kind }} resource with name {{ .ChildResource.NameComment }}.
func {{ .ChildResource.MutateFuncName }} (
	original client.Object,
	parent *{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}, {{ if .Builder.IsComponent }}collection *{{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }}.{{ .Builder.GetCollection.Spec.API.Kind }},{{ end }}
	reconciler workload.Reconciler, req *workload.Request,
) ([]client.Object, error) {
	// if either the reconciler or request are found to be nil, return the base object.
	if reconciler == nil || req == nil {
		return []client.Object{original}, nil
	}

	// mutation logic goes here

	return []client.Object{original}, nil
}
`
