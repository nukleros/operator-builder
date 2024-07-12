// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bufio"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	kbcliv3 "sigs.k8s.io/kubebuilder/v3/pkg/cli"
	cfgv3old "sigs.k8s.io/kubebuilder/v3/pkg/config/v3"
	pluginv3 "sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin/util"
	kustomizecommonv1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/kustomize/v1"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang"
	declarativev1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/declarative/v1"
	golangv3 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3"
	kbcliv4 "sigs.k8s.io/kubebuilder/v4/pkg/cli"
	cfgv3 "sigs.k8s.io/kubebuilder/v4/pkg/config/v3"
	pluginv4 "sigs.k8s.io/kubebuilder/v4/pkg/plugin"
	kustomizecommonv2 "sigs.k8s.io/kubebuilder/v4/pkg/plugins/common/kustomize/v2"

	"github.com/nukleros/operator-builder/internal/plugins"
	configv1 "github.com/nukleros/operator-builder/internal/plugins/config/v1"
	licensev1 "github.com/nukleros/operator-builder/internal/plugins/license/v1"
	licensev2 "github.com/nukleros/operator-builder/internal/plugins/license/v2"
	"github.com/nukleros/operator-builder/internal/plugins/workload"
	workloadv1 "github.com/nukleros/operator-builder/internal/plugins/workload/v1"
	workloadv2 "github.com/nukleros/operator-builder/internal/plugins/workload/v2"
)

var version = "unstable"

const (
	commandName = "operator-builder"
)

type command interface {
	Run() error
}

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
	// the v1 version of the plugin will be deprecated in a future release.  notify the user and ask if they want
	// to proceed.
	log.Println("workload v1 plugin selected, but will be deprecated in a future release.  Proceed with v1 [y/n]?")

	reader := bufio.NewReader(os.Stdin)
	proceed := util.YesNo(reader)
	if !proceed {
		return nil, nil
	}

	// we cannot upgrade to the latest v4 bundle in kubebuilder as it breaks several dependencies and
	// disallows flexibilities that we currently use such as the apis/ directory versus the api/ directory.
	base, err := pluginv3.NewBundle(golang.DefaultNameQualifier, pluginv3.Version{Number: 3},
		licensev1.Plugin{},
		kustomizecommonv1.Plugin{},
		configv1.Plugin{},
		golangv3.Plugin{},
		workloadv1.Plugin{},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize kubebuilder plugin bundle with version 1 plugin, %w", err)
	}

	c, err := kbcliv3.New(
		kbcliv3.WithCommandName(commandName),
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
		return nil, fmt.Errorf("unable to create command with version 1 plugin, %w", err)
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
		return nil, fmt.Errorf("unable to initialize kubebuilder plugin bundle with version 2 plugin, %w", err)
	}

	c, err := kbcliv4.New(
		kbcliv4.WithCommandName(commandName),
		kbcliv4.WithVersion(version),
		kbcliv4.WithPlugins(
			base,
			&licensev2.Plugin{},
			kustomizecommonv2.Plugin{},
			workloadv2.Plugin{},
		),
		kbcliv4.WithDefaultPlugins(cfgv3.Version, base),
		kbcliv4.WithDefaultProjectVersion(cfgv3.Version),
		kbcliv4.WithExtraCommands(NewInitConfigCmd()),
		kbcliv4.WithCompletion(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create command with version 2 plugin, %w", err)
	}

	return c, nil
}
