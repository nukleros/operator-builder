package resources

import (
	"path/filepath"
	"text/template"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

var _ machinery.Template = &Resources{}

// Types scaffolds child resource creation functions.
type Resources struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	PackageName     string
	CreateFuncNames []string
	InitFuncNames   []string
	IsComponent     bool
	Collection      *workloadv1.WorkloadCollection
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

func (f *Resources) GetFuncMap() template.FuncMap {
	funcMap := machinery.DefaultFuncMap()
	funcMap["quotestr"] = func(value string) string {
		if string(value[0]) != `"` {
			value = `"` + value
		}

		if string(value[len(value)-1]) != `"` {
			value += `"`
		}

		return value
	}

	return funcMap
}

//nolint:lll
const resourcesTemplate = `{{ .Boilerplate }}

package {{ .PackageName }}

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	{{- if .IsComponent }}
	{{ .Collection.Spec.APIGroup }}{{ .Collection.Spec.APIVersion }} "{{ .Repo }}/apis/{{ .Collection.Spec.APIGroup }}/{{ .Collection.Spec.APIVersion }}"
	{{ end -}}
)

// CreateFuncs is an array of functions that are called to create the child resources for the controller
// in memory during the reconciliation loop prior to persisting the changes or updates to the Kubernetes
// database.
var CreateFuncs = []func(
	*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	{{- if $.IsComponent }}
	*{{ .Collection.Spec.APIGroup }}{{ .Collection.Spec.APIVersion }}.{{ .Collection.Spec.APIKind }},
	{{ end -}}
) (metav1.Object, error){
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
	{{- if $.IsComponent }}
	*{{ .Collection.Spec.APIGroup }}{{ .Collection.Spec.APIVersion }}.{{ .Collection.Spec.APIKind }},
	{{ end -}}
) (metav1.Object, error){
	{{ range .InitFuncNames }}
		{{- . -}},
	{{ end }}
}
`
