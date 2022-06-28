// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/commands/subcommand"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/kinds"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/manifests"
)

var (
	ErrInvalidSubCommandName          = errors.New("invalid subcommand name")
	ErrInitConfigCollectionSubcommand = errors.New("error executing `init-config collection` subcommand")
	ErrInitConfigComponentSubcommand  = errors.New("error executing `init-config component` subcommand")
	ErrInitConfigStandaloneSubcommand = errors.New("error executing `init-config standalone` subcommand")
)

const (
	initConfigName        = "init-config"
	initConfigDescription = "Initialize a workload configuration"

	collectionSubCommandName = "collection"
	componentSubCommandName  = "component"
	standaloneSubCommandName = "standalone"

	collectionSubCommandDescription = "initialize a collection workload configuration"
	componentSubCommandDescription  = "initialize a component workload configuration"
	standaloneSubCommandDescription = "initialize a standalone workload configuration"

	collectionSampleName = "workload-collection-config"
	componentSampleName  = "component-workload-config"
	standaloneSampleName = "standalone-workload-config"

	sampleComponentFile = "/path/to/my/component-workload-config.yaml"
	sampleResourceFile  = "/path/to/my/child-resources.yaml"
)

type initConfigSubCommand struct {
	subCommandName        string
	subCommandDescription string
	options               *subcommand.InitConfigOptions
}

func NewInitConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   initConfigName,
		Short: initConfigDescription,
		Long:  initConfigDescription,
	}

	// loop through the workload types and create a subcommand for each
	for subCommandName, subCommandDescription := range map[string]string{
		collectionSubCommandName: collectionSubCommandDescription,
		componentSubCommandName:  componentSubCommandDescription,
		standaloneSubCommandName: standaloneSubCommandDescription,
	} {
		initConfigSubCommand := newInitConfigSubCommand(subCommandName, subCommandDescription)
		if err := initConfigSubCommand.addInitSubCommand(cmd); err != nil {
			panic(err)
		}
	}

	return cmd
}

func newInitConfigSubCommand(subCmdName, subCmdDescription string) *initConfigSubCommand {
	return &initConfigSubCommand{
		subCommandName:        subCmdName,
		subCommandDescription: subCmdDescription,
		options:               &subcommand.InitConfigOptions{},
	}
}

func (i *initConfigSubCommand) addInitSubCommand(parentCommand *cobra.Command) error {
	var returnErr, err error

	subCommand := &cobra.Command{
		Use:   i.subCommandName,
		Short: i.subCommandDescription,
		Long:  i.subCommandDescription,
	}

	switch subCommand.Use {
	case collectionSubCommandName:
		returnErr = ErrInitConfigCollectionSubcommand

		i.options.WorkloadConfig = newCollectionConfigSample()
	case componentSubCommandName:
		returnErr = ErrInitConfigComponentSubcommand

		i.options.WorkloadConfig = newComponentConfigSample()
	case standaloneSubCommandName:
		returnErr = ErrInitConfigStandaloneSubcommand

		i.options.WorkloadConfig = newStandaloneConfigSample()
	default:
		err = fmt.Errorf("%w - %s", ErrInvalidSubCommandName, subCommand.Use)
	}

	if err != nil {
		return err
	}

	subCommand.RunE = func(cmd *cobra.Command, args []string) error {
		if err := subcommand.InitConfig(i.options); err != nil {
			return fmt.Errorf("%w; %s", err, returnErr)
		}

		return nil
	}

	parentCommand.AddCommand(subCommand)

	return i.addCommonFlags(subCommand)
}

func (i *initConfigSubCommand) addCommonFlags(cmd *cobra.Command) error {
	// add the path flag
	cmd.Flags().StringVarP(&i.options.Path, "path", "p", "-", "file path to initialize workload at (default: stdout)")

	// add the force flag
	cmd.Flags().BoolVarP(&i.options.Force, "force", "f", false, "override the config if it already exists")

	return nil
}

func newCollectionConfigSample() *kinds.WorkloadCollection {
	sample := kinds.NewWorkloadCollection(
		collectionSampleName,
		*kinds.NewSampleAPISpec(),
		[]string{sampleComponentFile},
	)

	sample.Spec.CompanionCliRootcmd.SetDefaults(sample, false)
	sample.Spec.CompanionCliSubcmd.SetDefaults(sample, true)
	sample.Spec.Manifests = manifests.FromFiles([]string{sampleResourceFile})

	return sample
}

func newComponentConfigSample() *kinds.ComponentWorkload {
	sample := kinds.NewComponentWorkload(
		componentSampleName,
		*kinds.NewSampleAPISpec(),
		[]string{sampleComponentFile},
		[]string{componentSampleName + "-2"},
	)

	sample.Spec.CompanionCliSubcmd.SetDefaults(sample, true)

	return sample
}

func newStandaloneConfigSample() *kinds.StandaloneWorkload {
	sample := kinds.NewStandaloneWorkload(
		standaloneSampleName,
		*kinds.NewSampleAPISpec(),
		[]string{sampleResourceFile},
	)

	sample.Spec.CompanionCliRootcmd.SetDefaults(sample, false)

	return sample
}

