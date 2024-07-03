// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"

	"github.com/nukleros/operator-builder/internal/utils"
	"github.com/nukleros/operator-builder/internal/workload/v1/commands/companion"
	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

var _ machinery.Template = &CmdGenerateSub{}
var _ machinery.Inserter = &CmdGenerateSubUpdater{}

// cmdGenerateSubCommon include the common fields that are shared by all generate
// subcommand structs for templating purposes.
type cmdGenerateSubCommon struct {
	RootCmd    companion.CLI
	SubCmd     companion.CLI
	Collection *kinds.WorkloadCollection

	UseCollectionManifestFlag bool
	UseWorkloadManifestFlag   bool
}

// CmdGenerateSub scaffolds the companion CLI's generate subcommand for the
// workload.  This where the actual generate logic lives.
type CmdGenerateSub struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	// input fields
	Builder kinds.WorkloadBuilder

	// template fields
	cmdGenerateSubCommon
	GenerateCommandName  string
	GenerateCommandDescr string
	GenerateFuncInputs   string
}

func (f *CmdGenerateSub) SetTemplateDefaults() error {
	// set template fields
	f.RootCmd = *f.Builder.GetRootCommand()
	f.SubCmd = *f.Builder.GetSubCommand()
	f.Collection = f.Builder.GetCollection()

	// if we have a standalone simply use the default command name and description
	// for generate since the 'generate' command will be the last in the chain,
	// otherwise we will use the requested subcommand name.
	if f.Builder.IsStandalone() {
		f.GenerateCommandName = generateCommandName
		f.GenerateCommandDescr = generateCommandDescr
	} else {
		f.GenerateCommandName = f.SubCmd.Name
		f.GenerateCommandDescr = f.SubCmd.Description
		f.UseCollectionManifestFlag = true
	}

	// use the workload manifest flag for non-collection use cases
	if !f.Builder.IsCollection() {
		f.UseWorkloadManifestFlag = true
	}

	// determine the input string to the generated function
	switch {
	case f.UseCollectionManifestFlag && f.UseWorkloadManifestFlag:
		f.GenerateFuncInputs = "workloadFile, collectionFile"
	case f.UseCollectionManifestFlag && !f.UseWorkloadManifestFlag:
		f.GenerateFuncInputs = "collectionFile"
	default:
		f.GenerateFuncInputs = "workloadFile"
	}

	// set interface fields
	f.Path = f.SubCmd.GetSubCmdRelativeFileName(
		f.RootCmd.Name,
		"generate",
		f.Resource.Group,
		utils.ToFileName(f.Resource.Kind),
	)

	f.TemplateBody = fmt.Sprintf(
		cmdGenerateSub,
		machinery.NewMarkerFor(f.Path, generateImportMarker),
		machinery.NewMarkerFor(f.Path, generateVersionMapMarker),
	)

	return nil
}

// CmdGenerateSubUpdater updates a specific components version subcommand with
// appropriate initialization information.
type CmdGenerateSubUpdater struct { //nolint:maligned
	machinery.RepositoryMixin
	machinery.MultiGroupMixin
	machinery.ResourceMixin

	// input fields
	Builder kinds.WorkloadBuilder

	// template fields
	cmdGenerateSubCommon
	PackageName string
}

// GetPath implements file.Builder interface.
func (f *CmdGenerateSubUpdater) GetPath() string {
	return f.SubCmd.GetSubCmdRelativeFileName(
		f.Builder.GetRootCommand().Name,
		"generate",
		f.Resource.Group,
		utils.ToFileName(f.Resource.Kind),
	)
}

// GetIfExistsAction implements file.Builder interface.
func (*CmdGenerateSubUpdater) GetIfExistsAction() machinery.IfExistsAction {
	return machinery.OverwriteFile
}

const generateImportMarker = "operator-builder:imports"
const generateVersionMapMarker = "operator-builder:versionmap"

// GetMarkers implements file.Inserter interface.
func (f *CmdGenerateSubUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.GetPath(), generateImportMarker),
		machinery.NewMarkerFor(f.GetPath(), generateVersionMapMarker),
	}
}

// Code Fragments.
const (
	// this fragment is the imports which are created and updated upon each new
	// api version that is created.
	generateImportFragment = `%s%s "%s"
`
	// generateMapFragment is a fragment which provides a new switch for each api version
	// that is created for use by the api-version flag.
	generateMapFragment = `"%s": %s%s.GenerateForCLI,
`
)

// GetCodeFragments implements file.Inserter interface.
func (f *CmdGenerateSubUpdater) GetCodeFragments() machinery.CodeFragmentsMap {
	fragments := make(machinery.CodeFragmentsMap, 1)

	// set template fields
	f.RootCmd = *f.Builder.GetRootCommand()
	f.SubCmd = *f.Builder.GetSubCommand()
	f.PackageName = f.Builder.GetPackageName()
	f.Collection = f.Builder.GetCollection()

	// If resource is not being provided we are creating the file, not updating it
	if f.Resource == nil {
		return fragments
	}

	// use the collection flag for non-standalone use cases
	if !f.Builder.IsStandalone() {
		f.UseCollectionManifestFlag = true
	}

	// use the workload manifest flag for non-collection use cases
	if !f.Builder.IsCollection() {
		f.UseWorkloadManifestFlag = true
	}

	// Generate subCommands code fragments
	imports := make([]string, 0)
	switches := make([]string, 0)

	// add the imports
	imports = append(imports, fmt.Sprintf(generateImportFragment,
		f.Resource.Version,
		strings.ToLower(f.Resource.Kind),
		fmt.Sprintf("%s/%s", f.Resource.Path, f.Builder.GetPackageName()),
	))

	// add the switches fragment
	switches = append(switches, fmt.Sprintf(generateMapFragment,
		f.Resource.Version, f.Resource.Version, strings.ToLower(f.Resource.Kind)))

	// Only store code fragments in the map if the slices are non-empty
	if len(imports) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), generateImportMarker)] = imports
	}

	if len(switches) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), generateVersionMapMarker)] = switches
	}

	return fragments
}

const (
	cmdGenerateSub = `{{ .Boilerplate }}

package {{ .Resource.Group }}

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// common imports for subcommands
	cmdgenerate "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/generate"

	// specific imports for workloads
	{{- if .Builder.IsComponent }}
	{{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }} "{{ .Repo }}/apis/{{ .Collection.Spec.API.Group }}/{{ .Collection.Spec.API.Version }}"
	{{ end }}
	%s
)

// New{{ .Resource.Kind }}SubCommand creates a new command and adds it to its 
// parent command.
func New{{ .Resource.Kind }}SubCommand(parentCommand *cobra.Command) {
	generateCmd := &cmdgenerate.GenerateSubCommand{
		Name:                  "{{ .GenerateCommandName }}",
		Description:           "{{ .GenerateCommandDescr }}",
		SubCommandOf:          parentCommand,
		GenerateFunc:          Generate{{ .Resource.Kind }},
		{{- if .UseCollectionManifestFlag }}
		UseCollectionManifest: true,
		{{- if .Builder.IsCollection }}
		CollectionKind:        "{{ .Resource.Kind }}",
		{{- else }}
		CollectionKind:        "{{ .Collection.Spec.API.Kind }}",
		{{ end -}}
		{{ end -}}
		{{ if .UseWorkloadManifestFlag -}}
		UseWorkloadManifest:   true,
		WorkloadKind:          "{{ .Resource.Kind }}",
		{{ end -}}
	}

	generateCmd.Setup()
}

// Generate{{ .Resource.Kind }} runs the logic to generate child resources for a
// {{ .Resource.Kind }} workload.
func Generate{{ .Resource.Kind }}(g *cmdgenerate.GenerateSubCommand) error {
	var apiVersion string

	{{ if .UseWorkloadManifestFlag }}
	workloadFilename, _ := filepath.Abs(g.WorkloadManifest)
	workloadFile, err := os.ReadFile(workloadFilename)
	if err != nil {
		return fmt.Errorf("failed to open workload file %%s, %%w", workloadFile, err)
	}

	var workload map[string]interface{}

	if err := yaml.Unmarshal(workloadFile, &workload); err != nil {
		return fmt.Errorf("failed to unmarshal yaml into workload, %%w", err)
	}

	workloadGroupVersion := strings.Split(workload["apiVersion"].(string), "/")
	workloadAPIVersion := workloadGroupVersion[len(workloadGroupVersion)-1]

	apiVersion = workloadAPIVersion
	{{ end }}

	{{ if .UseCollectionManifestFlag }}
	collectionFilename, _ := filepath.Abs(g.CollectionManifest)
	collectionFile, err := os.ReadFile(collectionFilename)
	if err != nil {
		return fmt.Errorf("failed to open collection file %%s, %%w", collectionFile, err)
	}

	var collection map[string]interface{}

	if err := yaml.Unmarshal(collectionFile, &collection); err != nil {
		return fmt.Errorf("failed to unmarshal yaml into collection, %%w", err)
	}

	collectionGroupVersion := strings.Split(collection["apiVersion"].(string), "/")
	collectionAPIVersion := collectionGroupVersion[len(collectionGroupVersion)-1]

	apiVersion = collectionAPIVersion
	{{ end }}

	// generate a map of all versions to generate functions for each api version created
	{{- if .Builder.IsComponent }}
	type generateFunc func([]byte, []byte) ([]client.Object, error)
	{{ else }}
	type generateFunc func([]byte) ([]client.Object, error)
	{{ end -}}
	generateFuncMap := map[string]generateFunc{
		%s
	}

	generate := generateFuncMap[apiVersion]
	resourceObjects, err := generate({{ .GenerateFuncInputs }})
	if err != nil {
		return fmt.Errorf("unable to retrieve resources; %%w", err)
	}

	e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)

	outputStream := os.Stdout

	for _, o := range resourceObjects {
		if _, err := outputStream.WriteString("---\n"); err != nil {
			return fmt.Errorf("failed to write output, %%w", err)
		}

		if err := e.Encode(o, os.Stdout); err != nil {
			return fmt.Errorf("failed to write output, %%w", err)
		}
	}

	return nil
}
`
)
