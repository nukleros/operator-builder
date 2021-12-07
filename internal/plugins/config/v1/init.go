// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

type initSubcommand struct {
	workloadConfigPath string
}

var _ plugin.InitSubcommand = &initSubcommand{}

func (p *initSubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&p.workloadConfigPath, "workload-config", "", "path to workload config file")
}

func (p *initSubcommand) InjectConfig(c config.Config) error {
	workload, err := workloadv1.ProcessInitConfig(
		p.workloadConfigPath,
	)
	if err != nil {
		return fmt.Errorf("unable to inject config into %s, %w", p.workloadConfigPath, err)
	}

	pluginConfig := workloadv1.PluginConfig{
		WorkloadConfigPath: p.workloadConfigPath,
		CliRootCommandName: workload.GetRootCommand().Name,
	}

	if err := c.EncodePluginConfig(workloadv1.PluginConfigKey, pluginConfig); err != nil {
		return fmt.Errorf("unable to encode operatorbuilder config key at %s, %w", p.workloadConfigPath, err)
	}

	if err := c.SetDomain(workload.GetDomain()); err != nil {
		return fmt.Errorf("unable to set project domain, %w", err)
	}

	return nil
}

func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {
	return nil
}
