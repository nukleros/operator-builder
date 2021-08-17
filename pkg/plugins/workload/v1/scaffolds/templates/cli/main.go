package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Main{}

// Main scaffolds the main package for the companion CLI.
type Main struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin

	// RootCmd is the root command for the companion CLI
	RootCmd        string
	RootCmdVarName string
}

func (f *Main) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.RootCmd, "main.go")

	f.TemplateBody = cliMainTemplate

	return nil
}

const cliMainTemplate = `{{ .Boilerplate }}

package main

import (
	"{{ .Repo }}/cmd/{{ .RootCmd }}/commands"
)

func main() {
	{{ .RootCmd }} := commands.New{{ .RootCmdVarName }}Command()
	{{ .RootCmd }}.Run()
}
`
