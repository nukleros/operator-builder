package cli

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

var _ machinery.Template = &CliCmdGenerateSub{}

// CliCmdGenerateSub scaffolds the companion CLI's generate subcommand for the
// workload.  This where the actual generate logic lives.
type CliCmdGenerateSub struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	PackageName       string
	CliRootCmd        string
	CliSubCmdName     string
	CliSubCmdDescr    string
	CliSubCmdVarName  string
	CliSubCmdFileName string
	IsComponent       bool
	ComponentResource *resource.Resource
	Collection        *workloadv1.WorkloadCollection

	GenerateCommandName  string
	GenerateCommandDescr string
}

func (f *CliCmdGenerateSub) SetTemplateDefaults() error {
	if f.IsComponent {
		f.Path = filepath.Join(
			"cmd", f.CliRootCmd, "commands",
			fmt.Sprintf("%s_generate.go", f.CliSubCmdFileName),
		)
		f.Resource = f.ComponentResource
	} else {
		f.Path = filepath.Join("cmd", f.CliRootCmd, "commands", "generate.go")
	}

	f.GenerateCommandName = generateCommandName
	f.GenerateCommandDescr = generateCommandDescr

	f.TemplateBody = cliCmdGenerateSubTemplate

	return nil
}

//nolint: lll
const cliCmdGenerateSubTemplate = `{{ .Boilerplate }}

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
	{{- if .IsComponent }}
	{{ .Collection.Spec.APIGroup }}{{ .Collection.Spec.APIVersion }} "{{ .Repo }}/apis/{{ .Collection.Spec.APIGroup }}/{{ .Collection.Spec.APIVersion }}"
	{{ end -}}
)

{{ if not .IsComponent -}}
var workloadManifest string
{{ end }}

// {{ .CliSubCmdVarName }}GenerateCmd represents the {{ .CliSubCmdName }} generate subcommand.
var {{ .CliSubCmdVarName }}GenerateCmd = &cobra.Command{
	{{ if .IsComponent -}}
	Use:   "{{ .CliSubCmdName }}",
	Short: "{{ .CliSubCmdDescr }}",
	Long: "{{ .CliSubCmdDescr }}",
	{{- else -}}
	Use:   "{{ .GenerateCommandName }}",
	Short: "{{ .GenerateCommandDescr }}",
	Long: "{{ .GenerateCommandDescr }}",
	{{- end }}
	Run: func(cmd *cobra.Command, args []string) {
		{{ if .IsComponent }}
		// component workload
		wkFilename, _ := filepath.Abs(workloadManifest)
		wkYamlFile, err := ioutil.ReadFile(wkFilename)
		if err != nil {
			panic(err)
		}

		var workload {{ .Resource.ImportAlias }}.{{ .Resource.Kind }}

		err = yaml.Unmarshal(wkYamlFile, &workload)
		if err != nil {
			panic(err)
		}

		// workload collection
		colFilename, _ := filepath.Abs(collectionManifest)
		colYamlFile, err := ioutil.ReadFile(colFilename)
		if err != nil {
			panic(err)
		}

		var collection {{ $.Collection.Spec.APIGroup }}{{ $.Collection.Spec.APIVersion }}.{{ $.Collection.Spec.APIKind }}

		err = yaml.Unmarshal(colYamlFile, &collection)
		if err != nil {
			panic(err)
		}

		var resourceObjects []metav1.Object
		for _, f := range {{ .PackageName }}.CreateFuncs {
			resource, err := f(&workload, &collection)
			if err != nil {
				log.Fatal(err)
			}
			resourceObjects = append(resourceObjects, resource)
		}
		{{ else }}
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

		var resourceObjects []metav1.Object
		for _, f := range {{ .PackageName }}.CreateFuncs {
			resource, err := f(&workload)
			if err != nil {
				log.Fatal(err)
			}
			resourceObjects = append(resourceObjects, resource)
		}
		{{ end }}

		e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
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
	{{ if .IsComponent -}}
	generateCmd.AddCommand({{ .CliSubCmdVarName }}GenerateCmd)
	{{- else -}}
	rootCmd.AddCommand({{ .CliSubCmdVarName }}GenerateCmd)

	{{ .CliSubCmdVarName }}GenerateCmd.Flags().StringVarP(
		&workloadManifest,
		"workload-manifest",
		"w",
		"",
		"Filepath to the workload manifest to generate child resources for.",
	)
	{{- end -}}
}
`
