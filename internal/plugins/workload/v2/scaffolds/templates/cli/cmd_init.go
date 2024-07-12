// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"

	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

const (
	initCommandName  = "init"
	initCommandDescr = "write a sample custom resource manifest for a workload to standard out"
)

var _ machinery.Template = &CmdInit{}

// CmdInit scaffolds the companion CLI's init subcommand for
// component workloads.  The init logic will live in the workload's
// subcommand to this command; see cmd_init_sub.go.
type CmdInit struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin

	// input variables
	Initializer kinds.WorkloadBuilder

	// template variables
	InitCommandName  string
	InitCommandDescr string
}

func (f *CmdInit) SetTemplateDefaults() error {
	// set the template variables
	f.InitCommandName = initCommandName
	f.InitCommandDescr = initCommandDescr

	// set interface variables
	f.Path = filepath.Join("cmd", f.Initializer.GetRootCommand().Name, "commands", "init", "init.go")
	f.TemplateBody = cliCmdInitTemplate

	return nil
}

const cliCmdInitTemplate = `{{ .Boilerplate }}

package init

import (
	"fmt"

	"github.com/spf13/cobra"
)

type InitFunc func(*InitSubCommand) error

type InitSubCommand struct {
	*cobra.Command

	// flags
	APIVersion   string
	RequiredOnly bool

	// options
	Name         string
	Description  string
	SubCommandOf *cobra.Command

	InitFunc InitFunc
}

{{ if .Initializer.IsCollection }}
// NewBaseInitSubCommand returns a subcommand that is meant to belong to a parent
// subcommand but have subcommands itself.
func NewBaseInitSubCommand(parentCommand *cobra.Command) *InitSubCommand {
	initCmd := &InitSubCommand{
		Name:         "{{ .InitCommandName }}",
		Description:  "{{ .InitCommandDescr }}",
		SubCommandOf: parentCommand,
	}

	initCmd.Setup()

	return initCmd
}
{{ end }}

// Setup sets up this command to be used as a command.
func (i *InitSubCommand) Setup() {
	i.Command = &cobra.Command{
		Use:   i.Name,
		Short: i.Description,
		Long:  i.Description,
	}

	// run the initialize function if the function signature is set
	if i.InitFunc != nil {
		i.RunE = i.initialize
	}

	// always add the api-version flag
	i.Flags().StringVarP(
		&i.APIVersion,
		"api-version",
		"",
		"",
		"api version of the workload to generate a workload manifest for",
	)

	// always add the required-only flag
	i.Flags().BoolVarP(
		&i.RequiredOnly,
		"required-only",
		"r",
		false,
		"only print required fields in the manifest output",
	)

	// add this as a subcommand of another command if set
	if i.SubCommandOf != nil {
		i.SubCommandOf.AddCommand(i.Command)
	}
}

// GetParent is a convenience function written when the CLI code is scaffolded 
// to return the parent command and avoid scaffolding code with bad imports.
func GetParent(c interface{}) *cobra.Command {
	switch subcommand := c.(type) {
	case *InitSubCommand:
		return subcommand.Command
	case *cobra.Command:
		return subcommand
	}

	panic(fmt.Sprintf("subcommand is not proper type: %T", c))
}

// initialize creates sample workload manifests for a workload's custom resource.
func (i *InitSubCommand) initialize(cmd *cobra.Command, args []string) error {
	return i.InitFunc(i)
}
`
