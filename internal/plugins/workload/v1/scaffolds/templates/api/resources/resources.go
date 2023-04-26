// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package resources

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/nukleros/operator-builder/internal/plugins/workload/v1/scaffolds/templates/config/samples"
	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

var _ machinery.Template = &Resources{}

// Types scaffolds child resource creation functions.
type Resources struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	// input fields
	Builder kinds.WorkloadBuilder

	// template fields
	SpecFields      *kinds.APIFields
	IsClusterScoped bool
	CreateFuncNames []string
	InitFuncNames   []string
}

func (f *Resources) SetTemplateDefaults() error {
	// set template fields
	f.CreateFuncNames, f.InitFuncNames = f.Builder.GetManifests().FuncNames()
	f.SpecFields = f.Builder.GetAPISpecFields()
	f.IsClusterScoped = f.Builder.IsClusterScoped()

	// set interface fields
	f.Path = filepath.Join(
		"apis",
		f.Resource.Group,
		f.Resource.Version,
		f.Builder.GetPackageName(),
		"resources.go",
	)

	f.TemplateBody = fmt.Sprintf(resourcesTemplate,
		samples.SampleTemplate,
		samples.SampleTemplateRequiredOnly,
	)
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

//nolint:lll
const resourcesTemplate = `{{ .Boilerplate }}

package {{ .Builder.GetPackageName }}

import (
	{{ if ne .Builder.GetRootCommand.Name "" }}"fmt"{{ end }}

	{{ if ne .Builder.GetRootCommand.Name "" }}"sigs.k8s.io/yaml"{{ end }}
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"

	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	{{- if .Builder.IsComponent }}
	{{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }} "{{ .Repo }}/apis/{{ .Builder.GetCollection.Spec.API.Group }}/{{ .Builder.GetCollection.Spec.API.Version }}"
	{{ end -}}
)

// sample{{ .Resource.Kind }} is a sample containing all fields
const sample{{ .Resource.Kind }} = ` + "`" + `%s` + "`" + `

// sample{{ .Resource.Kind }}Required is a sample containing only required fields
const sample{{ .Resource.Kind }}Required = ` + "`" + `%s` + "`" + `

// Sample returns the sample manifest for this custom resource.
func Sample(requiredOnly bool) string {
	if requiredOnly {
		return sample{{ .Resource.Kind }}Required
	}

	return sample{{ .Resource.Kind }}
}

// Generate returns the child resources that are associated with this workload given
// appropriate structured inputs.
{{ if .Builder.IsComponent -}}
func Generate(
	workloadObj {{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	collectionObj {{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }}.{{ .Builder.GetCollection.Spec.API.Kind }},
{{ else if .Builder.IsCollection -}}
func Generate(
	collectionObj {{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }}.{{ .Builder.GetCollection.Spec.API.Kind }},
{{ else -}}
func Generate(
	workloadObj {{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
{{ end -}}
	reconciler workload.Reconciler,
	req *workload.Request,
) ([]client.Object, error) {
	resourceObjects := []client.Object{}

	for _, f := range CreateFuncs {
		{{ if .Builder.IsComponent -}}
		resources, err := f(&workloadObj, &collectionObj, reconciler, req)
		{{ else if .Builder.IsCollection -}}
		resources, err := f(&collectionObj, reconciler, req)
		{{ else -}}
		resources, err := f(&workloadObj, reconciler, req)
		{{ end }}
		if err != nil {
			return nil, err
		}

		resourceObjects = append(resourceObjects, resources...)
	}

	return resourceObjects, nil
}

{{ if ne .Builder.GetRootCommand.Name "" }}
// GenerateForCLI returns the child resources that are associated with this workload given
// appropriate YAML manifest files.
func GenerateForCLI(
	{{- if or (.Builder.IsStandalone) (.Builder.IsComponent) }}workloadFile []byte,{{ end -}}
	{{- if or (.Builder.IsComponent) (.Builder.IsCollection) }}collectionFile []byte,{{ end -}}
) ([]client.Object, error) {
	{{- if or (.Builder.IsStandalone) (.Builder.IsComponent) }}
	var workloadObj {{ .Resource.ImportAlias }}.{{ .Resource.Kind }}
	if err := yaml.Unmarshal(workloadFile, &workloadObj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml into workload, %%w", err)
	}

	if err := workload.Validate(&workloadObj); err != nil {
		return nil, fmt.Errorf("error validating workload yaml, %%w", err)
	}
	{{ end }}

	{{- if or (.Builder.IsComponent) (.Builder.IsCollection) }}
	var collectionObj {{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }}.{{ .Builder.GetCollection.Spec.API.Kind }}
	if err := yaml.Unmarshal(collectionFile, &collectionObj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml into collection, %%w", err)
	}

	if err := workload.Validate(&collectionObj); err != nil {
		return nil, fmt.Errorf("error validating collection yaml, %%w", err)
	}
	{{ end }}

	{{ if .Builder.IsComponent }}
	return Generate(workloadObj, collectionObj, nil, nil)
	{{ else if .Builder.IsCollection }}
	return Generate(collectionObj, nil, nil)
	{{ else }}
	return Generate(workloadObj, nil, nil)
	{{ end -}}
}
{{ end }}

// CreateFuncs is an array of functions that are called to create the child resources for the controller
// in memory during the reconciliation loop prior to persisting the changes or updates to the Kubernetes
// database.
var CreateFuncs = []func(
	*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	{{ if $.Builder.IsComponent -}}
	*{{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }}.{{ .Builder.GetCollection.Spec.API.Kind }},
	{{ end -}}
	workload.Reconciler,
	*workload.Request,
) ([]client.Object, error) {
	{{ range .CreateFuncNames }}
		{{- . -}},
	{{ end }}
}

// InitFuncs is an array of functions that are called prior to starting the controller manager.  This is
// necessary in instances which the controller needs to "own" objects which depend on resources to
// pre-exist in the cluster. A common use case for this is the need to own a custom resource.
// If the controller needs to own a custom resource type, the CRD that defines it must
// first exist. In this case, the InitFunc will create the CRD so that the controller
// can own custom resources of that type.  Without the InitFunc the controller will
// crash loop because when it tries to own a non-existent resource type during manager
// setup, it will fail.
var InitFuncs = []func(
	*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	{{ if $.Builder.IsComponent -}}
	*{{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }}.{{ .Builder.GetCollection.Spec.API.Kind }},
	{{ end -}}
	workload.Reconciler,
	*workload.Request,
) ([]client.Object, error) {
	{{ range .InitFuncNames }}
		{{- . -}},
	{{ end }}
}

{{ if $.Builder.IsComponent -}}
func ConvertWorkload(component, collection workload.Workload) (
	*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	*{{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }}.{{ .Builder.GetCollection.Spec.API.Kind }},
	error,
) {
{{- else }}
func ConvertWorkload(component workload.Workload) (*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}, error) {
{{- end }}
	p, ok := component.(*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }})
	if !ok {
		{{- if $.Builder.IsComponent }}
		return nil, nil, {{ .Resource.ImportAlias }}.ErrUnableToConvert{{ .Resource.Kind }}
	}

	c, ok := collection.(*{{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }}.{{ .Builder.GetCollection.Spec.API.Kind }})
	if !ok {
		return nil, nil, {{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }}.ErrUnableToConvert{{ .Builder.GetCollection.Spec.API.Kind }}
	}

	return p, c, nil
{{- else }}
		return nil, {{ .Resource.ImportAlias }}.ErrUnableToConvert{{ .Resource.Kind }}
  }

	return p, nil
{{- end }}
}
`
