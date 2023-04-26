// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	workloadconfig "github.com/nukleros/operator-builder/internal/workload/v1/config"
)

type initSubcommand struct {
	workloadConfigPath string
	controllerImage    string
	enableOlm          bool
}

var _ plugin.InitSubcommand = &initSubcommand{}

func (p *initSubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&p.workloadConfigPath, "workload-config", "", "path to workload config file")
	fs.StringVar(&p.controllerImage, "controller-image", "controller:latest", "controller image")
	fs.BoolVar(&p.enableOlm, "enable-olm", false, "enable support for Operator Lifecycle Manager (OLM)")
}

func (p *initSubcommand) InjectConfig(c config.Config) error {
	processor, err := workloadconfig.Parse(p.workloadConfigPath)
	if err != nil {
		return fmt.Errorf("unable to inject config into %s, %w", p.workloadConfigPath, err)
	}

	pluginConfig := workloadconfig.Plugin{
		WorkloadConfigPath: p.workloadConfigPath,
		CliRootCommandName: processor.Workload.GetRootCommand().Name,
		ControllerImg:      p.controllerImage,
		EnableOLM:          p.enableOlm,
	}

	if err := c.EncodePluginConfig(workloadconfig.PluginKey, pluginConfig); err != nil {
		return fmt.Errorf("unable to encode operatorbuilder config key at %s, %w", p.workloadConfigPath, err)
	}

	if err := c.SetDomain(processor.Workload.GetDomain()); err != nil {
		return fmt.Errorf("unable to set project domain, %w", err)
	}

	return nil
}

func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {
	return nil
}
