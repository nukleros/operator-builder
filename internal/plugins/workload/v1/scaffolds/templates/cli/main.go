// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"path/filepath"
	"text/template"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

var _ machinery.Template = &Main{}

// Main scaffolds the main package for the companion CLI.
type Main struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin

	RootCmd workloadv1.CliCommand
}

func (f *Main) SetTemplateDefaults() error {
	// set interface variables
	f.Path = filepath.Join("cmd", f.RootCmd.Name, "main.go")
	f.TemplateBody = cliMainTemplate

	f.IfExistsAction = machinery.SkipFile

	return nil
}

func (*Main) GetFuncMap() template.FuncMap {
	return utils.RemoveStringHelper()
}

const cliMainTemplate = `{{ .Boilerplate }}

package main

import (
	"{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands"
)

func main() {
	{{ .RootCmd.Name | removeString "-" }} := commands.New{{ .RootCmd.VarName }}Command()
	{{ .RootCmd.Name | removeString "-" }}.Run()
}
`
