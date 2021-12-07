// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package scaffolds

import (
	"fmt"
	"log"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3/scaffolds"

	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/cli"
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
	log.Println("Adding workload scaffolding...")

	boilerplate, err := afero.ReadFile(s.fs.FS, s.boilerplatePath)
	if err != nil {
		return fmt.Errorf("unable to read boilerplate file %s, %w", s.boilerplatePath, err)
	}

	// Initialize the machinery.Scaffold that will write the files to disk
	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
		machinery.WithBoilerplate(string(boilerplate)),
	)

	if s.workload.HasRootCmdName() {
		if err := scaffold.Execute(
			&cli.Main{RootCmd: *s.workload.GetRootCommand()},
			&cli.CmdRoot{Initializer: s.workload},
			&cli.CmdInit{Initializer: s.workload},
			&cli.CmdGenerate{Initializer: s.workload},
			&cli.CmdVersion{Initializer: s.workload},
		); err != nil {
			return fmt.Errorf("unable to scaffold initial configuration for companionCli, %w", err)
		}
	}

	if err := scaffold.Execute(
		&templates.Main{},
		&templates.GoMod{
			ControllerRuntimeVersion: scaffolds.ControllerRuntimeVersion,
			CobraVersion:             CobraVersion,
		},
		&templates.Dockerfile{},
		&templates.Makefile{RootCmdName: s.cliRootCommandName},
		&templates.Readme{RootCmdName: s.cliRootCommandName},
	); err != nil {
		return fmt.Errorf("unable to scaffold initial configuration, %w", err)
	}

	return nil
}
