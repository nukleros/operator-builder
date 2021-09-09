package templates

import (
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &GoMod{}

// GoMod scaffolds a file that defines the project dependencies.
type GoMod struct {
	machinery.TemplateMixin
	machinery.RepositoryMixin

	ControllerRuntimeVersion string
	CobraVersion             string
}

func (f *GoMod) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = "go.mod"
	}

	f.TemplateBody = goModTemplate

	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const goModTemplate = `
module {{ .Repo }}

go 1.15

require (
	sigs.k8s.io/controller-runtime {{ .ControllerRuntimeVersion }}
	github.com/spf13/cobra {{ .CobraVersion }}
)
`
