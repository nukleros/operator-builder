// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package utils

const (
	// specifies the minimum version required to build the generated project.
	GeneratedGoVersionMinimum = "1.21"

	// specifies the preferred version.
	GeneratedGoVersionPreferred = "1.22"

	// makefile and go.mod versions.
	// NOTE: please ensure the ControllerToolsVersion matches the go.mod file as we
	// use this both in code generation as well as the generated project code.
	ControllerToolsVersion = "v0.15.0"

	// NOTE: ControllerRuntimeVersion will need to match operator-builder-tools version
	// otherwise their could be inconsistencies in method calls which cause
	// ambiguous errors.
	ControllerRuntimeVersion = "v0.17.3"
	KustomizeVersion         = "v5.4.1"
	GolangCILintVersion      = "v1.57.2"
	EnvtestVersion           = "release-0.17"
	EnvtestK8SVersion        = "1.30.0"
	OperatorSDKVersion       = "v1.28.0"
)
