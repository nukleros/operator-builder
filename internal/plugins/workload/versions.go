// Copyright 2024 Nukleros
// SPDX-License-Identifier: MIT

package workload

import "os"

type PluginVersion int

const (
	PluginVersionUnknown PluginVersion = iota
	PluginVersionV1
	PluginVersionV2
)

const (
	DefaultPluginVersion = PluginVersionV2

	EnvPluginVersionVariable = "OPERATOR_BUILDER_PLUGIN_VERSION"
	EnvPluginVersionV1       = "v1"
	EnvPluginVersionV2       = "v2"
)

func FromEnv() PluginVersion {
	return map[string]PluginVersion{
		"":                 DefaultPluginVersion,
		EnvPluginVersionV1: PluginVersionV1,
		EnvPluginVersionV2: PluginVersionV2,
	}[os.Getenv(EnvPluginVersionVariable)]
}
