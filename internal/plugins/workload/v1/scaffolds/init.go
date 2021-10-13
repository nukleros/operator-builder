// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package scaffolds

import (
	"fmt"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3/scaffolds"

	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/cli"
	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

const CobraVersion = "v1.1.3"

var _ plugins.Scaffolder = &initScaffolder{}

type initScaffolder struct {
	config             config.Config
	boilerplatePath    string
	workload           workloadv1.WorkloadInitializer
	cliRootCommandName string

	fs machinery.Filesystem
}

// NewInitScaffolder returns a new Scaffolder for project initialization operations.
func NewInitScaffolder(
	cfg config.Config,
	workload workloadv1.WorkloadInitializer,
	cliRootCommandName string,
) plugins.Scaffolder {
	return &initScaffolder{
		config:             cfg,
		boilerplatePath:    "hack/boilerplate.go.txt",
		workload:           workload,
		cliRootCommandName: cliRootCommandName,
	}
}

// InjectFS implements cmdutil.Scaffolder.
func (s *initScaffolder) InjectFS(fs machinery.Filesystem) {
	s.fs = fs
}

// scaffold implements cmdutil.Scaffolder.
func (s *initScaffolder) Scaffold() error {
	fmt.Println("Adding workload scaffolding...")

	boilerplate, err := afero.ReadFile(s.fs.FS, s.boilerplatePath)
	if err != nil {
		return err
	}

	// Initialize the machinery.Scaffold that will write the files to disk
	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
		machinery.WithBoilerplate(string(boilerplate)),
	)

	if s.workload.HasRootCmdName() {
		err = scaffold.Execute(
			&cli.Main{
				RootCmd:        s.workload.GetRootCmdName(),
				RootCmdVarName: utils.ToPascalCase(s.workload.GetRootCmdName()),
			},
			&cli.CmdRoot{
				RootCmd:            s.workload.GetRootCmdName(),
				RootCmdVarName:     utils.ToPascalCase(s.workload.GetRootCmdName()),
				RootCmdDescription: s.workload.GetRootCmdDescr(),
			},
		)
		if err != nil {
			return err
		}
	}

	err = scaffold.Execute(
		&templates.Main{},
		&templates.GoMod{
			ControllerRuntimeVersion: scaffolds.ControllerRuntimeVersion,
			CobraVersion:             CobraVersion,
		},
		&templates.Dockerfile{},
		&templates.Makefile{
			RootCmd: s.workload.GetRootCmdName(),
		},
		&templates.Readme{
			RootCmd: s.workload.GetRootCmdName(),
		},
	)
	if err != nil {
		return err
	}

	return nil
}
