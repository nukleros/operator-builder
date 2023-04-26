// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"

	kbcli "sigs.k8s.io/kubebuilder/v3/pkg/cli"
	cfgv2 "sigs.k8s.io/kubebuilder/v3/pkg/config/v2"
	cfgv3 "sigs.k8s.io/kubebuilder/v3/pkg/config/v3"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/stage"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	kustomizecommonv1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/kustomize/v1"
	kustomizecommonv2alpha "sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/kustomize/v2-alpha"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang"
	declarativev1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/declarative/v1"
	golangv2 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v2"
	golangv3 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3"

	configv1 "github.com/nukleros/operator-builder/internal/plugins/config/v1"
	licensev1 "github.com/nukleros/operator-builder/internal/plugins/license/v1"
	workloadv1 "github.com/nukleros/operator-builder/internal/plugins/workload/v1"
)

var version = "unstable"

func NewKubebuilderCLI() (*kbcli.CLI, error) {
	gov3Bundle, _ := plugin.NewBundle(golang.DefaultNameQualifier, plugin.Version{Number: 3},
		licensev1.Plugin{},
		kustomizecommonv1.Plugin{},
		configv1.Plugin{},
		golangv3.Plugin{},
		workloadv1.Plugin{},
	)

	gov4Bundle, _ := plugin.NewBundle(golang.DefaultNameQualifier, plugin.Version{Number: 4, Stage: stage.Alpha},
		licensev1.Plugin{},
		kustomizecommonv2alpha.Plugin{},
		configv1.Plugin{},
		golangv3.Plugin{},
		workloadv1.Plugin{},
	)

	c, err := kbcli.New(
		kbcli.WithCommandName("operator-builder"),
		kbcli.WithVersion(version),
		kbcli.WithPlugins(
			golangv2.Plugin{},
			gov3Bundle,
			gov4Bundle,
			&licensev1.Plugin{},
			&kustomizecommonv1.Plugin{},
			&declarativev1.Plugin{},
			&workloadv1.Plugin{},
		),
		kbcli.WithDefaultPlugins(cfgv2.Version, golangv2.Plugin{}),
		kbcli.WithDefaultPlugins(cfgv3.Version, gov3Bundle),
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
