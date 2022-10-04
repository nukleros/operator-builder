// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package resources

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
	"github.com/nukleros/operator-builder/internal/workload/v1/manifests"
)

var _ machinery.Template = &Definition{}

// Types scaffolds the child resource definition files.
type Definition struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	// input fields
	Builder  kinds.WorkloadBuilder
	Manifest *manifests.Manifest

	// template fields
	UseStrConv bool
}

func (f *Definition) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"apis",
		f.Resource.Group,
		f.Resource.Version,
		f.Builder.GetPackageName(),
		f.Manifest.SourceFilename,
	)

	// determine if we need to import the strconv package
	for i := range f.Manifest.ChildResources {
		if f.Manifest.ChildResources[i].UseStrConv {
			f.UseStrConv = true

			break
		}
	}

	f.TemplateBody = definitionTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

//nolint:lll
const definitionTemplate = `{{ .Boilerplate }}

package {{ .Builder.GetPackageName }}

import (
	{{ if .UseStrConv }}"strconv"{{ end }}

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"

	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	"{{ .Resource.Path }}/{{ .Builder.GetPackageName }}/mutate"
	{{- if .Builder.IsComponent }}
	{{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }} "{{ .Repo }}/apis/{{ .Builder.GetCollection.Spec.API.Group }}/{{ .Builder.GetCollection.Spec.API.Version }}"
	{{ end -}}
)

{{ range .Manifest.ChildResources }}
{{ range .RBAC }}
{{- .ToMarker }}
{{ end }}
// {{ .CreateFuncName }} creates the {{ .Kind }} resource with name {{ .NameComment }}.
func {{ .CreateFuncName }} (
	parent *{{ $.Resource.ImportAlias }}.{{ $.Resource.Kind }},
	{{ if $.Builder.IsComponent -}}
	collection *{{ $.Builder.GetCollection.Spec.API.Group }}{{ $.Builder.GetCollection.Spec.API.Version }}.{{ $.Builder.GetCollection.Spec.API.Kind }},
	{{ end -}}
	reconciler workload.Reconciler,
	req *workload.Request,
) ([]client.Object, error) {

	{{- if ne .IncludeCode "" }}{{ .IncludeCode }}{{ end }}

	{{- .SourceCode }}

	{{ if not $.Builder.IsClusterScoped }}
	resourceObj.SetNamespace(parent.Namespace)
	{{ end }}

	{{ if $.Builder.IsComponent -}}
	return mutate.{{ .MutateFuncName }}(resourceObj, parent, collection, reconciler, req)
	{{ else -}}
	return mutate.{{ .MutateFuncName }}(resourceObj, parent, reconciler, req)
	{{ end -}}
}
{{ end }}
`
