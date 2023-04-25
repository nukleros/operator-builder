// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"errors"
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	"github.com/nukleros/operator-builder/internal/plugins/workload/v1/scaffolds"
	"github.com/nukleros/operator-builder/internal/workload/v1/commands/subcommand"
	workloadconfig "github.com/nukleros/operator-builder/internal/workload/v1/config"
	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

type initSubcommand struct {
	config config.Config

	workloadConfigPath string
	cliRootCommandName string
	controllerImg      string
	enableOlm          bool

	workload kinds.WorkloadBuilder
}

var ErrScaffoldInit = errors.New("unable to scaffold initial config")

var _ plugin.InitSubcommand = &initSubcommand{}

func (p *initSubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {
	subcmdMeta.Description = `Add workload management scaffolding to a new project
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Add project scaffolding defined by a workload config file
  %[1]s init --workload-config .source-manifests/workload.yaml
`, cliMeta.CommandName)
}

func (p *initSubcommand) InjectConfig(c config.Config) error {
	p.config = c

	// operator builder always uses multi-group APIs
	if err := c.SetMultiGroup(); err != nil {
		return fmt.Errorf("unable to enable multigroup apis, %w", err)
	}

	var pluginConfig workloadconfig.Plugin
	if err := c.DecodePluginConfig(workloadconfig.PluginKey, &pluginConfig); err != nil {
		return fmt.Errorf("unable to decode operatorbuilder config key for %s, %w", p.workloadConfigPath, err)
	}

	p.workloadConfigPath = pluginConfig.WorkloadConfigPath
	p.cliRootCommandName = pluginConfig.CliRootCommandName
	p.controllerImg = pluginConfig.ControllerImg
	p.enableOlm = pluginConfig.EnableOLM

	return nil
}

func (p *initSubcommand) PreScaffold(machinery.Filesystem) error {
	processor, err := workloadconfig.Parse(p.workloadConfigPath)
	if err != nil {
		return fmt.Errorf("%s for %s, %w", ErrScaffoldInit.Error(), p.workloadConfigPath, err)
	}

	if err := subcommand.Init(processor); err != nil {
		return fmt.Errorf("%s for %s, %w", ErrScaffoldInit.Error(), p.workloadConfigPath, err)
	}

	p.workload = processor.Workload

	return nil
}

func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {
	scaffolder := scaffolds.NewInitScaffolder(
		p.config,
		p.workload,
		p.cliRootCommandName,
		p.controllerImg,
		p.enableOlm,
	)
	scaffolder.InjectFS(fs)

	if err := scaffolder.Scaffold(); err != nil {
		return fmt.Errorf("%s for %s, %w", ErrScaffoldInit.Error(), p.workloadConfigPath, err)
	}

	return nil
}
