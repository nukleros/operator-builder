// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"

	kbcliv3 "sigs.k8s.io/kubebuilder/v3/pkg/cli"
	cfgv3old "sigs.k8s.io/kubebuilder/v3/pkg/config/v3"
	pluginv3 "sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	kustomizecommonv1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/kustomize/v1"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang"
	declarativev1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/declarative/v1"
	golangv3 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3"
	kbcliv4 "sigs.k8s.io/kubebuilder/v4/pkg/cli"
	cfgv3 "sigs.k8s.io/kubebuilder/v4/pkg/config/v3"
	kustomizecommonv2 "sigs.k8s.io/kubebuilder/v4/pkg/plugins/common/kustomize/v2"

	"github.com/nukleros/operator-builder/internal/plugins"
	licensev2 "github.com/nukleros/operator-builder/internal/plugins/license/v2"
	workloadv2 "github.com/nukleros/operator-builder/internal/plugins/workload/v2"
)

var version = "unstable"

func NewKubebuilderCLI(version workload.PluginVersion) (command, error) {
	switch version {
	case workload.PluginVersionV1:
		return NewWithV1()
	case workload.PluginVersionV2:
		return NewWithV2()
	default:
		return NewKubebuilderCLI(workload.DefaultPluginVersion)
	}
}

func NewWithV1() (*kbcliv3.CLI, error) {
	// we cannot upgrade to the latest v4 bundle in kubebuilder as it breaks several dependencies and
	// disallows flexibilities that we currently use such as the apis/ directory versus the api/ directory.
	base, _ := pluginv3.NewBundle(golang.DefaultNameQualifier, pluginv3.Version{Number: 3},
		licensev1.Plugin{},
		kustomizecommonv1.Plugin{},
		configv1.Plugin{},
		golangv3.Plugin{},
		workloadv1.Plugin{},
	)

	c, err := kbcliv3.New(
		kbcliv3.WithCommandName("operator-builder"),
		kbcliv3.WithVersion(version),
		kbcliv3.WithPlugins(
			base,
			&licensev1.Plugin{},
			&kustomizecommonv1.Plugin{},
			&declarativev1.Plugin{},
			&workloadv1.Plugin{},
		),
		kbcliv3.WithDefaultPlugins(cfgv3old.Version, base),
		kbcliv3.WithDefaultProjectVersion(cfgv3old.Version),
		kbcliv3.WithExtraCommands(NewUpdateCmd()),
		kbcliv3.WithExtraCommands(NewInitConfigCmd()),
		kbcliv3.WithCompletion(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create kcli command, %w", err)
	}

	return c, nil
}

func NewWithV2() (*kbcliv4.CLI, error) {
	base, err := pluginv4.NewBundleWithOptions(
		pluginv4.WithName(plugins.DefaultNameQualifier),
		pluginv4.WithVersion(pluginv4.Version{Number: 2}),
		pluginv4.WithPlugins(
			licensev2.Plugin{},
			kustomizecommonv2.Plugin{},
			workloadv2.Plugin{},
		),
	)

	if err != nil {
		return nil, fmt.Errorf("unable to initialize kubebuilder plugin bundle - %w", err)
	}

	c, err := kbcli.New(
		kbcli.WithCommandName("operator-builder"),
		kbcli.WithVersion(version),
		kbcli.WithPlugins(
			base,
			&licensev2.Plugin{},
			kustomizecommonv2.Plugin{},
			workloadv2.Plugin{},
		),
		kbcli.WithDefaultPlugins(cfgv3.Version, base),
		kbcli.WithDefaultProjectVersion(cfgv3.Version),
		kbcli.WithExtraCommands(NewUpdateCmd()),
		kbcli.WithExtraCommands(NewInitConfigCmd()),
		kbcli.WithCompletion(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create kcli command, %w", err)
	}

	return c, nil
}
