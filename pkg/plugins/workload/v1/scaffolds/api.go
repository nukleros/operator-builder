package scaffolds

import (
	"fmt"
	"strings"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/api"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/api/common"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/api/resources"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/cli"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/config/crd"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/config/samples"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/controller"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/controller/phases"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

var _ plugins.Scaffolder = &apiScaffolder{}

type apiScaffolder struct {
	config          config.Config
	resource        resource.Resource
	boilerplatePath string
	workload        workloadv1.WorkloadAPIBuilder
	project         *workloadv1.Project

	fs machinery.Filesystem
}

// NewAPIScaffolder returns a new Scaffolder for project initialization operations
func NewAPIScaffolder(
	config config.Config,
	res resource.Resource,
	workload workloadv1.WorkloadAPIBuilder,
	project *workloadv1.Project,
) plugins.Scaffolder {
	return &apiScaffolder{
		config:          config,
		resource:        res,
		boilerplatePath: "hack/boilerplate.go.txt",
		workload:        workload,
		project:         project,
	}
}

// InjectFS implements cmdutil.Scaffolder
func (s *apiScaffolder) InjectFS(fs machinery.Filesystem) {
	s.fs = fs
}

// scaffold implements cmdutil.Scaffolder
func (s *apiScaffolder) Scaffold() error {
	fmt.Println("Building API...")

	boilerplate, err := afero.ReadFile(s.fs.FS, s.boilerplatePath)
	if err != nil {
		return err
	}

	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
		machinery.WithBoilerplate(string(boilerplate)),
		machinery.WithResource(&s.resource),
	)

	var createFuncNames []string
	for _, sourceFile := range *s.workload.GetSourceFiles() {
		for _, childResource := range sourceFile.Children {
			funcName := fmt.Sprintf("Create%s", childResource.UniqueName)
			createFuncNames = append(createFuncNames, funcName)
		}
	}

	var crdSampleFilenames []string

	// companion CLI
	if s.workload.IsStandalone() && s.workload.GetRootcommandName() != "" {
		// build a subcommand for standalone, e.g. `webstorectl init`
		if err = scaffold.Execute(
			&cli.CliCmdInitSub{
				CliRootCmd:        s.project.CliRootCommandName,
				CliSubCmdName:     s.workload.GetSubcommandName(),
				CliSubCmdDescr:    s.workload.GetSubcommandDescr(),
				CliSubCmdVarName:  s.workload.GetSubcommandVarName(),
				CliSubCmdFileName: s.workload.GetSubcommandFileName(),
				SpecFields:        s.workload.GetAPISpecFields(),
				IsComponent:       s.workload.IsComponent(),
				ComponentResource: s.workload.GetComponentResource(
					s.config.GetDomain(),
					s.config.GetRepository(),
					s.workload.IsClusterScoped(),
				),
			},
			&cli.CliCmdGenerateSub{
				PackageName:    s.workload.GetPackageName(),
				CliRootCmd:     s.project.CliRootCommandName,
				CliSubCmdName:  s.workload.GetSubcommandName(),
				CliSubCmdDescr: s.workload.GetSubcommandDescr(),
				IsComponent:    s.workload.IsComponent(),
			},
		); err != nil {
			return err
		}
	} else if s.workload.IsCollection() && s.workload.GetRootcommandName() != "" {
		if err = scaffold.Execute(
			&cli.CliCmdInit{
				CliRootCmd: s.project.CliRootCommandName,
			},
			&cli.CliCmdGenerate{
				CliRootCmd: s.project.CliRootCommandName,
			},
		); err != nil {
			return err
		}
		for _, component := range *s.workload.GetComponents() {

			if component.GetSubcommandName() != "" {
				// build a subcommand for the component, e.g. `cnpctl init ingress`
				if err = scaffold.Execute(
					&cli.CliCmdInitSub{
						CliRootCmd:        s.project.CliRootCommandName,
						CliSubCmdName:     component.GetSubcommandName(),
						CliSubCmdDescr:    component.GetSubcommandDescr(),
						CliSubCmdVarName:  component.GetSubcommandVarName(),
						CliSubCmdFileName: component.GetSubcommandFileName(),
						SpecFields:        component.GetAPISpecFields(),
						IsComponent:       component.IsComponent(),
						ComponentResource: component.GetComponentResource(
							s.config.GetDomain(),
							s.config.GetRepository(),
							s.workload.IsClusterScoped(),
						),
					},
					&cli.CliCmdGenerateSub{
						PackageName:       component.GetPackageName(),
						CliRootCmd:        s.project.CliRootCommandName,
						CliSubCmdName:     component.GetSubcommandName(),
						CliSubCmdDescr:    component.GetSubcommandDescr(),
						CliSubCmdVarName:  component.GetSubcommandVarName(),
						CliSubCmdFileName: component.GetSubcommandFileName(),
						IsComponent:       component.IsComponent(),
						ComponentResource: component.GetComponentResource(
							s.config.GetDomain(),
							s.config.GetRepository(),
							s.workload.IsClusterScoped(),
						),
					},
				); err != nil {
					return err
				}
			}
		}
	}

	// API types
	if s.workload.IsStandalone() {
		if err = scaffold.Execute(
			&api.Types{
				SpecFields:    s.workload.GetAPISpecFields(),
				ClusterScoped: s.workload.IsClusterScoped(),
				Dependencies:  s.workload.GetDependencies(),
			},
			&common.Components{},
			&common.Conditions{},
			&resources.Resources{
				PackageName:     s.workload.GetPackageName(),
				CreateFuncNames: createFuncNames,
				SpecFields:      s.workload.GetAPISpecFields(),
			},
			&controller.Controller{
				PackageName:       s.workload.GetPackageName(),
				RBACRules:         s.workload.GetRBACRules(),
				HasChildResources: s.workload.HasChildResources(),
			},
			&controller.Common{},
			&phases.Types{},
			&phases.Common{},
			&phases.CreateResource{},
			&phases.ResourcePersist{},
			&phases.ResourceCreateInMemory{
				HasChildResources: s.workload.HasChildResources(),
			},
			&samples.CRDSample{
				SpecFields: s.workload.GetAPISpecFields(),
			},
		); err != nil {
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

		if err = scaffold.Execute(
			&api.Types{
				SpecFields:    s.workload.GetAPISpecFields(),
				ClusterScoped: s.workload.IsClusterScoped(),
				Dependencies:  s.workload.GetDependencies(),
			},
			&common.Components{},
			&common.Conditions{},
			&controller.Controller{
				PackageName:       s.workload.GetPackageName(),
				RBACRules:         &[]workloadv1.RBACRule{},
				HasChildResources: s.workload.HasChildResources(),
			},
			&crd.Kustomization{
				CRDSampleFilenames: crdSampleFilenames,
			},
		); err != nil {
			return err
		}
		for _, component := range *s.workload.GetComponents() {

			componentScaffold := machinery.NewScaffold(s.fs,
				machinery.WithConfig(s.config),
				machinery.WithBoilerplate(string(boilerplate)),
				machinery.WithResource(component.GetComponentResource(
					s.config.GetDomain(),
					s.config.GetRepository(),
					component.IsClusterScoped(),
				)),
			)

			var createFuncNames []string
			for _, sourceFile := range *component.GetSourceFiles() {
				for _, childResource := range sourceFile.Children {
					funcName := fmt.Sprintf("Create%s", childResource.UniqueName)
					createFuncNames = append(createFuncNames, funcName)
				}
			}

			crdSampleFilenames = append(
				crdSampleFilenames,
				strings.ToLower(fmt.Sprintf(
					"%s.%s_%s.yaml",
					component.GetAPIGroup(),
					s.workload.GetDomain(),
					utils.PluralizeKind(component.GetAPIKind()),
				)),
			)

			if err = componentScaffold.Execute(
				&templates.MainUpdater{
					WireResource:   true,
					WireController: true,
				},
				&api.Types{
					SpecFields:    &component.Spec.APISpecFields,
					ClusterScoped: component.IsClusterScoped(),
					Dependencies:  component.GetDependencies(),
				},
				&api.Group{},
				&resources.Resources{
					PackageName:     component.GetPackageName(),
					CreateFuncNames: createFuncNames,
					SpecFields:      component.GetAPISpecFields(),
				},
				&controller.Controller{
					PackageName:       component.GetPackageName(),
					RBACRules:         component.GetRBACRules(),
					HasChildResources: component.HasChildResources(),
				},
				&controller.Common{},
				&phases.Types{},
				&phases.Common{},
				&phases.CreateResource{},
				&phases.ResourcePersist{},
				&phases.ResourceCreateInMemory{
					HasChildResources: component.HasChildResources(),
				},
				&samples.CRDSample{
					SpecFields: &component.Spec.APISpecFields,
				},
				&crd.Kustomization{
					CRDSampleFilenames: crdSampleFilenames,
				},
			); err != nil {
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

				if err = scaffold.Execute(
					&resources.Definition{
						ClusterScoped: component.IsClusterScoped(),
						SourceFile:    sourceFile,
						PackageName:   component.GetPackageName(),
					},
				); err != nil {
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
			machinery.WithResource(&s.resource),
		)

		if err = scaffold.Execute(
			&resources.Definition{
				ClusterScoped: s.workload.IsClusterScoped(),
				SourceFile:    sourceFile,
				PackageName:   s.workload.GetPackageName(),
			},
		); err != nil {
			return err
		}
	}

	return nil
}
