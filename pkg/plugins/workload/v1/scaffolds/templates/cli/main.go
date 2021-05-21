package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &CliMain{}

// CliMain scaffolds the main package for the companion CLI
type CliMain struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin

	// CliRootCmd is the root command for the companion CLI
	CliRootCmd string
}

// SetTemplateDefaults implements file.Template
func (f *CliMain) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.CliRootCmd, "main.go")

	f.TemplateBody = cliMainTemplate

	return nil
}

var cliMainTemplate = `{{ .Boilerplate }}

package main

import (
	"{{ .Repo }}/cmd/{{ .CliRootCmd }}/commands"
)

func main() {
	commands.Execute()
}
`
