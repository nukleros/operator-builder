// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"fmt"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	"github.com/nukleros/operator-builder/internal/plugins/workload"
	workloadconfig "github.com/nukleros/operator-builder/internal/workload/v1/config"
	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

type createAPISubcommand struct {
	workloadConfigPath string
	workload           kinds.WorkloadBuilder
	enableOlm          bool
}

var _ plugin.CreateAPISubcommand = &createAPISubcommand{}

func (p *createAPISubcommand) BindFlags(fs *pflag.FlagSet) {
	workload.AddFlags(fs, &p.workloadConfigPath, &p.enableOlm)
}

func (p *createAPISubcommand) InjectConfig(c config.Config) error {
	processor, err := workloadconfig.Parse(p.workloadConfigPath)
	if err != nil {
		return fmt.Errorf("unable to inject config into %s, %w", p.workloadConfigPath, err)
	}

	p.workload = processor.Workload

	pluginConfig := workloadconfig.Plugin{
		WorkloadConfigPath: p.workloadConfigPath,
		CliRootCommandName: processor.Workload.GetRootCommand().Name,
	}

	if err := c.EncodePluginConfig(workloadconfig.PluginKey, pluginConfig); err != nil {
		return fmt.Errorf("unable to encode plugin config at key %s, %w", workloadconfig.PluginKey, err)
	}

	return nil
}

func (p *createAPISubcommand) InjectResource(res *resource.Resource) error {
	// set from config file if not provided with command line flag
	if res.Group == "" {
		res.Group = p.workload.GetAPIGroup()
	}

	if res.Version == "" {
		res.Version = p.workload.GetAPIVersion()
	}

	if res.Kind == "" {
		res.Kind = p.workload.GetAPIKind()
		res.Plural = resource.RegularPlural(p.workload.GetAPIKind())
	}

	return nil
}

func (p *createAPISubcommand) Scaffold(fs machinery.Filesystem) error {
	return nil
}
