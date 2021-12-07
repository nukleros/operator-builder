// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

var _ machinery.Template = &CmdGenerateSub{}
var _ machinery.Inserter = &CmdGenerateSubUpdater{}

// cmdGenerateSubCommon include the common fields that are shared by all generate
// subcommand structs for templating purposes.
type cmdGenerateSubCommon struct {
	RootCmd    workloadv1.CliCommand
	SubCmd     workloadv1.CliCommand
	Collection *workloadv1.WorkloadCollection

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
	Builder           workloadv1.WorkloadAPIBuilder
	ComponentResource *resource.Resource

	// template fields
	cmdGenerateSubCommon
	GenerateCommandName  string
	GenerateCommandDescr string
}

func (f *CmdGenerateSub) SetTemplateDefaults() error {
	if f.Builder.IsComponent() {
		f.Resource = f.ComponentResource
	}

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
		machinery.NewMarkerFor(f.Path, generateSwitchMarker),
		machinery.NewMarkerFor(f.Path, generateFuncMarker),
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
	Builder           workloadv1.WorkloadAPIBuilder
	ComponentResource *resource.Resource

	// template fields
	cmdGenerateSubCommon
	PackageName string
}

// GetPath implements file.Builder interface.
func (f *CmdGenerateSubUpdater) GetPath() string {
	if f.Builder.IsComponent() {
		f.Resource = f.ComponentResource
	}

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
const generateFuncMarker = "operator-builder:generatefunc"
const generateSwitchMarker = "operator-builder:generateswitch"

// GetMarkers implements file.Inserter interface.
func (f *CmdGenerateSubUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.GetPath(), generateImportMarker),
		machinery.NewMarkerFor(f.GetPath(), generateSwitchMarker),
		machinery.NewMarkerFor(f.GetPath(), generateFuncMarker),
	}
}

// Code Fragments.
const (
	// this fragment is the function that a standalone workload will use to generate
	// its child resources.
	generateFuncStandaloneFragment = `// %sGenerate%s returns the child resources that are associated
// with this %s workload.
func %sGenerate%s(yamlFile []byte) ([]client.Object, error) {
	var workload %s.%s

	if err := yaml.Unmarshal(yamlFile, &workload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml into workload, %%w", err)
	}

	if err := cmdutils.ValidateWorkload(&workload); err != nil {
		return nil, fmt.Errorf("error validating workload yaml, %%w", err)
	}

	resourceObjects := make([]client.Object, len(%s.CreateFuncs))

	for i, f := range %s.CreateFuncs {
		resource, err := f(&workload)
		if err != nil {
			return nil, err
		}

		resourceObjects[i] = resource
	}

	return resourceObjects, nil
}
`

	// this fragment is the function that a workload with a collection will use to generate
	// its child resources.
	generateFuncWithCollectionFragment = `// %sGenerate%s returns the child resources that are associated
// with this %s workload.
func %sGenerate%s(workloadFile, collectionFile []byte) ([]client.Object, error) {
	var workload %s.%s
	var collection %s.%s

	if err := yaml.Unmarshal(workloadFile, &workload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml into workload, %%w", err)
	}

	if err := yaml.Unmarshal(collectionFile, &collection); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml into workload, %%w", err)
	}

	if err := cmdutils.ValidateWorkload(&workload); err != nil {
		return nil, fmt.Errorf("error validating workload yaml, %%w", err)
	}

	if err := cmdutils.ValidateWorkload(&collection); err != nil {
		return nil, fmt.Errorf("error validating workload yaml, %%w", err)
	}

	resourceObjects := make([]client.Object, len(%s.CreateFuncs))

	for i, f := range %s.CreateFuncs {
		resource, err := f(&workload, &collection)
		if err != nil {
			return nil, err
		}

		resourceObjects[i] = resource
	}

	return resourceObjects, nil
}
`

	// this fragment is the imports which are created and updated upon each new
	// api version that is created.
	generateImportFragment = `%s "%s"
	%s "%s/%s"
`

	// this fragment switches between each new api version that is created in order
	// to generate that specific api versions child resources.
	generateSwitchesFragment = `case "%s":
	resourceObjects, err = %sGenerate%s(%s)
`
)

// GetCodeFragments implements file.Inserter interface.
func (f *CmdGenerateSubUpdater) GetCodeFragments() machinery.CodeFragmentsMap {
	fragments := make(machinery.CodeFragmentsMap, 1)

	if f.Builder.IsComponent() {
		f.Resource = f.ComponentResource
	}

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
	funcs := make([]string, 0)

	// set some common variables
	versionPath := fmt.Sprintf("%s/apis/%s/%s", f.Repo, f.Resource.Group, f.Resource.Version)
	groupVersion := f.Resource.Group + f.Resource.Version
	packageVersion := f.PackageName + f.Resource.Version

	// add the imports fragment
	imports = append(imports, fmt.Sprintf(generateImportFragment,
		groupVersion, versionPath,
		packageVersion, versionPath, f.PackageName))

	// determine the input string to the generated function
	var funcInputs string

	switch {
	case f.UseCollectionManifestFlag && f.UseWorkloadManifestFlag:
		funcInputs = "workloadFile, collectionFile"
	case f.UseCollectionManifestFlag && !f.UseWorkloadManifestFlag:
		funcInputs = "collectionFile"
	default:
		funcInputs = "workloadFile"
	}

	// add the switches fragment
	switches = append(switches, fmt.Sprintf(generateSwitchesFragment,
		f.Resource.Version, f.Resource.Version, f.Resource.Kind, funcInputs))

	// add the function fragment
	if f.Builder.IsStandalone() || f.Builder.IsCollection() {
		funcs = append(funcs, fmt.Sprintf(generateFuncStandaloneFragment,
			// function comments
			f.Resource.Version, f.Resource.Kind, f.Resource.Kind,

			// function body
			f.Resource.Version, f.Resource.Kind, groupVersion, f.Resource.Kind, packageVersion,

			// function loop
			packageVersion,
		))
	} else {
		funcs = append(funcs, fmt.Sprintf(generateFuncWithCollectionFragment,
			// function comments
			f.Resource.Version, f.Resource.Kind, f.Resource.Kind,

			// function body
			f.Resource.Version, f.Resource.Kind, groupVersion, f.Resource.Kind,
			f.Collection.Spec.API.Group+f.Collection.Spec.API.Version, f.Collection.Spec.API.Kind, packageVersion,

			// function loop
			packageVersion,
		))
	}

	// Only store code fragments in the map if the slices are non-empty
	if len(imports) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), generateImportMarker)] = imports
	}

	if len(switches) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), generateSwitchMarker)] = switches
	}

	if len(funcs) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), generateFuncMarker)] = funcs
	}

	return fragments
}

const (
	//nolint: lll
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
	"sigs.k8s.io/yaml"

	// common imports for subcommands
	cmdgenerate "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/generate"
	cmdutils "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/utils"

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
		UseCollectionManifest: {{ .UseCollectionManifestFlag }},
		UseWorkloadManifest:   {{ .UseWorkloadManifestFlag }},
		SubCommandOf:          parentCommand,
		GenerateFunc:          Generate{{ .Resource.Kind }},
	}

	generateCmd.Setup()
}

// Generate{{ .Resource.Kind }} runs the logic to generate child resources for a
// {{ .Resource.Kind }} workload.
func Generate{{ .Resource.Kind }}(g *cmdgenerate.GenerateSubCommand) error {
	var resourceObjects []client.Object
	var apiVersion string

	{{ if .UseWorkloadManifestFlag }}
	workloadFilename, _ := filepath.Abs(g.WorkloadManifest)
	workloadFile, err := os.ReadFile(workloadFilename)
	if err != nil {
		return fmt.Errorf("failed to open workload file %%s, %%w", workloadFile, err)
	}

	var workload interface{}

	if err := yaml.Unmarshal(workloadFile, &workload); err != nil {
		return fmt.Errorf("failed to unmarshal yaml into workload, %%w", err)
	}

	workloadGroupVersion := strings.Split(workload.(map[string]interface{})["apiVersion"].(string), "/")
	workloadAPIVersion := workloadGroupVersion[len(workloadGroupVersion)-1]

	apiVersion = workloadAPIVersion
	{{ end }}

	{{ if .UseCollectionManifestFlag }}
	collectionFilename, _ := filepath.Abs(g.CollectionManifest)
	collectionFile, err := os.ReadFile(collectionFilename)
	if err != nil {
		return fmt.Errorf("failed to open collection file %%s, %%w", collectionFile, err)
	}

	var collection interface{}

	if err := yaml.Unmarshal(collectionFile, &collection); err != nil {
		return fmt.Errorf("failed to unmarshal yaml into collection, %%w", err)
	}

	collectionGroupVersion := strings.Split(collection.(map[string]interface{})["apiVersion"].(string), "/")
	collectionAPIVersion := collectionGroupVersion[len(collectionGroupVersion)-1]

	apiVersion = collectionAPIVersion
	{{ end }}

	switch apiVersion {
	default:
		return fmt.Errorf("unsupported API Version: " + apiVersion)
	%s
	}

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

%s
`
)
