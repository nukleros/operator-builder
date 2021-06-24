package api

import (
	"fmt"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

var _ machinery.Template = &Types{}

// Types scaffolds a workload's API type
type Types struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	SpecFields    *[]workloadv1.APISpecField
	ClusterScoped bool
	Dependencies  *[]workloadv1.ComponentWorkload
	IsStandalone  bool
}

// SetTemplateDefaults implements file.Template
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

var typesTemplate = `{{ .Boilerplate }}

package {{ .Resource.Version }}

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	{{ if not .IsStandalone }}"k8s.io/apimachinery/pkg/runtime/schema"{{ end }}

	common "{{ .Repo }}/apis/common"
	{{- $Repo := .Repo }}{{- range .Dependencies }}
	{{- if ne .Spec.APIGroup $.Resource.Group }}
	{{ .Spec.APIGroup }}{{ .Spec.APIVersion }} "{{ $Repo }}/apis/{{ .Spec.APIGroup }}/{{ .Spec.APIVersion }}"
	{{ end }}
	{{ end }}
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// {{ .Resource.Kind }}Spec defines the desired state of {{ .Resource.Kind }}
type {{ .Resource.Kind }}Spec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	{{ range .SpecFields }}
		{{ if .DefaultVal }}
			// +kubebuilder:default={{ .DefaultVal }}
			// +kubebuilder:validation:Optional
			{{ .ApiSpecContent }}
		{{ else }}
			{{ .ApiSpecContent }}
		{{ end }}
	{{ end }}
}

// {{ .Resource.Kind }}Status defines the observed state of {{ .Resource.Kind }}
type {{ .Resource.Kind }}Status struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Created               bool               ` + "`" + `json:"created,omitempty"` + "`" + `
	DependenciesSatisfied bool               ` + "`" + `json:"dependenciesSatisfied,omitempty"` + "`" + `
	Conditions            []common.Condition ` + "`" + `json:"conditions,omitempty"` + "`" + `
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
{{- if .ClusterScoped }}
// +kubebuilder:resource:scope=Cluster
{{ end }}

// {{ .Resource.Kind }} is the Schema for the {{ .Resource.Plural }} API
type {{ .Resource.Kind }} struct {
	metav1.TypeMeta   ` + "`" + `json:",inline"` + "`" + `
	metav1.ObjectMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `
	Spec   {{ .Resource.Kind }}Spec   ` + "`" + `json:"spec,omitempty"` + "`" + `
	Status {{ .Resource.Kind }}Status ` + "`" + `json:"status,omitempty"` + "`" + `
}

// +kubebuilder:object:root=true

// {{ .Resource.Kind }}List contains a list of {{ .Resource.Kind }}
type {{ .Resource.Kind }}List struct {
	metav1.TypeMeta ` + "`" + `json:",inline"` + "`" + `
	metav1.ListMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `
	Items           []{{ .Resource.Kind }} ` + "`" + `json:"items"` + "`" + `
}

// interface methods

// GetReadyStatus returns the ready status for a component
func (component {{ .Resource.Kind }}) GetReadyStatus() bool {
	return component.Status.Created
}

{{- if not .IsStandalone }}
// SetReadyStatus sets the ready status for a component
func (component *{{ .Resource.Kind }}) SetReadyStatus(status bool) {
	component.Status.Created = status
}

// GetDependencyStatus returns the dependency status for a component
func (component *{{ .Resource.Kind }}) GetDependencyStatus() bool {
	return component.Status.DependenciesSatisfied
}

// SetDependencyStatus sets the dependency status for a component
func (component *{{ .Resource.Kind }}) SetDependencyStatus(dependencyStatus bool) {
	component.Status.DependenciesSatisfied = dependencyStatus
}
{{ end }}

// GetStatusConditions returns the status conditions for a component
func (component {{ .Resource.Kind }}) GetStatusConditions() []common.Condition {
	return component.Status.Conditions
}

// SetStatusConditions sets the status conditions for a component
func (component *{{ .Resource.Kind }}) SetStatusConditions(condition common.Condition) {
	component.Status.Conditions = append(component.Status.Conditions, condition)
}

{{- if not .IsStandalone }}
// GetDependencies returns the dependencies for a component
func (component {{ .Resource.Kind }}) GetDependencies() []common.Component {
	return []common.Component{
		{{ range .Dependencies }}
		{{- if eq .Spec.APIGroup $.Resource.Group }}
			&{{ .Spec.APIKind }}{},
		{{ else }}
			&{{ .Spec.APIGroup }}{{ .Spec.APIVersion }}.{{ .Spec.APIKind }}{},
		{{ end }}
		{{ end }}
	}
}

// GetComponentGVK returns a GVK object for the component
func (component {{ .Resource.Kind }}) GetComponentGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   GroupVersion.Group,
		Version: GroupVersion.Version,
		Kind:    "{{ .Resource.Kind }}",
	}
}
{{ end }}

func init() {
	SchemeBuilder.Register(&{{ .Resource.Kind }}{}, &{{ .Resource.Kind }}List{})
}
`
