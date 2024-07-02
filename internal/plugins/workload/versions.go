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
)

func FromEnv() PluginVersion {
	fromEnv := os.Getenv("OPERATOR_BUILDER_PLUGIN_VERSION")
	if fromEnv == "" {
		return DefaultPluginVersion
	}

	return map[string]PluginVersion{
		"v1": PluginVersionV1,
		"v2": PluginVersionV2,
	}[fromEnv]
}
