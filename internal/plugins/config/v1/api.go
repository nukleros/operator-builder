// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

type createAPISubcommand struct {
	workloadConfigPath string
}

var _ plugin.CreateAPISubcommand = &createAPISubcommand{}

func (p *createAPISubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&p.workloadConfigPath, "workload-config", "", "path to workload config file")
}

func (p *createAPISubcommand) InjectConfig(c config.Config) error {
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
		return fmt.Errorf("unable to encode plugin config at key %s, %w", workloadv1.PluginConfigKey, err)
	}

	return nil
}

func (p *createAPISubcommand) InjectResource(res *resource.Resource) error {
	workload, err := workloadv1.ProcessAPIConfig(
		p.workloadConfigPath,
	)
	if err != nil {
		return fmt.Errorf("unable to inject resource into %s, %w", p.workloadConfigPath, err)
	}

	// set from config file if not provided with command line flag
	if res.Group == "" {
		res.Group = workload.GetAPIGroup()
	}

	if res.Version == "" {
		res.Version = workload.GetAPIVersion()
	}

	if res.Kind == "" {
		res.Kind = workload.GetAPIKind()
		res.Plural = resource.RegularPlural(workload.GetAPIKind())
	}

	return nil
}

func (p *createAPISubcommand) Scaffold(fs machinery.Filesystem) error {
	return nil
}
