// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package scaffolds

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v4/pkg/config"
	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v4/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v4/pkg/plugins"

	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates/api"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates/api/resources"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates/cli"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates/config/crd"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates/config/samples"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates/controller"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates/int/dependencies"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates/int/mutate"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds/templates/test/e2e"
	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

const boilerplatePath = "hack/boilerplate.go.txt"

var _ plugins.Scaffolder = &apiScaffolder{}

var (
	ErrScaffoldWorkload             = errors.New("error scaffolding workload")
	ErrScaffoldMainUpdater          = errors.New("error updating main.go")
	ErrScaffoldCRDSample            = errors.New("error scaffolding CRD sample file")
	ErrScaffoldKustomization        = errors.New("error scaffolding kustomization overlay")
	ErrScaffoldAPITypes             = errors.New("error scaffolding api types")
	ErrScaffoldAPIKindInfo          = errors.New("error scaffolding api kind information")
	ErrScaffoldAPIResources         = errors.New("error scaffolding api resource methods")
	ErrScaffoldAPIChildResources    = errors.New("error scaffolding api child resource definitions")
	ErrScaffoldController           = errors.New("error scaffolding controller logic")
	ErrScaffoldE2ETest              = errors.New("error scaffolding e2e tests")
	ErrScaffoldCompanionCLI         = errors.New("error scaffolding companion CLI")
	ErrScaffoldCompanionCLIInit     = errors.New("error scaffolding companion CLI init sub-command")
	ErrScaffoldCompanionCLIGenerate = errors.New("error scaffolding companion CLI generate sub-command")
	ErrScaffoldCompanionCLIVersion  = errors.New("error scaffolding companion CLI version sub-command")
	ErrScaffoldCompanionCLIRoot     = errors.New("error scaffolding companion CLI root.go entrypoint")
)

type apiScaffolder struct {
	fs machinery.Filesystem

	config             config.Config
	resource           *resource.Resource
	boilerplate        string
	workload           kinds.WorkloadBuilder
	cliRootCommandName string
	enableOlm          bool
}

// NewAPIScaffolder returns a new Scaffolder for project initialization operations.
func NewAPIScaffolder(
	cfg config.Config,
	res *resource.Resource,
	workload kinds.WorkloadBuilder,
	cliRootCommandName string,
	enableOlm bool,
) plugins.Scaffolder {
	return &apiScaffolder{
		config:             cfg,
		resource:           res,
		workload:           workload,
		cliRootCommandName: cliRootCommandName,
		enableOlm:          enableOlm,
	}
}

// InjectFS implements cmdutil.Scaffolder.
func (s *apiScaffolder) InjectFS(fs machinery.Filesystem) {
	s.fs = fs
}

// scaffold implements cmdutil.Scaffolder.
func (s *apiScaffolder) Scaffold() error {
	log.Println("Building API...")

	boilerplate, err := afero.ReadFile(s.fs.FS, boilerplatePath)
	if err != nil {
		return fmt.Errorf("unable to read boilerplate file %s, %w", boilerplatePath, err)
	}

	s.boilerplate = string(boilerplate)

	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
		machinery.WithBoilerplate(s.boilerplate),
		machinery.WithResource(s.resource),
	)

	// scaffold the workload
	if err := s.scaffoldWorkload(scaffold, s.workload); err != nil {
		return fmt.Errorf("%w; %s for workload type %T", err, ErrScaffoldWorkload.Error(), s.workload)
	}

	return nil
}

// scaffoldWorkload performs the execution of the scaffold for an individual workload.
//
//nolint:funlen
func (s *apiScaffolder) scaffoldWorkload(
	scaffold *machinery.Scaffold,
	workload kinds.WorkloadBuilder,
) error {
	componentResource := workload.GetComponentResource(
		s.config.GetDomain(),
		s.config.GetRepository(),
		workload.IsClusterScoped(),
	)

	// convert the component resource to a v4 resource
	v4ComponentResource := &resource.Resource{
		GVK:        resource.GVK(componentResource.GVK),
		Plural:     componentResource.Plural,
		Path:       componentResource.Path,
		API:        (*resource.API)(componentResource.API),
		Controller: componentResource.Controller,
		Webhooks:   (*resource.Webhooks)(componentResource.Webhooks),
	}

	// override the scaffold if we have a component.  this will allow the Resource
	// attribute of the scaffolder to be set appropriately so that things like Group,
	// Version, and Kind are passed from the child component and not the parent
	// workload.
	if workload.IsComponent() {
		scaffold = machinery.NewScaffold(s.fs,
			machinery.WithConfig(s.config),
			machinery.WithBoilerplate(s.boilerplate),
			machinery.WithResource(v4ComponentResource),
		)
	}

	// inject the resource as this resource so that our PROJECT file is up to date for each
	// resource that we loop through
	if err := s.config.UpdateResource(*v4ComponentResource); err != nil {
		return fmt.Errorf("%w; error updating resource", err)
	}

	// scaffold the workload api.  this generates files within the apis/ folder to include
	// items such as common resource methods, api type definitions and child resource typed
	// object definitions.
	if err := s.scaffoldAPI(scaffold, workload); err != nil {
		return fmt.Errorf("%w; %s", err, ErrScaffoldAPIResources.Error())
	}

	// scaffold the controller.  this generates the main controller logic.
	if err := scaffold.Execute(
		&controller.Controller{Builder: workload},
		&controller.Phases{PackageName: workload.GetPackageName()},
		&controller.SuiteTest{},
		&dependencies.Component{},
		&mutate.Component{},
		&crd.Kustomization{},
	); err != nil {
		return fmt.Errorf("%w; %s", err, ErrScaffoldController.Error())
	}

	// update controller main entrypoint.  this updates the main.go file with logic related to
	// creating the new controllers.
	if err := scaffold.Execute(
		&templates.MainUpdater{
			WireResource:   true,
			WireController: true,
		},
	); err != nil {
		return fmt.Errorf("%w; %s", err, ErrScaffoldMainUpdater.Error())
	}

	// scaffold the custom resource sample files.  this will generate sample manifest files.
	if err := scaffold.Execute(
		&samples.CRDSample{
			SpecFields:      workload.GetAPISpecFields(),
			IsClusterScoped: workload.IsClusterScoped(),
		},
	); err != nil {
		return fmt.Errorf("%w; %s", err, ErrScaffoldCRDSample.Error())
	}

	// scaffold the kustomization sample if OLM is enabled.
	if err := scaffold.Execute(&samples.Kustomization{}); err != nil {
		return fmt.Errorf("%w; %s", err, ErrScaffoldKustomization.Error())
	}

	// scaffold the end-to-end tests.  this will generate some common end-to-end tests for
	// the controller.
	if err := scaffold.Execute(&e2e.WorkloadTest{Builder: workload}); err != nil {
		return fmt.Errorf("%w; %s - error updating test/e2e/%s_%s_%s_test.go", err, ErrScaffoldController.Error(),
			workload.GetAPIGroup(), workload.GetAPIVersion(), strings.ToLower(workload.GetAPIKind()))
	}

	// scaffold the companion CLI only if we have a root command name
	if s.cliRootCommandName != "" {
		if err := s.scaffoldCLI(scaffold, workload); err != nil {
			return fmt.Errorf("%w; %s", err, ErrScaffoldCompanionCLI.Error())
		}
	}

	// scaffold the components of a collection if we have a collection.  this will scaffold the
	// logic for a companion CLI.
	if workload.IsCollection() {
		for _, component := range workload.GetComponents() {
			if err := s.scaffoldWorkload(scaffold, component); err != nil {
				return fmt.Errorf("%w; %s for workload type %T", err, ErrScaffoldWorkload.Error(), component)
			}
		}
	}

	return nil
}

// scaffoldAPI runs the specific logic to scaffold anything existing in the apis directory.
func (s *apiScaffolder) scaffoldAPI(
	scaffold *machinery.Scaffold,
	workload kinds.WorkloadBuilder,
) error {
	// scaffold the base api types
	if err := scaffold.Execute(
		&api.Types{Builder: workload},
		&api.Group{},
	); err != nil {
		return fmt.Errorf("%w; %s", err, ErrScaffoldAPITypes.Error())
	}

	// scaffold the specific kind files
	if err := scaffold.Execute(
		&api.Kind{},
		&api.KindLatest{PackageName: workload.GetPackageName()},
		&api.KindUpdater{},
	); err != nil {
		return fmt.Errorf("%w; %s", err, ErrScaffoldAPIKindInfo.Error())
	}

	// scaffold the resources
	if err := scaffold.Execute(
		&resources.Resources{Builder: workload},
	); err != nil {
		return fmt.Errorf("%w; %s", err, ErrScaffoldAPIResources.Error())
	}

	// scaffolds the child resource definition files
	// these are the resources defined in the static yaml manifests
	for _, manifest := range *workload.GetManifests() {
		if err := scaffold.Execute(
			&resources.Definition{Builder: workload, Manifest: manifest},
		); err != nil {
			return fmt.Errorf("%w; %s", err, ErrScaffoldAPIChildResources.Error())
		}

		// update the child resource mutation for each child resource
		for i := range manifest.ChildResources {
			if err := scaffold.Execute(
				&resources.Mutate{Builder: workload, ChildResource: manifest.ChildResources[i]},
			); err != nil {
				return fmt.Errorf("%w; %s", err, ErrScaffoldAPIChildResources.Error())
			}
		}
	}

	// scaffold the child resource naming helpers
	if err := scaffold.Execute(
		&resources.Constants{Builder: workload},
	); err != nil {
		return fmt.Errorf("%w; %s", err, ErrScaffoldAPIResources.Error())
	}

	return nil
}

// scaffoldCLI runs the specific logic to scaffold the companion CLI for an
// individual workload.
func (s *apiScaffolder) scaffoldCLI(
	scaffold *machinery.Scaffold,
	workload kinds.WorkloadBuilder,
) error {
	// scaffold init subcommand
	if err := scaffold.Execute(
		&cli.CmdInitSub{Builder: workload},
		&cli.CmdInitSubUpdater{Builder: workload},
	); err != nil {
		return fmt.Errorf("%w; %s", err, ErrScaffoldCompanionCLIInit.Error())
	}

	// scaffold the generate command unless we have a collection without resources
	if (workload.HasChildResources() && workload.IsCollection()) || !workload.IsCollection() {
		if err := scaffold.Execute(
			&cli.CmdGenerateSub{Builder: workload},
			&cli.CmdGenerateSubUpdater{Builder: workload},
		); err != nil {
			return fmt.Errorf("%w; %s", err, ErrScaffoldCompanionCLIGenerate.Error())
		}
	}

	// scaffold version subcommand
	if err := scaffold.Execute(
		&cli.CmdVersionSub{Builder: workload},
		&cli.CmdVersionSubUpdater{Builder: workload},
	); err != nil {
		return fmt.Errorf("%w; %s", err, ErrScaffoldCompanionCLIVersion.Error())
	}

	// scaffold the root command
	if err := scaffold.Execute(
		&cli.CmdRootUpdater{
			InitCommand:     true,
			GenerateCommand: true,
			VersionCommand:  true,
			Builder:         workload,
		},
	); err != nil {
		return fmt.Errorf("%w; %s", err, ErrScaffoldCompanionCLIRoot.Error())
	}

	return nil
}
