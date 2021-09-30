// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package scaffolds

import (
	"fmt"
	"strings"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"

	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/api"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/api/common"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/api/resources"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/cli"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/config/crd"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/config/samples"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/controller"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/int/controllers/phases"
	controllersutils "github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/int/controllers/utils"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/int/dependencies"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/int/helpers"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/int/mutate"
	resourcespkg "github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/int/resources"
	"github.com/vmware-tanzu-labs/operator-builder/internal/plugins/workload/v1/scaffolds/templates/int/wait"
	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

var _ plugins.Scaffolder = &apiScaffolder{}

type apiScaffolder struct {
	config             config.Config
	resource           *resource.Resource
	boilerplatePath    string
	workload           workloadv1.WorkloadAPIBuilder
	cliRootCommandName string

	fs machinery.Filesystem
}

// NewAPIScaffolder returns a new Scaffolder for project initialization operations.
func NewAPIScaffolder(
	cfg config.Config,
	res *resource.Resource,
	workload workloadv1.WorkloadAPIBuilder,
	cliRootCommandName string,
) plugins.Scaffolder {
	return &apiScaffolder{
		config:             cfg,
		resource:           res,
		boilerplatePath:    "hack/boilerplate.go.txt",
		workload:           workload,
		cliRootCommandName: cliRootCommandName,
	}
}

// InjectFS implements cmdutil.Scaffolder.
func (s *apiScaffolder) InjectFS(fs machinery.Filesystem) {
	s.fs = fs
}

// scaffold implements cmdutil.Scaffolder.
func (s *apiScaffolder) Scaffold() error {
	fmt.Println("Building API...")

	boilerplate, err := afero.ReadFile(s.fs.FS, s.boilerplatePath)
	if err != nil {
		return err
	}

	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
		machinery.WithBoilerplate(string(boilerplate)),
		machinery.WithResource(s.resource),
	)

	createFuncNames, initFuncNames := s.workload.GetFuncNames()

	var crdSampleFilenames []string

	// companion CLI
	if s.workload.IsStandalone() && s.workload.GetRootcommandName() != "" {
		// build a subcommand for standalone, e.g. `webstorectl init`
		err = scaffold.Execute(
			&cli.CmdInitSub{
				RootCmd:        s.cliRootCommandName,
				RootCmdVarName: utils.ToPascalCase(s.cliRootCommandName),
				SubCmdName:     s.workload.GetSubcommandName(),
				SubCmdDescr:    s.workload.GetSubcommandDescr(),
				SubCmdVarName:  s.workload.GetSubcommandVarName(),
				SubCmdFileName: s.workload.GetSubcommandFileName(),
				SpecFields:     s.workload.GetAPISpecFields(),
				IsComponent:    s.workload.IsComponent(),
				ComponentResource: s.workload.GetComponentResource(
					s.config.GetDomain(),
					s.config.GetRepository(),
					s.workload.IsClusterScoped(),
				),
			},
			&cli.CmdGenerateSub{
				PackageName:    s.workload.GetPackageName(),
				RootCmd:        s.cliRootCommandName,
				RootCmdVarName: utils.ToPascalCase(s.cliRootCommandName),
				SubCmdName:     s.workload.GetSubcommandName(),
				SubCmdDescr:    s.workload.GetSubcommandDescr(),
				IsComponent:    s.workload.IsComponent(),
			},
		)
		if err != nil {
			return err
		}
	} else if s.workload.IsCollection() && s.workload.GetRootcommandName() != "" {
		err = scaffold.Execute(
			&cli.CmdInit{
				RootCmd:        s.cliRootCommandName,
				RootCmdVarName: utils.ToPascalCase(s.cliRootCommandName),
				Collection:     s.workload.(*workloadv1.WorkloadCollection),
			},
			&cli.CmdGenerate{
				RootCmd:        s.cliRootCommandName,
				RootCmdVarName: utils.ToPascalCase(s.cliRootCommandName),
				Collection:     s.workload.(*workloadv1.WorkloadCollection),
			},
		)
		if err != nil {
			return err
		}

		for _, component := range s.workload.GetComponents() {
			if component.GetSubcommandName() != "" {
				// build a subcommand for the component, e.g. `cnpctl init ingress`
				err = scaffold.Execute(
					&cli.CmdInitSub{
						RootCmd:        s.cliRootCommandName,
						RootCmdVarName: utils.ToPascalCase(s.cliRootCommandName),
						SubCmdName:     component.GetSubcommandName(),
						SubCmdDescr:    component.GetSubcommandDescr(),
						SubCmdVarName:  component.GetSubcommandVarName(),
						SubCmdFileName: component.GetSubcommandFileName(),
						SpecFields:     component.GetAPISpecFields(),
						IsComponent:    component.IsComponent(),
						ComponentResource: component.GetComponentResource(
							s.config.GetDomain(),
							s.config.GetRepository(),
							s.workload.IsClusterScoped(),
						),
					},
					&cli.CmdGenerateSub{
						PackageName:    component.GetPackageName(),
						RootCmd:        s.cliRootCommandName,
						RootCmdVarName: utils.ToPascalCase(s.cliRootCommandName),
						SubCmdName:     component.GetSubcommandName(),
						SubCmdDescr:    component.GetSubcommandDescr(),
						SubCmdVarName:  component.GetSubcommandVarName(),
						SubCmdFileName: component.GetSubcommandFileName(),
						IsComponent:    component.IsComponent(),
						ComponentResource: component.GetComponentResource(
							s.config.GetDomain(),
							s.config.GetRepository(),
							s.workload.IsClusterScoped(),
						),
						Collection: s.workload.(*workloadv1.WorkloadCollection),
					},
				)
				if err != nil {
					return err
				}
			}
		}
	}

	err = scaffold.Execute(
		&cli.CmdRootUpdater{
			RootCmd:         s.cliRootCommandName,
			InitCommand:     true,
			GenerateCommand: true,
		},
	)
	if err != nil {
		return fmt.Errorf("error updating root.go: %v", err)
	}

	// API types
	if s.workload.IsStandalone() {
		err = scaffold.Execute(
			&templates.MainUpdater{
				WireResource:   true,
				WireController: true,
			},
			&api.Types{
				SpecFields:    s.workload.GetAPISpecFields(),
				ClusterScoped: s.workload.IsClusterScoped(),
				Dependencies:  s.workload.GetDependencies(),
				IsStandalone:  s.workload.IsStandalone(),
			},
			&common.Components{
				IsStandalone: s.workload.IsStandalone(),
			},
			&common.Conditions{},
			&common.Resources{},
			&resources.Resources{
				PackageName:     s.workload.GetPackageName(),
				CreateFuncNames: createFuncNames,
				InitFuncNames:   initFuncNames,
				IsComponent:     s.workload.IsComponent(),
			},
			&resourcespkg.ResourceType{},
			&resourcespkg.Resources{},
			&resourcespkg.NamespaceType{},
			&resourcespkg.ConfigMapType{},
			&resourcespkg.CustomResourceDefinitionType{},
			&resourcespkg.DaemonSetType{},
			&resourcespkg.DeploymentType{},
			&resourcespkg.StatefulSetType{},
			&resourcespkg.JobType{},
			&resourcespkg.SecretType{},
			&resourcespkg.ServiceType{},
			&controller.Controller{
				PackageName:       s.workload.GetPackageName(),
				RBACRules:         s.workload.GetRBACRules(),
				OwnershipRules:    s.workload.GetOwnershipRules(),
				HasChildResources: s.workload.HasChildResources(),
				IsStandalone:      s.workload.IsStandalone(),
				IsComponent:       s.workload.IsComponent(),
			},
			&controllersutils.Utils{
				IsStandalone: s.workload.IsStandalone(),
			},
			&controllersutils.RateLimiter{},
			&phases.Types{},
			&phases.Common{},
			&phases.CreateResource{
				IsStandalone: s.workload.IsStandalone(),
			},
			&phases.ResourcePersist{},
			&phases.Dependencies{},
			&phases.PreFlight{},
			&phases.ResourceWait{},
			&phases.CheckReady{},
			&phases.Complete{},
			&helpers.Common{},
			&helpers.Component{},
			&dependencies.Component{},
			&mutate.Component{},
			&wait.Component{},
			&samples.CRDSample{
				SpecFields: s.workload.GetAPISpecFields(),
			},
		)
		if err != nil {
			return err
		}
	} else {
		// collection API
		crdSampleFilenames = append(
			crdSampleFilenames,
			strings.ToLower(fmt.Sprintf(
				"%s.%s_%s.yaml",
				s.workload.GetAPIGroup(),
				s.workload.GetDomain(),
				utils.PluralizeKind(s.workload.GetAPIKind()),
			)),
		)

		err = scaffold.Execute(
			&templates.MainUpdater{
				WireResource:   true,
				WireController: true,
			},
			&api.Types{
				SpecFields:    s.workload.GetAPISpecFields(),
				ClusterScoped: s.workload.IsClusterScoped(),
				Dependencies:  s.workload.GetDependencies(),
				IsStandalone:  s.workload.IsStandalone(),
			},
			&common.Components{
				IsStandalone: s.workload.IsStandalone(),
			},
			&common.Conditions{},
			&common.Resources{},
			&resources.Resources{
				PackageName:     s.workload.GetPackageName(),
				CreateFuncNames: createFuncNames,
				InitFuncNames:   initFuncNames,
				IsComponent:     s.workload.IsComponent(),
			},
			&resourcespkg.ResourceType{},
			&resourcespkg.Resources{},
			&resourcespkg.NamespaceType{},
			&resourcespkg.ConfigMapType{},
			&resourcespkg.CustomResourceDefinitionType{},
			&resourcespkg.DaemonSetType{},
			&resourcespkg.DeploymentType{},
			&resourcespkg.StatefulSetType{},
			&resourcespkg.JobType{},
			&resourcespkg.SecretType{},
			&resourcespkg.ServiceType{},
			&controller.Controller{
				PackageName:       s.workload.GetPackageName(),
				RBACRules:         &[]workloadv1.RBACRule{},
				OwnershipRules:    s.workload.GetOwnershipRules(),
				HasChildResources: s.workload.HasChildResources(),
				IsStandalone:      s.workload.IsStandalone(),
				IsComponent:       s.workload.IsComponent(),
			},
			&controllersutils.Utils{
				IsStandalone: s.workload.IsStandalone(),
			},
			&controllersutils.RateLimiter{},
			&phases.Types{},
			&phases.Common{},
			&phases.CreateResource{
				IsStandalone: s.workload.IsStandalone(),
			},
			&phases.ResourcePersist{},
			&phases.Dependencies{},
			&phases.PreFlight{},
			&phases.ResourceWait{},
			&phases.CheckReady{},
			&phases.Complete{},
			&helpers.Common{},
			&helpers.Component{},
			&dependencies.Component{},
			&mutate.Component{},
			&wait.Component{},
			&samples.CRDSample{
				SpecFields: s.workload.GetAPISpecFields(),
			},
			&crd.Kustomization{
				CRDSampleFilenames: crdSampleFilenames,
			},
		)
		if err != nil {
			return err
		}

		for _, component := range s.workload.GetComponents() {
			componentScaffold := machinery.NewScaffold(s.fs,
				machinery.WithConfig(s.config),
				machinery.WithBoilerplate(string(boilerplate)),
				machinery.WithResource(component.GetComponentResource(
					s.config.GetDomain(),
					s.config.GetRepository(),
					component.IsClusterScoped(),
				)),
			)

			createFuncNames, initFuncNames := component.GetFuncNames()

			crdSampleFilenames = append(
				crdSampleFilenames,
				strings.ToLower(fmt.Sprintf(
					"%s.%s_%s.yaml",
					component.GetAPIGroup(),
					s.workload.GetDomain(),
					utils.PluralizeKind(component.GetAPIKind()),
				)),
			)

			err = componentScaffold.Execute(
				&templates.MainUpdater{
					WireResource:   true,
					WireController: true,
				},
				&api.Types{
					SpecFields:    component.Spec.APISpecFields,
					ClusterScoped: component.IsClusterScoped(),
					Dependencies:  component.GetDependencies(),
					IsStandalone:  component.IsStandalone(),
				},
				&api.Group{},
				&resources.Resources{
					PackageName:     component.GetPackageName(),
					CreateFuncNames: createFuncNames,
					InitFuncNames:   initFuncNames,
					IsComponent:     component.IsComponent(),
					Collection:      s.workload.(*workloadv1.WorkloadCollection),
				},
				&controller.Controller{
					PackageName:       component.GetPackageName(),
					RBACRules:         component.GetRBACRules(),
					OwnershipRules:    component.GetOwnershipRules(),
					HasChildResources: component.HasChildResources(),
					IsStandalone:      component.IsStandalone(),
					IsComponent:       component.IsComponent(),
					Collection:        s.workload.(*workloadv1.WorkloadCollection),
				},
				&dependencies.Component{},
				&mutate.Component{},
				&helpers.Component{},
				&wait.Component{},
				&samples.CRDSample{
					SpecFields: component.Spec.APISpecFields,
				},
				&crd.Kustomization{
					CRDSampleFilenames: crdSampleFilenames,
				},
			)
			if err != nil {
				return err
			}

			// component child resource definition files
			// these are the resources defined in the static yaml manifests
			for _, sourceFile := range *component.GetSourceFiles() {
				scaffold := machinery.NewScaffold(s.fs,
					machinery.WithConfig(s.config),
					machinery.WithBoilerplate(string(boilerplate)),
					machinery.WithResource(component.GetComponentResource(
						s.config.GetDomain(),
						s.config.GetRepository(),
						component.IsClusterScoped(),
					)),
				)

				err = scaffold.Execute(
					&resources.Definition{
						ClusterScoped: component.IsClusterScoped(),
						SourceFile:    sourceFile,
						PackageName:   component.GetPackageName(),
						SpecFields:    component.GetAPISpecFields(),
						IsComponent:   component.IsComponent(),
						Collection:    s.workload.(*workloadv1.WorkloadCollection),
					},
				)
				if err != nil {
					return err
				}
			}
		}
	}

	// child resource definition files
	// these are the resources defined in the static yaml manifests
	for _, sourceFile := range *s.workload.GetSourceFiles() {
		scaffold := machinery.NewScaffold(s.fs,
			machinery.WithConfig(s.config),
			machinery.WithBoilerplate(string(boilerplate)),
			machinery.WithResource(s.resource),
		)

		err = scaffold.Execute(
			&resources.Definition{
				ClusterScoped: s.workload.IsClusterScoped(),
				SourceFile:    sourceFile,
				PackageName:   s.workload.GetPackageName(),
				SpecFields:    s.workload.GetAPISpecFields(),
				IsComponent:   s.workload.IsComponent(),
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}
