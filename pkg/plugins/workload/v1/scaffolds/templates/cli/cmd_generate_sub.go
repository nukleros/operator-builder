package cli

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &CliCmdGenerateSub{}

// CliCmdGenerateSub scaffolds the companion CLI's generate subcommand for the
// workload.  This where the actual generate logic lives.
type CliCmdGenerateSub struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin

	CliRootCmd     string
	CliSubCmdName  string
	CliSubCmdDescr string
	Component      bool
	PackageName    string

	GenerateCommandName  string
	GenerateCommandDescr string
}

func (f *CliCmdGenerateSub) SetTemplateDefaults() error {

	if f.Component {
		f.Path = filepath.Join(
			"cmd", f.CliRootCmd, "commands",
			fmt.Sprintf("%s_generate.go", f.CliSubCmdName),
		)
	} else {
		f.Path = filepath.Join("cmd", f.CliRootCmd, "commands", "generate.go")
	}

	f.GenerateCommandName = generateCommandName
	f.GenerateCommandDescr = generateCommandDescr

	f.TemplateBody = cliCmdGenerateSubTemplate

	return nil
}

var cliCmdGenerateSubTemplate = `{{ .Boilerplate }}

package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	"{{ .Resource.Path }}/{{ .PackageName }}"
)

{{ if not .Component -}}
var workloadManifest string
{{ end }}

// {{ .CliSubCmdName }}GenerateCmd represents the {{ .CliSubCmdName }} generate subcommand
var {{ .CliSubCmdName }}GenerateCmd = &cobra.Command{
	{{ if .Component -}}
	Use:   "{{ .CliSubCmdName }}",
	Short: "{{ .CliSubCmdDescr }}",
	Long: "{{ .CliSubCmdDescr }}",
	{{- else -}}
	Use:   "{{ .GenerateCommandName }}",
	Short: "{{ .GenerateCommandDescr }}",
	Long: "{{ .GenerateCommandDescr }}",
	{{- end }}
	Run: func(cmd *cobra.Command, args []string) {
		filename, _ := filepath.Abs(workloadManifest)
		yamlFile, err := ioutil.ReadFile(filename)
		if err != nil {
			panic(err)
		}

		var workload {{ .Resource.ImportAlias }}.{{ .Resource.Kind }}

		err = yaml.Unmarshal(yamlFile, &workload)
		if err != nil {
			panic(err)
		}

		e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)

		//objects := {{ .Resource.Kind | lower }}.Create()
		//for _, o := range objects {
		//	fmt.Println("---")
		//	err := e.Encode(o.(runtime.Object), os.Stdout)
		//	if err != nil {
		//		panic(err)
		//	}
		//}

		var resourceObjects []metav1.Object
		for _, f := range {{ .PackageName }}.CreateFuncs {
			resource, err := f(&workload)
			if err != nil {
				log.Fatal(err)
			}
			resourceObjects = append(resourceObjects, resource)
		}

		for _, o := range resourceObjects {
			fmt.Println("---")
			err := e.Encode(o.(runtime.Object), os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}
func init() {
	{{ if .Component -}}
	generateCmd.AddCommand({{ .CliSubCmdName }}GenerateCmd)
	{{- else -}}
	rootCmd.AddCommand({{ .CliSubCmdName }}GenerateCmd)

	{{ .CliSubCmdName }}GenerateCmd.Flags().StringVarP(&workloadManifest, "workload-manifest", "w", "", "Filepath to the workload manifest to generate child resources for.")
	{{- end -}}
}
`
