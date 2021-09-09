package cli

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

var _ machinery.Template = &CmdInitSub{}

// CmdInitSub scaffolds the companion CLI's init subcommand for the
// workload.  This where the actual init logic lives.
type CmdInitSub struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin

	RootCmd           string
	RootCmdVarName    string
	SubCmdName        string
	SubCmdDescr       string
	SubCmdVarName     string
	SubCmdFileName    string
	SpecFields        []*workloadv1.APISpecField
	IsComponent       bool
	ComponentResource *resource.Resource

	InitCommandName  string
	InitCommandDescr string
}

func (f *CmdInitSub) SetTemplateDefaults() error {
	if f.IsComponent {
		f.Path = filepath.Join(
			"cmd", f.RootCmd, "commands",
			fmt.Sprintf("%s_init.go", f.SubCmdFileName),
		)
		f.Resource = f.ComponentResource
	} else {
		f.Path = filepath.Join("cmd", f.RootCmd, "commands", "init.go")
	}

	f.InitCommandName = initCommandName
	f.InitCommandDescr = initCommandDescr

	f.TemplateBody = cliCmdInitSubTemplate

	return nil
}

const cliCmdInitSubTemplate = `{{ .Boilerplate }}

package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const defaultManifest{{ .SubCmdVarName }} = ` + "`" + `apiVersion: {{ .Resource.QualifiedGroup }}/{{ .Resource.Version }}
kind: {{ .Resource.Kind }}
metadata:
  name: {{ lower .Resource.Kind }}-sample
spec:
{{- range .SpecFields }}
  {{ .SampleField -}}
{{ end }}
` + "`" + `

{{ if not .IsComponent -}}
// newInitCommand creates a new instance of the init subcommand.
func (c *{{ .RootCmdVarName }}Command) newInitCommand() {
{{- else }}
// newInit{{ .SubCmdVarName }}Command creates a new instance of the  init{{ .SubCmdVarName }} subcommand.
func (i *initCommand) newInit{{ .SubCmdVarName }}Command() {
{{- end }}
	init{{ .SubCmdVarName }}Cmd := &cobra.Command{
		{{ if .IsComponent -}}
		Use:   "{{ .SubCmdName }}",
		Short: "{{ .SubCmdDescr }}",
		Long: "{{ .SubCmdDescr }}",
		{{- else -}}
		Use:   "{{ .InitCommandName }}",
		Short: "{{ .InitCommandDescr }}",
		Long: "{{ .InitCommandDescr }}",
		{{- end }}
		RunE: func(cmd *cobra.Command, args []string) error {
			outputStream := os.Stdout

			if _, err := outputStream.WriteString(defaultManifest{{ .SubCmdVarName }}); err != nil {
				return fmt.Errorf("failed to write outout, %w", err)
			}

			return nil
		},
	}

	{{ if .IsComponent -}}
	i.AddCommand(init{{ .SubCmdVarName }}Cmd)
	{{- else -}}
	c.AddCommand(init{{ .SubCmdVarName }}Cmd)
	{{- end -}}
}
`
