package templates

import (
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Project{}

// Project scaffolds the WORKLOAD project config.
type Project struct {
	machinery.TemplateMixin

	RootCmd string
}

func (f *Project) SetTemplateDefaults() error {
	f.Path = "WORKLOAD"

	f.TemplateBody = projectTemplate

	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const projectTemplate = `
cliRootCommandName: {{ .RootCmd }}
`
