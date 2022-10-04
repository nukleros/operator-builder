// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/nukleros/operator-builder/internal/workload/v1/commands/companion"
	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

var (
	_ machinery.Template = &CmdRoot{}
	_ machinery.Inserter = &CmdRootUpdater{}
)

// CmdRoot scaffolds the root command file for the companion CLI.
type CmdRoot struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin

	// input variables
	Initializer kinds.WorkloadBuilder

	// template variables
	RootCmd      companion.CLI
	IsCollection bool
}

func (f *CmdRoot) SetTemplateDefaults() error {
	// set template variables
	f.IsCollection = f.Initializer.IsCollection()
	f.RootCmd = *f.Initializer.GetRootCommand()

	// set interface variables
	f.Path = filepath.Join("cmd", f.RootCmd.Name, "commands", "root.go")
	f.TemplateBody = fmt.Sprintf(CmdRootTemplate,
		machinery.NewMarkerFor(f.Path, subcommandsImportsMarker),
		machinery.NewMarkerFor(f.Path, subcommandsInitMarker),
		machinery.NewMarkerFor(f.Path, subcommandsGenerateMarker),
		machinery.NewMarkerFor(f.Path, subcommandsVersionMarker),
	)

	return nil
}

// CmdRootUpdater updates root.go to run sub commands.
type CmdRootUpdater struct { //nolint:maligned
	machinery.RepositoryMixin
	machinery.MultiGroupMixin
	machinery.ResourceMixin

	// input variables
	Builder         kinds.WorkloadBuilder
	InitCommand     bool
	GenerateCommand bool
	VersionCommand  bool

	// template variables
	RootCmdName string
}

// GetPath implements file.Builder interface.
func (f *CmdRootUpdater) GetPath() string {
	return filepath.Join("cmd", f.Builder.GetRootCommand().Name, "commands", "root.go")
}

// GetIfExistsAction implements file.Builder interface.
func (*CmdRootUpdater) GetIfExistsAction() machinery.IfExistsAction {
	return machinery.OverwriteFile
}

const subcommandsImportsMarker = "operator-builder:subcommands:imports"
const subcommandsInitMarker = "operator-builder:subcommands:init"
const subcommandsGenerateMarker = "operator-builder:subcommands:generate"
const subcommandsVersionMarker = "operator-builder:subcommands:version"

// GetMarkers implements file.Inserter interface.
func (f *CmdRootUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.GetPath(), subcommandsImportsMarker),
		machinery.NewMarkerFor(f.GetPath(), subcommandsInitMarker),
		machinery.NewMarkerFor(f.GetPath(), subcommandsGenerateMarker),
		machinery.NewMarkerFor(f.GetPath(), subcommandsVersionMarker),
	}
}

// Code Fragments.
const (
	subcommandCodeFragment = `%s%s.New%sSubCommand(parentCommand)
`
	importSubCommandCodeFragment = `%s%s "%s"
`
)

// GetCodeFragments implements file.Inserter interface.
func (f *CmdRootUpdater) GetCodeFragments() machinery.CodeFragmentsMap {
	fragments := make(machinery.CodeFragmentsMap, 1)

	// If resource is not being provided we are creating the file, not updating it
	if f.Resource == nil {
		return fragments
	}

	f.RootCmdName = f.Builder.GetRootCommand().Name

	// Generate a command path for imports
	commandPath := fmt.Sprintf("%s/cmd/%s/commands", f.Repo, f.RootCmdName)

	// Generate subCommands and imports code fragments
	imports := make([]string, 0)
	generateCommands := make([]string, 0)
	initCommands := make([]string, 0)
	versionCommands := make([]string, 0)

	if f.InitCommand {
		imports = append(imports, fmt.Sprintf(importSubCommandCodeFragment,
			"init",
			f.Builder.GetAPIGroup(),
			fmt.Sprintf("%s/init/%s", commandPath, f.Builder.GetAPIGroup())),
		)

		initCommands = append(initCommands, fmt.Sprintf(subcommandCodeFragment,
			"init",
			f.Builder.GetAPIGroup(),
			f.Builder.GetAPIKind()),
		)
	}

	// scaffold the generate command code fragments unless we have a collection without resources
	if (f.Builder.HasChildResources() && f.Builder.IsCollection()) || !f.Builder.IsCollection() {
		if f.GenerateCommand {
			imports = append(imports, fmt.Sprintf(importSubCommandCodeFragment,
				"generate",
				f.Builder.GetAPIGroup(),
				fmt.Sprintf("%s/generate/%s", commandPath, f.Builder.GetAPIGroup())),
			)

			generateCommands = append(generateCommands, fmt.Sprintf(subcommandCodeFragment,
				"generate",
				f.Builder.GetAPIGroup(),
				f.Builder.GetAPIKind()),
			)
		}
	}

	if f.VersionCommand {
		imports = append(imports, fmt.Sprintf(importSubCommandCodeFragment,
			"version",
			f.Builder.GetAPIGroup(),
			fmt.Sprintf("%s/version/%s", commandPath, f.Builder.GetAPIGroup())),
		)

		versionCommands = append(versionCommands, fmt.Sprintf(subcommandCodeFragment,
			"version",
			f.Builder.GetAPIGroup(),
			f.Builder.GetAPIKind()),
		)
	}

	// Only store code fragments in the map if the slices are non-empty
	if len(imports) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), subcommandsImportsMarker)] = imports
	}

	if len(initCommands) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), subcommandsInitMarker)] = initCommands
	}

	if len(generateCommands) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), subcommandsGenerateMarker)] = generateCommands
	}

	if len(versionCommands) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), subcommandsVersionMarker)] = versionCommands
	}

	return fragments
}

const CmdRootTemplate = `{{ .Boilerplate }}

package commands

import (
	"github.com/spf13/cobra"

	// common imports for subcommands
	cmdinit "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/init"
	cmdgenerate "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/generate"
	cmdversion "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/version"

	// specific imports for workloads
	%s
)

// {{ .RootCmd.VarName }}Command represents the base command when called without any subcommands.
type {{ .RootCmd.VarName }}Command struct {
	*cobra.Command
}

// New{{ .RootCmd.VarName }}Command returns an instance of the {{ .RootCmd.VarName }}Command.
func New{{ .RootCmd.VarName }}Command() *{{ .RootCmd.VarName }}Command {
	c := &{{ .RootCmd.VarName }}Command{
		Command: &cobra.Command{
			Use:   "{{ .RootCmd.Name }}",
			Short: "{{ .RootCmd.Description }}",
			Long:  "{{ .RootCmd.Description }}",
		},
	}

	c.addSubCommands()

	return c
}

// Run represents the main entry point into the command
// This is called by main.main() to execute the root command.
func (c *{{ .RootCmd.VarName }}Command) Run() {
	cobra.CheckErr(c.Execute())
}

func (c *{{ .RootCmd.VarName }}Command) newInitSubCommand() {
	{{- if .IsCollection }}
	parentCommand := cmdinit.GetParent(cmdinit.NewBaseInitSubCommand(c.Command))
	{{ else }}
	parentCommand := cmdinit.GetParent(c.Command)
	{{ end -}}
	_ = parentCommand

	// add the init subcommands
	%s
}

func (c *{{ .RootCmd.VarName }}Command) newGenerateSubCommand() {
	{{- if .IsCollection }}
	parentCommand := cmdgenerate.GetParent(cmdgenerate.NewBaseGenerateSubCommand(c.Command))
	{{ else }}
	parentCommand := cmdgenerate.GetParent(c.Command)
	{{ end -}}
	_ = parentCommand

	// add the generate subcommands
	%s
}

func (c *{{ .RootCmd.VarName }}Command) newVersionSubCommand() {
	{{- if .IsCollection }}
	parentCommand := cmdversion.GetParent(cmdversion.NewBaseVersionSubCommand(c.Command))
	{{ else }}
	parentCommand := cmdversion.GetParent(c.Command)
	{{ end -}}
	_ = parentCommand

	// add the version subcommands
	%s
}

// addSubCommands adds any additional subCommands to the root command.
func (c *{{ .RootCmd.VarName }}Command) addSubCommands() {
	c.newInitSubCommand()
	c.newGenerateSubCommand()
	c.newVersionSubCommand()
}
`
