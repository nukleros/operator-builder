// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/commands/companion"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/kinds"
)

var (
	_ machinery.Template = &CmdVersionSub{}
	_ machinery.Inserter = &CmdVersionSubUpdater{}
)

// cmdVersionSubCommon include the common fields that are shared by all version
// subcommand structs for templating purposes.
type cmdVersionSubCommon struct {
	RootCmd companion.CLI
	SubCmd  companion.CLI
}

// CmdVersionSub scaffolds the root command file for the companion CLI.
type CmdVersionSub struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin
	machinery.RepositoryMixin

	// input fields
	Builder kinds.WorkloadBuilder

	// template fields
	cmdVersionSubCommon
	VersionCommandName  string
	VersionCommandDescr string
}

func (f *CmdVersionSub) SetTemplateDefaults() error {
	// set template fields
	f.RootCmd = *f.Builder.GetRootCommand()
	f.SubCmd = *f.Builder.GetSubCommand()

	if f.Builder.IsStandalone() {
		f.VersionCommandName = versionCommandName
		f.VersionCommandDescr = versionCommandDescr
	} else {
		f.VersionCommandName = f.SubCmd.Name
		f.VersionCommandDescr = f.SubCmd.Description
	}

	// set interface fields
	f.Path = f.SubCmd.GetSubCmdRelativeFileName(
		f.RootCmd.Name,
		"version",
		f.Resource.Group,
		utils.ToFileName(f.Resource.Kind),
	)

	f.TemplateBody = cmdVersionSub

	return nil
}

// CmdVersionSubUpdater updates a specific components version subcommand with
// appropriate version information.
type CmdVersionSubUpdater struct { //nolint:maligned
	machinery.RepositoryMixin
	machinery.MultiGroupMixin
	machinery.ResourceMixin

	// input fields
	Builder kinds.WorkloadBuilder

	// template fields
	cmdVersionSubCommon
}

// GetPath implements file.Builder interface.
func (f *CmdVersionSubUpdater) GetPath() string {
	return f.SubCmd.GetSubCmdRelativeFileName(
		f.Builder.GetRootCommand().Name,
		"version",
		f.Resource.Group,
		utils.ToFileName(f.Resource.Kind),
	)
}

// GetIfExistsAction implements file.Builder interface.
func (*CmdVersionSubUpdater) GetIfExistsAction() machinery.IfExistsAction {
	return machinery.OverwriteFile
}

const apiVersionsMarker = "operator-builder:apiversions"

// GetMarkers implements file.Inserter interface.
func (f *CmdVersionSubUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.GetPath(), apiVersionsMarker),
	}
}

// Code Fragments.
const (
	versionCodeFragment = `"%s",
`
)

// GetCodeFragments implements file.Inserter interface.
func (f *CmdVersionSubUpdater) GetCodeFragments() machinery.CodeFragmentsMap {
	fragments := make(machinery.CodeFragmentsMap, 1)

	// If resource is not being provided we are creating the file, not updating it
	if f.Resource == nil {
		return fragments
	}

	// set template fields
	f.RootCmd = *f.Builder.GetRootCommand()
	f.SubCmd = *f.Builder.GetSubCommand()

	// Generate subCommands code fragments
	apiVersions := make([]string, 0)
	apiVersions = append(apiVersions, fmt.Sprintf(versionCodeFragment, f.Resource.Version))

	// Only store code fragments in the map if the slices are non-empty
	if len(apiVersions) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), apiVersionsMarker)] = apiVersions
	}

	return fragments
}

const (
	cmdVersionSub = `{{ .Boilerplate }}

package {{ .Resource.Group }}

import (
	"github.com/spf13/cobra"
	
	cmdversion "{{ .Repo }}/cmd/{{ .RootCmd.Name }}/commands/version"

	"{{ .Repo }}/apis/{{ .Resource.Group }}"
)

// New{{ .Resource.Kind }}SubCommand creates a new command and adds it to its 
// parent command.
func New{{ .Resource.Kind }}SubCommand(parentCommand *cobra.Command) {
	versionCmd := &cmdversion.VersionSubCommand{
		Name:         "{{ .VersionCommandName }}",
		Description:  "{{ .VersionCommandDescr }}",
		VersionFunc:  Version{{ .Resource.Kind }},
		SubCommandOf: parentCommand,
	}

	versionCmd.Setup()
}

func Version{{ .Resource.Kind }}(v *cmdversion.VersionSubCommand) error {
	apiVersions := make([]string, len({{ .Resource.Group }}.{{ .Resource.Kind }}GroupVersions()))

	for i, groupVersion := range {{ .Resource.Group }}.{{ .Resource.Kind }}GroupVersions() {
		apiVersions[i] = groupVersion.Version
	}

	versionInfo := cmdversion.VersionInfo{
		CLIVersion:  cmdversion.CLIVersion,
		APIVersions: apiVersions,
	}

	return versionInfo.Display()
}
`
)
