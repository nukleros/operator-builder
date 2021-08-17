package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

const (
	generateCommandName  = "generate"
	generateCommandDescr = "Generate child resource manifests from a workload's custom resource"
)

var _ machinery.Template = &CmdGenerate{}

// CmdGenerate scaffolds the companion CLI's generate subcommand for
// component workloads.  The generate logic will live in the workload's
// subcommand to this command; see cmd_generate_sub.go.
type CmdGenerate struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin

	RootCmd        string
	RootCmdVarName string

	GenerateCommandName  string
	GenerateCommandDescr string
	Collection           *workloadv1.WorkloadCollection
}

func (f *CmdGenerate) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.RootCmd, "commands", "generate.go")

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

type generateCommand struct{
	*cobra.Command
	workloadManifest string
	collectionManifest string
}

// newGenerateCommand creates a new instance of the generate subcommand.
func (c *{{ .RootCmdVarName }}Command) newGenerateCommand() {
	generateCmd := &generateCommand{}

	generateCmd.Command = &cobra.Command{
		Use:   "{{ .GenerateCommandName }}",
		Short: "{{ .GenerateCommandDescr }}",
		Long: "{{ .GenerateCommandDescr }}",
	}

	generateCmd.Command.PersistentFlags().StringVarP(
		&generateCmd.workloadManifest,
		"workload-manifest",
		"w",
		"",
		"Filepath to the workload manifest to generate child resources for.",
	)

	generateCmd.Command.PersistentFlags().StringVarP(
		&generateCmd.collectionManifest,
		"collection-manifest",
		"c",
		"",
		"Filepath to the workload collection manifest.",
	)

	generateCmd.addCommands()
	c.AddCommand(generateCmd.Command)
}

func (g *generateCommand) addCommands() {
	{{- range $component := .Collection.Spec.Components }}
	g.newGenerate{{ $component.Spec.CompanionCliSubcmd.VarName }}Command()
	{{- end }}
}
`
