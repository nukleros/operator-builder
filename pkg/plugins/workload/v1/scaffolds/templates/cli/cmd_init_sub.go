package cli

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

var _ machinery.Template = &CliCmdInitSub{}

// CliCmdInitSub scaffolds the companion CLI's init subcommand for the
// workload.  This where the actual init logic lives.
type CliCmdInitSub struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin

	CliRootCmd     string
	CliSubCmdName  string
	CliSubCmdDescr string
	SpecFields     *[]workloadv1.APISpecField
	Component      bool

	InitCommandName  string
	InitCommandDescr string
}

func (f *CliCmdInitSub) SetTemplateDefaults() error {

	if f.Component {
		f.Path = filepath.Join(
			"cmd", f.CliRootCmd, "commands",
			fmt.Sprintf("%s_init.go", f.CliSubCmdName),
		)
	} else {
		f.Path = filepath.Join("cmd", f.CliRootCmd, "commands", "init.go")
	}

	f.InitCommandName = initCommandName
	f.InitCommandDescr = initCommandDescr

	f.TemplateBody = cliCmdInitSubTemplate

	return nil
}

var cliCmdInitSubTemplate = `{{ .Boilerplate }}

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

const defaultManifest = ` + "`" + `apiVersion: {{ .Resource.QualifiedGroup }}/{{ .Resource.Version }}
kind: {{ .Resource.Kind }}
metadata:
  name: {{ lower .Resource.Kind }}-sample
spec:
{{- range .SpecFields }}
  {{ .SampleField -}}
{{ end }}
` + "`" + `

// {{ .CliSubCmdName }}InitCmd represents the {{ .CliSubCmdName }} init subcommand
var {{ .CliSubCmdName }}InitCmd = &cobra.Command{
	{{ if .Component -}}
	Use:   "{{ .CliSubCmdName }}",
	Short: "{{ .CliSubCmdDescr }}",
	Long: "{{ .CliSubCmdDescr }}",
	{{- else -}}
	Use:   "{{ .InitCommandName }}",
	Short: "{{ .InitCommandDescr }}",
	Long: "{{ .InitCommandDescr }}",
	{{- end }}
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(defaultManifest)
	},
}
func init() {
	{{ if .Component -}}
	initCmd.AddCommand({{ .CliSubCmdName }}InitCmd)
	{{- else -}}
	rootCmd.AddCommand({{ .CliSubCmdName }}InitCmd)
	{{- end -}}
}
`
