package cli

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

var _ machinery.Template = &CmdGenerateSub{}

// CmdGenerateSub scaffolds the companion CLI's generate subcommand for the
// workload.  This where the actual generate logic lives.
type CmdGenerateSub struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	PackageName       string
	RootCmd           string
	RootCmdVarName    string
	SubCmdName        string
	SubCmdDescr       string
	SubCmdVarName     string
	SubCmdFileName    string
	IsComponent       bool
	ComponentResource *resource.Resource
	Collection        *workloadv1.WorkloadCollection

	GenerateCommandName  string
	GenerateCommandDescr string
}

func (f *CmdGenerateSub) SetTemplateDefaults() error {
	if f.IsComponent {
		f.Path = filepath.Join(
			"cmd", f.RootCmd, "commands",
			fmt.Sprintf("%s_generate.go", f.SubCmdFileName),
		)
		f.Resource = f.ComponentResource
	} else {
		f.Path = filepath.Join("cmd", f.RootCmd, "commands", "generate.go")
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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/yaml"

	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	"{{ .Resource.Path }}/{{ .PackageName }}"
	{{- if .IsComponent }}
	{{ .Collection.Spec.APIGroup }}{{ .Collection.Spec.APIVersion }} "{{ .Repo }}/apis/{{ .Collection.Spec.APIGroup }}/{{ .Collection.Spec.APIVersion }}"
	{{ end -}}
)

{{ if not .IsComponent -}}
type generateCommand struct {
	*cobra.Command
	workloadManifest string
}
{{- end }}

// generate creates child resource manifests from a workload's custom resource.
func (g *generateCommand) generate{{ .SubCmdVarName }}(cmd *cobra.Command, args []string) error {
	{{- if .IsComponent }}
	// component workload
	wkFilename, _ := filepath.Abs(g.workloadManifest)

	wkYamlFile, err := ioutil.ReadFile(wkFilename)
	if err != nil {
		return fmt.Errorf("failed to open file %s, %w", wkFilename, err)
	}

	var workload {{ .Resource.ImportAlias }}.{{ .Resource.Kind }}

	err = yaml.Unmarshal(wkYamlFile, &workload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal yaml %s into workload, %w", wkFilename, err)
	}

	// workload collection
	colFilename, _ := filepath.Abs(g.collectionManifest)

	colYamlFile, err := ioutil.ReadFile(colFilename)
	if err != nil {
		return fmt.Errorf("failed to open file %s, %w", colFilename, err)
	}

	var collection {{ $.Collection.Spec.APIGroup }}{{ $.Collection.Spec.APIVersion }}.{{ $.Collection.Spec.APIKind }}

	err = yaml.Unmarshal(colYamlFile, &collection)
	if err != nil {
		return fmt.Errorf("failed to unmarshal yaml %s into workload, %w", colFilename, err)
	}

	resourceObjects := make([]metav1.Object, len({{ .PackageName }}.CreateFuncs))
	
	for i, f := range {{ .PackageName }}.CreateFuncs {
		resource, err := f(&workload, &collection)
		if err != nil {
			return err
		}
		
		resourceObjects[i] = resource
	}
	{{ else }}
	filename, _ := filepath.Abs(g.workloadManifest)

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s, %w", filename, err)
	}

	var workload {{ .Resource.ImportAlias }}.{{ .Resource.Kind }}

	err = yaml.Unmarshal(yamlFile, &workload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal yaml %s into workload, %w", filename, err)
	}

	resourceObjects := make([]metav1.Object, len({{ .PackageName }}.CreateFuncs))

	for i, f := range {{ .PackageName }}.CreateFuncs {
		resource, err := f(&workload)
		if err != nil {
			return err
		}

		resourceObjects[i] = resource
	}
	{{ end }}

	e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)

	outputStream := os.Stdout

	for _, o := range resourceObjects {
		if _, err := outputStream.WriteString("---\n"); err != nil {
			return fmt.Errorf("failed to write output, %w", err)
		}

		if err := e.Encode(o.(runtime.Object), os.Stdout); err != nil {
			return fmt.Errorf("failed to write output, %w", err)
		}
	}

	return nil
}


{{ if not .IsComponent -}}
// newGenerateCommand creates a new instance of the generate subcommand.
func (c *{{ .RootCmdVarName }}Command) newGenerateCommand() {
	g := &generateCommand{}
{{- else }}
// newGenerate{{ .SubCmdVarName }}Command creates a new instance of the generaete{{ .SubCmdVarName }} subcommand.
func (g *generateCommand) newGenerate{{ .SubCmdVarName }}Command() {
{{- end }}
	generate{{ .SubCmdVarName }}Cmd := &cobra.Command{
		{{ if .IsComponent -}}
		Use:   "{{ .SubCmdName }}",
		Short: "{{ .SubCmdDescr }}",
		Long: "{{ .SubCmdDescr }}",
		{{- else -}}
		Use:   "{{ .GenerateCommandName }}",
		Short: "{{ .GenerateCommandDescr }}",
		Long: "{{ .GenerateCommandDescr }}",
		{{- end }}
		RunE: g.generate{{ .SubCmdVarName }},
	}

	{{ if .IsComponent -}}
	g.AddCommand(generate{{ .SubCmdVarName }}Cmd)
	{{- else -}}

	generate{{ .SubCmdVarName }}Cmd.Flags().StringVarP(
		&g.workloadManifest,
		"workload-manifest",
		"w",
		"",
		"Filepath to the workload manifest to generate child resources for.",
	)

	c.AddCommand(generate{{ .SubCmdVarName }}Cmd)
	{{- end -}}
}
`
