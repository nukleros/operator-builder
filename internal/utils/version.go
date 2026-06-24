// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package utils

const (
	// specifies the minimum version required to build the generated project.
	GeneratedGoVersionMinimum = "1.25.11"

	// specifies the preferred version.
	GeneratedGoVersionPreferred = "1.26.4"

	// makefile and go.mod versions.
	// NOTE: please ensure the ControllerToolsVersion matches the go.mod file as we
	// use this both in code generation as well as the generated project code.
	ControllerToolsVersion = "v0.21.0"

	// NOTE: ControllerRuntimeVersion will need to match operator-builder-tools version
	// otherwise their could be inconsistencies in method calls which cause
	// ambiguous errors.
	ControllerRuntimeVersion = "v0.24.1"
	KustomizeVersion         = "v5.4.1"
	GolangCILintVersion      = "v2.12.2"
	EnvtestVersion           = "release-0.17"
	EnvtestK8SVersion        = "1.30.0"
	OperatorSDKVersion       = "v1.28.0"

	// the following are dependency versions that are shared across both v1 and v2 plugins.
	// any updates to these versions will affect both plugins and should be tested accordingly.
	GoLogrVersion               = "v1.4.3"
	OperatorBuilderToolsVersion = "v0.8.0"
	GinkgoVersion               = "v2.32.0"
	GomegaVersion               = "v1.42.0"
	CobraVersion                = "v1.10.2"
	TestifyVersion              = "v1.11.1"
	YAMLVersionV2               = "v2.4.0"
	KubernetesYAMLVersion       = "v1.6.0"

	// the following is the kubernetes library version.  it affects the api/machinery/client-go
	// packages.
	KubernetesLibraryVersion = "v0.36.2"

	// the following are kubebuilder dependency versions.  they will affect on the plugins
	// which are associated with the approriate kubebuilder version.
	KubebuilderVersionV3 = "v3.7.0"
	KubebuilderVersionV4 = "v4.15.0"
)
