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
		machinery.NewMarkerFor(f.Path, initSamplesMarker),
		machinery.NewMarkerFor(f.Path, initSwitchesMarker),
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
}

func (f *CmdInitSubLatest) SetTemplateDefaults() error {
	if f.Builder.IsComponent() {
		f.Resource = f.ComponentResource
	}

	// set template fields
	f.RootCmd = *f.Builder.GetRootCommand()
	f.SubCmd = *f.Builder.GetSubCommand()

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

const initSamplesMarker = "operator-builder:samples"
const initSwitchesMarker = "operator-builder:switches"

// GetMarkers implements file.Inserter interface.
func (f *CmdInitSubUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.GetPath(), initSamplesMarker),
		machinery.NewMarkerFor(f.GetPath(), initSwitchesMarker),
	}
}

// Code Fragments.
const (
	// initSamplesFragment is a fragment which provides the code fragment to display
	// the sample custom resource manifest for an individual component.
	initSamplesFragment = `const %s = ` + "`" + `apiVersion: %s/%s
kind: %s
metadata:
  name: %s-sample
%s` + "`" + `
`

	// initSwitchFragment is a fragment which provides a new switch for each api version
	// that is created for use by the api-version flag.
	initSwitchesFragment = `case "%s":
	return %s, nil
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
	samples := make([]string, 0)
	switches := make([]string, 0)

	// add the samples
	manifestVarName := fmt.Sprintf("%s%s", f.Resource.Version, f.Resource.Kind)
	samples = append(samples, fmt.Sprintf(initSamplesFragment,
		manifestVarName,
		f.Resource.QualifiedGroup(),
		f.Resource.Version,
		f.Resource.Kind,
		strings.ToLower(f.Resource.Kind),
		f.Builder.GetAPISpecFields().GenerateSampleSpec()),
	)

	// add the switches
	switches = append(switches, fmt.Sprintf(initSwitchesFragment,
		f.Resource.Version,
		manifestVarName),
	)

	// Only store code fragments in the map if the slices are non-empty
	if len(samples) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), initSamplesMarker)] = samples
	}

	if len(switches) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), initSwitchesMarker)] = switches
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
)

%s

// get{{ .Resource.Kind }}Manifest returns the sample {{ .Resource.Kind }} manifest
// based upon API Version input.
func get{{ .Resource.Kind }}Manifest(i *cmdinit.InitSubCommand) (string, error) {
	switch i.APIVersion {
	// return the latest version if unspecified or latest requested
	case "", "latest":
		return latest{{ .Resource.Kind }}, nil
	%s
	default:
		return "", fmt.Errorf("unsupported API Version: "+i.APIVersion)
	}
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

	const latest{{ .Resource.Kind }} = {{ .Resource.Version }}{{ .Resource.Kind }}
`
)
