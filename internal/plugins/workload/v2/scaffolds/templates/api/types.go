// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package api

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"

	"github.com/nukleros/operator-builder/internal/utils"
	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

var _ machinery.Template = &Types{}

// Types scaffolds a workload's API type.
type Types struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	// input fields
	Builder kinds.WorkloadBuilder
}

// SetTemplateDefaults implements file.Template.
func (f *Types) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"apis",
		f.Resource.Group,
		f.Resource.Version,
		fmt.Sprintf("%s_types.go", strings.ToLower(f.Resource.Kind)),
	)

	f.TemplateBody = typesTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

func (*Types) GetFuncMap() template.FuncMap {
	return utils.ContainsStringHelper()
}

const typesTemplate = `{{ .Boilerplate }}

package {{ .Resource.Version }}

import (
	"errors"

	"github.com/nukleros/operator-builder-tools/pkg/status"
	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	{{- $Repo := .Repo }}{{- $Added := "" }}{{- range .Builder.GetDependencies }}
	{{- if ne .Spec.API.Group $.Resource.Group }}
	{{- if not (containsString (printf "%s%s" .Spec.API.Group .Spec.API.Version) $Added) }}
	{{- $Added = (printf "%s%s" $Added (printf "%s%s" .Spec.API.Group .Spec.API.Version)) }}
	{{ .Spec.API.Group }}{{ .Spec.API.Version }} "{{ $Repo }}/apis/{{ .Spec.API.Group }}/{{ .Spec.API.Version }}"
	{{ end }}
	{{ end }}
	{{ end }}
)

var ErrUnableToConvert{{ .Resource.Kind }} = errors.New("unable to convert to {{ .Resource.Kind }}")

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

{{ .Builder.GetAPISpecFields.GenerateAPISpec .Resource.Kind }}

// {{ .Resource.Kind }}Status defines the observed state of {{ .Resource.Kind }}.
type {{ .Resource.Kind }}Status struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Created               bool                       ` + "`" + `json:"created,omitempty"` + "`" + `
	DependenciesSatisfied bool                       ` + "`" + `json:"dependenciesSatisfied,omitempty"` + "`" + `
	Conditions            []*status.PhaseCondition   ` + "`" + `json:"conditions,omitempty"` + "`" + `
	Resources             []*status.ChildResource    ` + "`" + `json:"resources,omitempty"` + "`" + `
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
{{- if .Builder.IsClusterScoped }}
// +kubebuilder:resource:scope=Cluster
{{ end }}

// {{ .Resource.Kind }} is the Schema for the {{ .Resource.Plural }} API.
type {{ .Resource.Kind }} struct {
	metav1.TypeMeta   ` + "`" + `json:",inline"` + "`" + `
	metav1.ObjectMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `
	Spec   {{ .Resource.Kind }}Spec   ` + "`" + `json:"spec,omitempty"` + "`" + `
	Status {{ .Resource.Kind }}Status ` + "`" + `json:"status,omitempty"` + "`" + `
}

// +kubebuilder:object:root=true

// {{ .Resource.Kind }}List contains a list of {{ .Resource.Kind }}.
type {{ .Resource.Kind }}List struct {
	metav1.TypeMeta ` + "`" + `json:",inline"` + "`" + `
	metav1.ListMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `
	Items           []{{ .Resource.Kind }} ` + "`" + `json:"items"` + "`" + `
}

// interface methods

// GetReadyStatus returns the ready status for a component.
func (component *{{ .Resource.Kind }}) GetReadyStatus() bool {
	return component.Status.Created
}

// SetReadyStatus sets the ready status for a component.
func (component *{{ .Resource.Kind }}) SetReadyStatus(ready bool) {
	component.Status.Created = ready
}

// GetDependencyStatus returns the dependency status for a component.
func (component *{{ .Resource.Kind }}) GetDependencyStatus() bool {
	return component.Status.DependenciesSatisfied
}

// SetDependencyStatus sets the dependency status for a component.
func (component *{{ .Resource.Kind }}) SetDependencyStatus(dependencyStatus bool) {
	component.Status.DependenciesSatisfied = dependencyStatus
}

// GetPhaseConditions returns the phase conditions for a component.
func (component *{{ .Resource.Kind }}) GetPhaseConditions() []*status.PhaseCondition {
	return component.Status.Conditions
}

// SetPhaseCondition sets the phase conditions for a component.
func (component *{{ .Resource.Kind }}) SetPhaseCondition(condition *status.PhaseCondition) {
	for i, currentCondition := range component.GetPhaseConditions() {
		if currentCondition.Phase == condition.Phase {
			component.Status.Conditions[i] = condition

			return
		}
	}

	// phase not found, lets add it to the list.
	component.Status.Conditions = append(component.Status.Conditions, condition)
}

// GetResources returns the child resource status for a component.
func (component *{{ .Resource.Kind }}) GetChildResourceConditions() []*status.ChildResource {
	return component.Status.Resources
}

// SetResources sets the phase conditions for a component.
func (component *{{ .Resource.Kind }}) SetChildResourceCondition(resource *status.ChildResource) {
	for i, currentResource := range component.GetChildResourceConditions() {
		if currentResource.Group == resource.Group && currentResource.Version == resource.Version && currentResource.Kind == resource.Kind {
			if currentResource.Name == resource.Name && currentResource.Namespace == resource.Namespace {
				component.Status.Resources[i] = resource

				return
			}
		}
	}

	// phase not found, lets add it to the collection
	component.Status.Resources = append(component.Status.Resources, resource)
}

// GetDependencies returns the dependencies for a component.
func (*{{ .Resource.Kind }}) GetDependencies() []workload.Workload {
	return []workload.Workload{
		{{- range .Builder.GetDependencies }}
		{{- if eq .Spec.API.Group $.Resource.Group }}
			&{{ .Spec.API.Kind }}{},
		{{- else }}
			&{{ .Spec.API.Group }}{{ .Spec.API.Version }}.{{ .Spec.API.Kind }}{},
		{{- end }}
		{{- end }}
	}
}

// GetComponentGVK returns a GVK object for the component.
func (*{{ .Resource.Kind }}) GetWorkloadGVK() schema.GroupVersionKind {
	return GroupVersion.WithKind("{{ .Resource.Kind }}")
}

func init() {
	SchemeBuilder.Register(&{{ .Resource.Kind }}{}, &{{ .Resource.Kind }}List{})
}
`
