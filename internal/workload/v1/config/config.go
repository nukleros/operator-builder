// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package config

const PluginKey = "operatorBuilder"

// Plugin contains the project config values which are stored in the
// PROJECT file under plugins.operatorBuilder.
type Plugin struct {
	WorkloadConfigPath string `json:"workloadConfigPath" yaml:"workloadConfigPath"`
	CliRootCommandName string `json:"cliRootCommandName" yaml:"cliRootCommandName"`
	ControllerImg      string `json:"controllerImg" yaml:"controllerImg"`
}

