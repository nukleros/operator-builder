// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"

	kbcli "sigs.k8s.io/kubebuilder/v3/pkg/cli"
	cfgv3 "sigs.k8s.io/kubebuilder/v3/pkg/config/v3"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	kustomizecommonv1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/kustomize/v1"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang"
	declarativev1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/declarative/v1"
	golangv3 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3"

	configv1 "github.com/nukleros/operator-builder/internal/plugins/config/v1"
	licensev1 "github.com/nukleros/operator-builder/internal/plugins/license/v1"
	workloadv1 "github.com/nukleros/operator-builder/internal/plugins/workload/v1"
)

var version = "unstable"

func NewKubebuilderCLI() (*kbcli.CLI, error) {
	// we cannot upgrade to the latest v4 bundle in kubebuilder as it breaks several dependencies and
	// disallows flexibilities that we currently use such as the apis/ directory versus the api/ directory.
	base, _ := plugin.NewBundle(golang.DefaultNameQualifier, plugin.Version{Number: 3},
		licensev1.Plugin{},
		kustomizecommonv1.Plugin{},
		configv1.Plugin{},
		golangv3.Plugin{},
		workloadv1.Plugin{},
	)

	c, err := kbcli.New(
		kbcli.WithCommandName("operator-builder"),
		kbcli.WithVersion(version),
		kbcli.WithPlugins(
			base,
			&licensev1.Plugin{},
			&kustomizecommonv1.Plugin{},
			&declarativev1.Plugin{},
			&workloadv1.Plugin{},
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
