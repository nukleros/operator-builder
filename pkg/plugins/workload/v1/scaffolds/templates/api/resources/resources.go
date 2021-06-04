package resources

import (
	"path/filepath"
	"text/template"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

var _ machinery.Template = &Resources{}

// Types scaffolds child resource creation functions
type Resources struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin

	PackageName     string
	CreateFuncNames []string
	SpecFields      *[]workloadv1.APISpecField
}

func (f *Resources) SetTemplateDefaults() error {

	f.Path = filepath.Join(
		"apis",
		f.Resource.Group,
		f.Resource.Version,
		f.PackageName,
		"resources.go",
	)

	f.TemplateBody = resourcesTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

func (f Resources) GetFuncMap() template.FuncMap {

	funcMap := machinery.DefaultFuncMap()
	funcMap["quotestr"] = func(value string) string {
		if string(value[0]) != `"` {
			value = `"` + value
		}
		if string(value[len(value)-1]) != `"` {
			value = value + `"`
		}
		return value
	}
	return funcMap
}

const resourcesTemplate = `{{ .Boilerplate }}

package {{ .PackageName }}

import (
	"fmt"
	"bytes"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
)

var CreateFuncs = []func(*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}) (metav1.Object, error){
	{{ range .CreateFuncNames }}
		{{- . -}},
	{{ end }}
}

// runTemplate renders a template for a child object to the custom resource
func runTemplate(templateName, templateValue string, data *{{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	funcMap template.FuncMap) (string, error) {

	t, err := template.New(templateName).Funcs(funcMap).Parse(templateValue)
	if err != nil {
		return "", fmt.Errorf("error parsing template %s: %v", templateName, err)
	}

	var b bytes.Buffer
	if err := t.Execute(&b, &data); err != nil {
		return "", fmt.Errorf("error rendering template %s: %v", templateName, err)
	}

	return b.String(), nil
}

{{ range .SpecFields }}
{{ if .DefaultVal }}
{{ if eq .DataType "string" }}
const {{ .ManifestFieldName }}Default = {{ .DefaultVal | quotestr }}
{{ else }}
const {{ .ManifestFieldName }}Default = {{ .DefaultVal }}
{{ end }}

func default{{ .FieldName }}(value {{ .DataType }}) {{ .DataType }} {

	if value == {{ .ZeroVal }} {
		return {{ .ManifestFieldName }}Default
	}

	return value
}
{{ end }}
{{ end }}
`
