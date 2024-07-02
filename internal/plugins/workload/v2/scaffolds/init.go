// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package scaffolds

import (
	"fmt"
	"log"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v4/pkg/config"
	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v4/pkg/plugins"

	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates/cli"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates/config/manifests"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates/config/scorecard"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates/hack"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates/test/e2e"
	"github.com/nukleros/operator-builder/internal/utils"
	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

var _ plugins.Scaffolder = &initScaffolder{}

type initScaffolder struct {
	config             config.Config
	boilerplatePath    string
	workload           kinds.WorkloadBuilder
	cliRootCommandName string
	controllerImg      string
	enableOlm          bool
	license            string
	owner              string

	fs machinery.Filesystem
}

// NewInitScaffolder returns a new Scaffolder for project initialization operations.
func NewInitScaffolder(
	cfg config.Config,
	workload kinds.WorkloadBuilder,
	cliRootCommandName string,
	controllerImg string,
	enableOlm bool,
	license string,
	owner string,
) plugins.Scaffolder {
	return &initScaffolder{
		config:             cfg,
		boilerplatePath:    hack.DefaultBoilerplatePath,
		workload:           workload,
		cliRootCommandName: cliRootCommandName,
		controllerImg:      controllerImg,
		enableOlm:          enableOlm,
		license:            license,
		owner:              owner,
	}
}

// InjectFS implements cmdutil.Scaffolder.
func (s *initScaffolder) InjectFS(fs machinery.Filesystem) {
	s.fs = fs
}

// scaffold implements cmdutil.Scaffolder.
func (s *initScaffolder) Scaffold() error {
	log.Println("Adding workload scaffolding...")

	// Initialize the machinery.Scaffold that will write the boilerplate file to disk
	// The boilerplate file needs to be scaffolded as a separate step as it is going to
	// be used by the rest of the files, even those scaffolded in this command call.
	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
	)

	if s.license != "none" {
		bpFile := &hack.Boilerplate{
			License: s.license,
			Owner:   s.owner,
		}

		bpFile.Path = s.boilerplatePath
		if err := scaffold.Execute(bpFile); err != nil {
			return err
		}

		boilerplate, err := afero.ReadFile(s.fs.FS, s.boilerplatePath)
		if err != nil {
			return err
		}
		// Initialize the machinery.Scaffold that will write the files to disk
		scaffold = machinery.NewScaffold(s.fs,
			machinery.WithConfig(s.config),
			machinery.WithBoilerplate(string(boilerplate)),
		)
	} else {
		s.boilerplatePath = ""
		// Initialize the machinery.Scaffold without boilerplate
		scaffold = machinery.NewScaffold(s.fs,
			machinery.WithConfig(s.config),
		)
	}

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
		&templates.GoMod{},
		&templates.Dockerfile{},
		&templates.Readme{
			RootCmdName:   s.cliRootCommandName,
			EnableOLM:     s.enableOlm,
			ControllerImg: s.controllerImg,
		},
		&templates.Makefile{
			RootCmdName:              s.cliRootCommandName,
			ControllerImg:            s.controllerImg,
			EnableOLM:                s.enableOlm,
			KustomizeVersion:         utils.KustomizeVersion,
			ControllerToolsVersion:   utils.ControllerToolsVersion,
			OperatorSDKVersion:       utils.OperatorSDKVersion,
			ControllerRuntimeVersion: utils.ControllerRuntimeVersion,
			EnvtestVersion:           utils.EnvtestVersion,
			EnvtestK8SVersion:        utils.EnvtestK8SVersion,
			GolangCILintVersion:      utils.GolangCILintVersion,
		},
		&e2e.Test{},
	); err != nil {
		return fmt.Errorf("unable to scaffold initial configuration, %w", err)
	}

	if s.enableOlm {
		if err := scaffold.Execute(
			&scorecard.Scorecard{ScorecardType: scorecard.ScorecardTypeBase},
			&scorecard.Scorecard{ScorecardType: scorecard.ScorecardTypeKustomize},
			&scorecard.Scorecard{ScorecardType: scorecard.ScorecardTypePatchesBasic},
			&scorecard.Scorecard{ScorecardType: scorecard.ScorecardTypePatchesOLM},
		); err != nil {
			return fmt.Errorf("unable to scaffold OLM scorecard configuration, %w", err)
		}

		if err := scaffold.Execute(
			&manifests.Kustomization{
				SupportsKustomizeV4: false,
				SupportsWebhooks:    false,
			},
		); err != nil {
			return fmt.Errorf("unable to scaffold manifests, %w", err)
		}
	}

	return nil
}
