package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

const (
	generateCommandName  = "generate"
	generateCommandDescr = "Generate child resource manifests from a workload's custom resource"
)

var _ machinery.Template = &CliCmdGenerate{}

// CliCmdGenerate scaffolds the companion CLI's generate subcommand for
// comopnent workloads.  The generate logic will live in the workload's
// subcommand to this command; see cmd_generate_sub.go.
type CliCmdGenerate struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin

	CliRootCmd string

	GenerateCommandName  string
	GenerateCommandDescr string
}

func (f *CliCmdGenerate) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.CliRootCmd, "commands", "generate.go")

	f.GenerateCommandName = generateCommandName
	f.GenerateCommandDescr = generateCommandDescr

	f.TemplateBody = cliCmdGenerateTemplate

	return nil
}

const cliCmdGenerateTemplate = `{{ .Boilerplate }}

package commands

import (
	"github.com/spf13/cobra"
)

var (
	workloadManifest string
	collectionManifest string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "{{ .GenerateCommandName }}",
	Short: "{{ .GenerateCommandDescr }}",
	Long: "{{ .GenerateCommandDescr }}",
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.PersistentFlags().StringVarP(
		&workloadManifest,
		"workload-manifest",
		"w",
		"",
		"Filepath to the workload manifest to generate child resources for.",
	)
	generateCmd.PersistentFlags().StringVarP(
		&collectionManifest,
		"collection-manifest",
		"c",
		"",
		"Filepath to the workload collection manifest.",
	)
}
`
