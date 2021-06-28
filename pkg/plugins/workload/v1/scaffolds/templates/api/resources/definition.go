package resources

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

var _ machinery.Template = &Definition{}

// Types scaffolds the child resource definition files.
type Definition struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	ClusterScoped bool
	SourceFile    workloadv1.SourceFile
	PackageName   string
	SpecFields    *[]workloadv1.APISpecField
	IsComponent   bool
	Collection    *workloadv1.WorkloadCollection
}

func (f *Definition) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"apis",
		f.Resource.Group,
		f.Resource.Version,
		f.PackageName,
		f.SourceFile.Filename,
	)

	f.TemplateBody = definitionTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const definitionTemplate = `{{ .Boilerplate }}

package {{ .PackageName }}

import (
	{{ if .SourceFile.HasStatic }}
	"text/template"
	{{ end }}
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	{{- if .SourceFile.HasStatic }}
	k8s_yaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	{{ end }}

	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	{{- if .IsComponent }}
	{{ .Collection.Spec.APIGroup }}{{ .Collection.Spec.APIVersion }} "{{ .Repo }}/apis/{{ .Collection.Spec.APIGroup }}/{{ .Collection.Spec.APIVersion }}"
	{{ end -}}
)

{{ range .SourceFile.Children }}
{{ if .StaticCreateStrategy }}
const resource{{ .UniqueName }} = ` + "`" + `
{{ .StaticContent }}
` + "`" + `

// Create{{ .UniqueName }} creates the {{ .Name }} {{ .Kind }} resource
func Create{{ .UniqueName }}(
	parent *{{ $.Resource.ImportAlias }}.{{ $.Resource.Kind }},
	{{- if $.IsComponent }}
	collection *{{ $.Collection.Spec.APIGroup }}{{ $.Collection.Spec.APIVersion }}.{{ $.Collection.Spec.APIKind }},
	{{ end -}}
) (metav1.Object, error) {

	fmap := template.FuncMap{
		{{ range $.SpecFields }}
		{{- if .DefaultVal -}}
		"default{{ .FieldName }}": default{{ .FieldName }},
		{{- end }}
		{{ end }}
	}

	childContent, err := runTemplate("{{ .Name }}", resource{{ .UniqueName }}, parent, fmap)
	if err != nil {
		return nil, err
	}

	// decode YAML into unstructured.Unstructured
	resourceObj := &unstructured.Unstructured{}
	decode := k8s_yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err = decode.Decode([]byte(childContent), nil, resourceObj)
	if err != nil {
		return nil, err
	}

{{ else }}
// Create{{ .UniqueName }} creates the {{ .Name }} {{ .Kind }} resource
func Create{{ .UniqueName }} (
	parent *{{ $.Resource.ImportAlias }}.{{ $.Resource.Kind }},
	{{- if $.IsComponent }}
	collection *{{ $.Collection.Spec.APIGroup }}{{ $.Collection.Spec.APIVersion }}.{{ $.Collection.Spec.APIKind }},
	{{ end -}}
) (metav1.Object, error) {

	{{ .SourceCode }}

{{ end }}
	{{ if not $.ClusterScoped }}
	resourceObj.SetNamespace(parent.Namespace)
	{{ end }}

	return resourceObj, nil
}
{{ end }}
`
