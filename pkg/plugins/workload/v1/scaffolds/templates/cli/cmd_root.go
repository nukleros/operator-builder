package cli

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var (
	_ machinery.Template = &CmdRoot{}
	_ machinery.Inserter = &CmdRootUpdater{}
)

// CmdRoot scaffolds the root command file for the companion CLI.
type CmdRoot struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin

	// RootCmd is the root command for the companion CLI
	RootCmd        string
	RootCmdVarName string
	// RootCmdDescription is the command description given by the CLI help info
	RootCmdDescription string
}

func (f *CmdRoot) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.RootCmd, "commands", "root.go")

	f.TemplateBody = fmt.Sprintf(CmdRootTemplate, machinery.NewMarkerFor(f.Path, subcommandsMarker))

	return nil
}

// CmdRootUpdater updates root.go to run sub commands.
type CmdRootUpdater struct { //nolint:maligned
	machinery.RepositoryMixin
	machinery.MultiGroupMixin
	machinery.ResourceMixin

	RootCmd string

	// Flags to indicate which parts need to be included when updating the file
	InitCommand, GenerateCommand bool
}

// GetPath implements file.Builder interface.
func (f *CmdRootUpdater) GetPath() string {
	return filepath.Join("cmd", f.RootCmd, "commands", "root.go")
}

// GetIfExistsAction implements file.Builder interface.
func (*CmdRootUpdater) GetIfExistsAction() machinery.IfExistsAction {
	return machinery.OverwriteFile
}

const subcommandsMarker = "operator-builder:subcommands"

// GetMarkers implements file.Inserter interface.
func (f *CmdRootUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.GetPath(), subcommandsMarker),
	}
}

// Code Fragments.
const (
	initCommandCodeFragment = `c.newInitCommand()
`
	generateCommandCodeFragement = `c.newGenerateCommand()
`
)

// GetCodeFragments implements file.Inserter interface.
func (f *CmdRootUpdater) GetCodeFragments() machinery.CodeFragmentsMap {
	fragments := make(machinery.CodeFragmentsMap, 1)

	// If resource is not being provided we are creating the file, not updating it
	if f.Resource == nil {
		return fragments
	}

	// Generate subCommands code fragments
	subCommands := make([]string, 0)
	if f.InitCommand {
		subCommands = append(subCommands, initCommandCodeFragment)
	}

	if f.GenerateCommand {
		subCommands = append(subCommands, generateCommandCodeFragement)
	}

	// Only store code fragments in the map if the slices are non-empty
	if len(subCommands) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), subcommandsMarker)] = subCommands
	}

	return fragments
}

const CmdRootTemplate = `{{ .Boilerplate }}

package commands

import (
	"github.com/spf13/cobra"
)

// {{ .RootCmdVarName }}Command represents the base command when called without any subcommands.
type {{ .RootCmdVarName }}Command struct {
	*cobra.Command
}

// New{{ .RootCmdVarName }}Command returns an instance of the {{ .RootCmdVarName }}Command.
func New{{ .RootCmdVarName }}Command() *{{ .RootCmdVarName }}Command {
	c := &{{ .RootCmdVarName }}Command{
		Command: &cobra.Command{
			Use:   "{{ .RootCmd }}",
			Short: "{{ .RootCmdDescription }}",
			Long:  "{{ .RootCmdDescription }}",
		},
	}

	c.addSubCommands()

	return c
}

// Run represents the main entry point into the command
// This is called by main.main() to execute the root command.
func (c *{{ .RootCmdVarName }}Command) Run() {
	cobra.CheckErr(c.Execute())
}

// addSubCommands adds any additional subCommands to the root command.
func (c *{{ .RootCmdVarName }}Command) addSubCommands() {
	%s
}
`
