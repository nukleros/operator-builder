// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

var _ machinery.Template = &CmdInitSub{}
var _ machinery.Template = &CmdInitSubLatest{}
var _ machinery.Inserter = &CmdInitSubUpdater{}

// cmdInitSubCommon include the common fields that are shared by all init
// subcommand structs for templating purposes.
type cmdInitSubCommon struct {
	RootCmd workloadv1.CliCommand
	SubCmd  workloadv1.CliCommand
}

// CmdInitSub scaffolds the companion CLI's init subcommand for the
// workload.  This where the actual init logic lives.
type CmdInitSub struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin
	machinery.RepositoryMixin

	// input fields
	Builder           workloadv1.WorkloadAPIBuilder
	ComponentResource *resource.Resource

	// template fields
	cmdInitSubCommon
	InitCommandName  string
	InitCommandDescr string
}

func (f *CmdInitSub) SetTemplateDefaults() error {
	if f.Builder.IsComponent() {
		f.Resource = f.ComponentResource
	}

	// set template fields
	f.RootCmd = *f.Builder.GetRootCommand()
	f.SubCmd = *f.Builder.GetSubCommand()

	if f.Builder.IsStandalone() {
		f.InitCommandName = initCommandName
		f.InitCommandDescr = initCommandDescr
	} else {
		f.InitCommandName = f.SubCmd.Name
		f.InitCommandDescr = f.SubCmd.Description
	}

	// set interface fields
	f.Path = f.SubCmd.GetSubCmdRelativeFileName(
		f.RootCmd.Name,
		"init",
		f.Resource.Group,
		utils.ToFileName(f.Resource.Kind),
	)

	f.TemplateBody = fmt.Sprintf(
		cmdInitSub,
		machinery.NewMarkerFor(f.Path, initImportsMarker),
		machinery.NewMarkerFor(f.Path, initVersionMapMarker),
	)

	return nil
}

// CmdInitSubLatest scaffolds the companion CLI's init subcommand logic
// for the latest API that was created.
type CmdInitSubLatest struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin

	// input fields
	Builder           workloadv1.WorkloadAPIBuilder
	ComponentResource *resource.Resource

	// template fields
	cmdInitSubCommon
	PackageName string
}

func (f *CmdInitSubLatest) SetTemplateDefaults() error {
	if f.Builder.IsComponent() {
		f.Resource = f.ComponentResource
	}

	// set template fields
	f.RootCmd = *f.Builder.GetRootCommand()
	f.SubCmd = *f.Builder.GetSubCommand()
	f.PackageName = f.Builder.GetPackageName()

	// set interface fields
	f.Path = f.SubCmd.GetSubCmdRelativeFileName(
		f.RootCmd.Name,
		"init",
		f.Resource.Group,
		utils.ToFileName(f.Resource.Kind+"_latest"),
	)
	f.TemplateBody = cmdInitSubLatest

	// always overwrite the file to ensure we update this with the latest
	// version as we generate them.
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

// CmdInitSubUpdater updates a specific components init subcommand with
// appropriate initialization information.
type CmdInitSubUpdater struct { //nolint:maligned
	machinery.RepositoryMixin
	machinery.MultiGroupMixin
	machinery.ResourceMixin

	// input fields
	Builder           workloadv1.WorkloadAPIBuilder
	ComponentResource *resource.Resource

	// template fields
	cmdInitSubCommon
}

// GetPath implements file.Builder interface.
func (f *CmdInitSubUpdater) GetPath() string {
	if f.Builder.IsComponent() {
		f.Resource = f.ComponentResource
	}

	return f.SubCmd.GetSubCmdRelativeFileName(
		f.Builder.GetRootCommand().Name,
		"init",
		f.Resource.Group,
		utils.ToFileName(f.Resource.Kind),
	)
}

// GetIfExistsAction implements file.Builder interface.
func (*CmdInitSubUpdater) GetIfExistsAction() machinery.IfExistsAction {
	return machinery.OverwriteFile
}

const initImportsMarker = "operator-builder:imports"
const initVersionMapMarker = "operator-builder:versionmap"

// GetMarkers implements file.Inserter interface.
func (f *CmdInitSubUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.GetPath(), initImportsMarker),
		machinery.NewMarkerFor(f.GetPath(), initVersionMapMarker),
	}
}

// Code Fragments.
const (
	// initImportsFragment is a fragment which provides the package to import
	// for the workload.
	initImportsFragment = `%s%s "%s"
`

	// initSwitchFragment is a fragment which provides a new switch for each api version
	// that is created for use by the api-version flag.
	initVersionMapFragment = `"%s": %s,
`
)

// GetCodeFragments implements file.Inserter interface.
func (f *CmdInitSubUpdater) GetCodeFragments() machinery.CodeFragmentsMap {
	fragments := make(machinery.CodeFragmentsMap, 1)

	if f.Builder.IsComponent() {
		f.Resource = f.ComponentResource
	}

	// set template fields
	f.RootCmd = *f.Builder.GetRootCommand()
	f.SubCmd = *f.Builder.GetSubCommand()

	// If resource is not being provided we are creating the file, not updating it
	if f.Resource == nil {
		return fragments
	}

	// Generate subCommands code fragments
	imports := make([]string, 0)
	switches := make([]string, 0)

	// add the imports
	imports = append(imports, fmt.Sprintf(initImportsFragment,
		f.Resource.Version,
		strings.ToLower(f.Resource.Kind),
		fmt.Sprintf("%s/%s", f.Resource.Path, f.Builder.GetPackageName()),
	))

	// add the switches
	switches = append(switches, fmt.Sprintf(initVersionMapFragment,
		f.Resource.Version,
		fmt.Sprintf("%s%s.Sample()",
			f.Resource.Version,
			strings.ToLower(f.Resource.Kind),
		)),
	)

	// Only store code fragments in the map if the slices are non-empty
	if len(imports) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), initImportsMarker)] = imports
	}

	if len(switches) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), initVersionMapMarker)] = switches
	}

	return fragments
}

const (
	// cmdInitSub scaffolds the CLI subcommand logic for an individual component.
	cmdInitSub = `{{ .Boilerplate }}

package {{ .Resource.Group }}

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmdinit "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/init"
	%s
)

// get{{ .Resource.Kind }}Manifest returns the sample {{ .Resource.Kind }} manifest
// based upon API Version input.
func get{{ .Resource.Kind }}Manifest(i *cmdinit.InitSubCommand) (string, error) {
	apiVersion := i.APIVersion
	if apiVersion == "" || apiVersion == "latest" {
		return latest{{ .Resource.Kind }}, nil
	}

	// generate a map of all versions to samples for each api version created
	manifestMap := map[string]string{
		%s
	}

	// return the manifest if it is not blank
	manifest := manifestMap[apiVersion]
	if manifest != "" {
		return manifest, nil
	}

	// return an error if we did not find a manifest for an api version
	return "", fmt.Errorf("unsupported API Version: " + apiVersion)
}

// New{{ .Resource.Kind }}SubCommand creates a new command and adds it to its 
// parent command.
func New{{ .Resource.Kind }}SubCommand(parentCommand *cobra.Command) {
	initCmd := &cmdinit.InitSubCommand{
		Name:         "{{ .InitCommandName }}",
		Description:  "{{ .InitCommandDescr }}",
		InitFunc:     Init{{ .Resource.Kind }},
		SubCommandOf: parentCommand,
	}

	initCmd.Setup()
}

func Init{{ .Resource.Kind }}(i *cmdinit.InitSubCommand) error {
	manifest, err := get{{ .Resource.Kind }}Manifest(i)
	if err != nil {
		return fmt.Errorf("unable to get manifest for {{ .Resource.Kind }}; %%w", err)
	}

	outputStream := os.Stdout

	if _, err := outputStream.WriteString(manifest); err != nil {
		return fmt.Errorf("failed to write to stdout, %%w", err)
	}

	return nil
}
`
	// cmdInitSubLatest scaffolds the CLI subcommand logic for an individual component's
	// latest version information for use by the api-version flag.
	cmdInitSubLatest = `{{ .Boilerplate }}

	// Code generated by operator-builder. DO NOT EDIT.

	package {{ .Resource.Group }}

	import {{ .Resource.Version }}{{ lower .Resource.Kind }} "{{ .Resource.Path }}/{{ .PackageName }}"

	var latest{{ .Resource.Kind }} = {{ .Resource.Version }}{{ lower .Resource.Kind }}.Sample()
`
)
