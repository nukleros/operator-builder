// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package resources

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

var _ machinery.Template = &Definition{}

// Types scaffolds the child resource definition files.
type Definition struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	// input fields
	Builder    workloadv1.WorkloadAPIBuilder
	SourceFile workloadv1.SourceFile
}

func (f *Definition) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"apis",
		f.Resource.Group,
		f.Resource.Version,
		f.Builder.GetPackageName(),
		f.SourceFile.Filename,
	)

	f.TemplateBody = definitionTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

//nolint:lll
const definitionTemplate = `{{ .Boilerplate }}

package {{ .Builder.GetPackageName }}

import (
	{{ if .SourceFile.HasStatic }}
	"text/template"
	{{ end }}
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	{{- if .SourceFile.HasStatic }}
	k8s_yaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	{{ end }}

	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	{{- if .Builder.IsComponent }}
	{{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }} "{{ .Repo }}/apis/{{ .Builder.GetCollection.Spec.API.Group }}/{{ .Builder.GetCollection.Spec.API.Version }}"
	{{ end -}}
)

{{ range .SourceFile.Children }}
// Create{{ .UniqueName }} creates the {{ .Name }} {{ .Kind }} resource.
func Create{{ .UniqueName }} (
	parent *{{ $.Resource.ImportAlias }}.{{ $.Resource.Kind }},
	{{- if $.Builder.IsComponent }}
	collection *{{ $.Builder.GetCollection.Spec.API.Group }}{{ $.Builder.GetCollection.Spec.API.Version }}.{{ $.Builder.GetCollection.Spec.API.Kind }},
	{{ end -}}
) (client.Object, error) {
	{{- .SourceCode }}

	{{ if not $.Builder.IsClusterScoped }}
	resourceObj.SetNamespace(parent.Namespace)
	{{ end }}

	return resourceObj, nil
}
{{ end }}
`
