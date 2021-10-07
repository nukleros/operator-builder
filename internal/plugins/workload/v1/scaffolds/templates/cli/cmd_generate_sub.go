// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
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
	IsCollection      bool
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
	{{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }} "{{ .Repo }}/apis/{{ .Collection.Spec.API.Group }}/{{ .Collection.Spec.API.Version }}"
	{{ end -}}
)

{{ if .IsCollection -}}
type generate{{ .SubCmdVarName }}Command struct {
	*cobra.Command
	collectionManifest string
}
{{- else if .IsComponent -}}
type generate{{ .SubCmdVarName }}Command struct {
	*cobra.Command
	workloadManifest string
	collectionManifest string
}
{{- else }}
type generateCommand struct {
	*cobra.Command
	workloadManifest string
}
{{- end }}

{{ if not .IsComponent -}}
// newGenerateCommand creates a new instance of the generate subcommand.
func (c *{{ .RootCmdVarName }}Command) newGenerateCommand() {
	g := &generateCommand{}
{{- else }}
// newGenerate{{ .SubCmdVarName }}Command creates a new instance of the generate{{ .SubCmdVarName }} subcommand.
func (g *generateCommand) newGenerate{{ .SubCmdVarName }}Command() {
{{- end }}
	{{ if not .IsComponent -}}
	generateCmd := &cobra.Command{
		Use:   "{{ .GenerateCommandName }}",
		Short: "{{ .GenerateCommandDescr }}",
		Long:  "{{ .GenerateCommandDescr }}",
		RunE: g.generate,
	}
	{{- else -}}
	generate{{ .SubCmdVarName }}Cmd := &generate{{ .SubCmdVarName }}Command{}

	generate{{ .SubCmdVarName }}Cmd.Command = &cobra.Command{
		Use:   "{{ .SubCmdName }}",
		Short: "{{ .SubCmdDescr }}",
		Long:  "{{ .SubCmdDescr }}",
		RunE: generate{{ .SubCmdVarName }}Cmd.generate{{ .SubCmdVarName }},
	}
	{{- end }}

	{{ if .IsCollection -}}

	generate{{ .SubCmdVarName }}Cmd.Command.Flags().StringVarP(
		&generate{{ .SubCmdVarName }}Cmd.collectionManifest,
		"collection-manifest",
		"c",
		"",
		"Filepath to the workload collection manifest.",
	)
	generate{{ .SubCmdVarName }}Cmd.MarkFlagRequired("collection-manifest")

	g.AddCommand(generate{{ .SubCmdVarName }}Cmd.Command)

	{{- else if .IsComponent -}}

	generate{{ .SubCmdVarName }}Cmd.Command.Flags().StringVarP(
		&generate{{ .SubCmdVarName }}Cmd.workloadManifest,
		"workload-manifest",
		"w",
		"",
		"Filepath to the workload manifest.",
	)
	generate{{ .SubCmdVarName }}Cmd.MarkFlagRequired("workload-manifest")

	generate{{ .SubCmdVarName }}Cmd.Command.Flags().StringVarP(
		&generate{{ .SubCmdVarName }}Cmd.collectionManifest,
		"collection-manifest",
		"c",
		"",
		"Filepath to the workload collection manifest.",
	)
	generate{{ .SubCmdVarName }}Cmd.MarkFlagRequired("collection-manifest")

	g.AddCommand(generate{{ .SubCmdVarName }}Cmd.Command)

	{{- else -}}

	generate{{ .SubCmdVarName }}Cmd.Flags().StringVarP(
		&g.workloadManifest,
		"workload-manifest",
		"w",
		"",
		"Filepath to the workload manifest to generate child resources for.",
	)
	generate{{ .SubCmdVarName }}Cmd.MarkFlagRequired("workload-manifest")

	c.AddCommand(generate{{ .SubCmdVarName }}Cmd)
	{{- end -}}
}

// generate creates child resource manifests from a workload's custom resource.
{{- if .IsComponent }}
func (g *generate{{ .SubCmdVarName }}Command) generate{{ .SubCmdVarName }}(cmd *cobra.Command, args []string) error {
{{- else }}
func (g *generateCommand) generate(cmd *cobra.Command, args []string) error {
{{- end }}
	{{- if and (.IsComponent) (not .IsCollection) }}
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

	err = validateWorkload(&workload)
	if err != nil {
		return fmt.Errorf("error validating yaml %s, %w", wkFilename, err)
	}
	{{- end }}

	{{ if .IsComponent }}
	// workload collection
	colFilename, _ := filepath.Abs(g.collectionManifest)

	colYamlFile, err := ioutil.ReadFile(colFilename)
	if err != nil {
		return fmt.Errorf("failed to open file %s, %w", colFilename, err)
	}

	var collection {{ $.Collection.Spec.API.Group }}{{ $.Collection.Spec.API.Version }}.{{ $.Collection.Spec.API.Kind }}

	err = yaml.Unmarshal(colYamlFile, &collection)
	if err != nil {
		return fmt.Errorf("failed to unmarshal yaml %s into workload, %w", colFilename, err)
	}

	err = validateWorkload(&collection)
	if err != nil {
		return fmt.Errorf("error validating yaml %s, %w", colFilename, err)
	}

	resourceObjects := make([]metav1.Object, len({{ .PackageName }}.CreateFuncs))

	for i, f := range {{ .PackageName }}.CreateFuncs {
		{{ if .IsCollection }}
		resource, err := f(&collection)
		{{- else }}
		resource, err := f(&workload, &collection)
		{{- end }}
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

	err = validateWorkload(&workload)
	if err != nil {
		return fmt.Errorf("error validating yaml %s, %w", filename, err)
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
`
