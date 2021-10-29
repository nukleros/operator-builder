// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package api

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

var _ machinery.Template = &Types{}

// Types scaffolds a workload's API type.
type Types struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	SpecFields    *workloadv1.APIFields
	ClusterScoped bool
	Dependencies  []*workloadv1.ComponentWorkload
	IsStandalone  bool
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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"{{ .Repo }}/apis/common"
	{{- $Repo := .Repo }}{{- $Added := "" }}{{- range .Dependencies }}
	{{- if ne .Spec.API.Group $.Resource.Group }}
	{{- if not (containsString (printf "%s%s" .Spec.API.Group .Spec.API.Version) $Added) }}
	{{- $Added = (printf "%s%s" $Added (printf "%s%s" .Spec.API.Group .Spec.API.Version)) }}
	{{ .Spec.API.Group }}{{ .Spec.API.Version }} "{{ $Repo }}/apis/{{ .Spec.API.Group }}/{{ .Spec.API.Version }}"
	{{ end }}
	{{ end }}
	{{ end }}
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

{{ .SpecFields.GenerateAPISpec .Resource.Kind }}

// {{ .Resource.Kind }}Status defines the observed state of {{ .Resource.Kind }}.
type {{ .Resource.Kind }}Status struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Created               bool                       ` + "`" + `json:"created,omitempty"` + "`" + `
	DependenciesSatisfied bool                       ` + "`" + `json:"dependenciesSatisfied,omitempty"` + "`" + `
	Conditions            []common.PhaseCondition    ` + "`" + `json:"conditions,omitempty"` + "`" + `
	Resources             []common.Resource          ` + "`" + `json:"resources,omitempty"` + "`" + `
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
{{- if .ClusterScoped }}
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
func (component *{{ .Resource.Kind }}) SetReadyStatus(status bool) {
	component.Status.Created = status
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
func (component {{ .Resource.Kind }}) GetPhaseConditions() []common.PhaseCondition {
	return component.Status.Conditions
}

// SetPhaseCondition sets the phase conditions for a component.
func (component *{{ .Resource.Kind }}) SetPhaseCondition(condition common.PhaseCondition) {
	if found := condition.GetPhaseConditionIndex(component); found >= 0 {
		if condition.LastModified == "" {
			condition.LastModified = time.Now().UTC().String()
		}
		component.Status.Conditions[found] = condition
	} else {
		component.Status.Conditions = append(component.Status.Conditions, condition)
	}
}

// GetResources returns the resources for a component.
func (component {{ .Resource.Kind }}) GetResources() []common.Resource {
	return component.Status.Resources
}

// SetResources sets the phase conditions for a component.
func (component *{{ .Resource.Kind }}) SetResource(resource common.Resource) {

	if found := resource.GetResourceIndex(component); found >= 0 {
		if resource.ResourceCondition.LastModified == "" {
			resource.ResourceCondition.LastModified = time.Now().UTC().String()
		}
		component.Status.Resources[found] = resource
	} else {
		component.Status.Resources = append(component.Status.Resources, resource)
	}
}

// GetDependencies returns the dependencies for a component.
func (*{{ .Resource.Kind }}) GetDependencies() []common.Component {
	return []common.Component{
		{{- range .Dependencies }}
		{{- if eq .Spec.API.Group $.Resource.Group }}
			&{{ .Spec.API.Kind }}{},
		{{- else }}
			&{{ .Spec.API.Group }}{{ .Spec.API.Version }}.{{ .Spec.API.Kind }}{},
		{{- end }}
		{{- end }}
	}
}

// GetComponentGVK returns a GVK object for the component.
func (*{{ .Resource.Kind }}) GetComponentGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   GroupVersion.Group,
		Version: GroupVersion.Version,
		Kind:    "{{ .Resource.Kind }}",
	}
}

func init() {
	SchemeBuilder.Register(&{{ .Resource.Kind }}{}, &{{ .Resource.Kind }}List{})
}
`
