package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

const (
	initCommandName  = "init"
	initCommandDescr = "Write a sample custom resource manifest for a workload to standard out"
)

var _ machinery.Template = &CliCmdInit{}

// CliCmdInit scaffolds the companion CLI's init subcommand for
// component workloads.  The init logic will live in the workload's
// subcommand to this command; see cmd_init_sub.go.
type CliCmdInit struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin

	CliRootCmd string

	InitCommandName  string
	InitCommandDescr string
}

func (f *CliCmdInit) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.CliRootCmd, "commands", "init.go")

	f.InitCommandName = initCommandName
	f.InitCommandDescr = initCommandDescr

	f.TemplateBody = cliCmdInitTemplate

	return nil
}

const cliCmdInitTemplate = `{{ .Boilerplate }}

package commands

import (
	"github.com/spf13/cobra"
)

// initCmd represents the init command.
var initCmd = &cobra.Command{
	Use:   "{{ .InitCommandName }}",
	Short: "{{ .InitCommandDescr }}",
	Long: "{{ .InitCommandDescr }}",
}

func init() {
	rootCmd.AddCommand(initCmd)
}
`
