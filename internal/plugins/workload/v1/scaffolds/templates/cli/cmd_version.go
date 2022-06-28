// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/kinds"
)

const (
	versionCommandName  = "version"
	versionCommandDescr = "display the version information"
)

var _ machinery.Template = &CmdVersion{}

// CmdVersion scaffolds the companion CLI's version subcommand for
// component workloads.  The version logic will live in the workload's
// subcommand to this command; see cmd_version_sub.go.
type CmdVersion struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin

	// input variables
	Initializer kinds.WorkloadBuilder

	// template variables
	VersionCommandName  string
	VersionCommandDescr string
}

func (f *CmdVersion) SetTemplateDefaults() error {
	// set template variables
	f.VersionCommandName = versionCommandName
	f.VersionCommandDescr = versionCommandDescr

	// set interface variables
	f.Path = filepath.Join("cmd", f.Initializer.GetRootCommand().Name, "commands", "version", "version.go")
	f.TemplateBody = cliCmdVersionTemplate

	return nil
}

const cliCmdVersionTemplate = `{{ .Boilerplate }}

package version

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var CLIVersion = "dev"

type VersionInfo struct {
	CLIVersion  string   ` + "`" + `json:"cliVersion"` + "`" + `
	APIVersions []string ` + "`" + `json:"apiVersions"` + "`" + `
}

type VersionFunc func(*VersionSubCommand) error

type VersionSubCommand struct {
	*cobra.Command

	// options
	Name         string
	Description  string
	SubCommandOf *cobra.Command

	VersionFunc VersionFunc
}

{{ if .Initializer.IsCollection }}
// NewBaseVersionSubCommand returns a subcommand that is meant to belong to a parent
// subcommand but have subcommands itself.
func NewBaseVersionSubCommand(parentCommand *cobra.Command) *VersionSubCommand {
	versionCmd := &VersionSubCommand{
		Name:         "{{ .VersionCommandName }}",
		Description:  "{{ .VersionCommandDescr }}",
		SubCommandOf: parentCommand,
	}

	versionCmd.Setup()

	return versionCmd
}
{{ end }}

// Setup sets up this command to be used as a command.
func (v *VersionSubCommand) Setup() {
	v.Command = &cobra.Command{
		Use:   v.Name,
		Short: v.Description,
		Long:  v.Description,
	}

	// run the version function if the function signature is set
	if v.VersionFunc != nil {
		v.RunE = v.version
	}

	// add this as a subcommand of another command if set
	if v.SubCommandOf != nil {
		v.SubCommandOf.AddCommand(v.Command)
	}
}

// version run the function to display version information about a workload.
func (v *VersionSubCommand) version(cmd *cobra.Command, args []string) error {
	return v.VersionFunc(v)
}

// GetParent is a convenience function written when the CLI code is scaffolded 
// to return the parent command and avoid scaffolding code with bad imports.
func GetParent(c interface{}) *cobra.Command {
	switch subcommand := c.(type) {
	case *VersionSubCommand:
		return subcommand.Command
	case *cobra.Command:
		return subcommand
	}

	panic(fmt.Sprintf("subcommand is not proper type: %T", c))
}

// Display will parse and print the information stored on the VersionInfo object.
func (v *VersionInfo) Display() error {
	output, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to determine versionInfo, %s", err)
	}

	outputStream := os.Stdout

	if _, err := outputStream.WriteString(fmt.Sprintln(string(output))); err != nil {
		return fmt.Errorf("failed to write to stdout, %s", err)
	}

	return nil
}
`

