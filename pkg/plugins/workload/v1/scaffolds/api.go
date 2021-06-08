package scaffolds

import (
	"fmt"
	"strings"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/api"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/api/common"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/api/resources"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/cli"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/config/samples"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/controller"
	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/controller/phases"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

var _ plugins.Scaffolder = &apiScaffolder{}

type apiScaffolder struct {
	config          config.Config
	resource        resource.Resource
	boilerplatePath string
	workload        workloadv1.WorkloadAPIBuilder
	workloadPath    string
	apiSpecFields   *[]workloadv1.APISpecField
	sourceFiles     *[]workloadv1.SourceFile
	rbacRules       *[]workloadv1.RBACRule
	project         *workloadv1.Project

	fs machinery.Filesystem
}

// NewAPIScaffolder returns a new Scaffolder for project initialization operations
func NewAPIScaffolder(
	config config.Config,
	res resource.Resource,
	workload workloadv1.WorkloadAPIBuilder,
	workloadPath string,
	apiSpecFields *[]workloadv1.APISpecField,
	sourceFiles *[]workloadv1.SourceFile,
	rbacRules *[]workloadv1.RBACRule,
	project *workloadv1.Project,
) plugins.Scaffolder {
	return &apiScaffolder{
		config:          config,
		resource:        res,
		boilerplatePath: "hack/boilerplate.go.txt",
		workload:        workload,
		workloadPath:    workloadPath,
		apiSpecFields:   apiSpecFields,
		sourceFiles:     sourceFiles,
		rbacRules:       rbacRules,
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

	// Initialize the machinery.Scaffold that will write the files to disk
	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
		machinery.WithBoilerplate(string(boilerplate)),
		machinery.WithResource(&s.resource),
	)

	packageName := strings.ToLower(strings.Replace(s.workload.GetName(), "-", "_", 0))

	var createFuncNames []string
	for _, sourceFile := range *s.sourceFiles {
		for _, childResource := range sourceFile.Children {
			funcName := fmt.Sprintf("Create%s", childResource.UniqueName)
			createFuncNames = append(createFuncNames, funcName)
		}
	}

	// companion CLI subcommands
	if s.workload.GetSubcommandName() != "" {
		// build a subcommand for the component, e.g. `cnpctl init ingress`
		if err = scaffold.Execute(
			&cli.CliCmdInit{
				CliRootCmd: s.project.CliRootCommandName,
			},
			&cli.CliCmdInitSub{
				CliRootCmd:     s.project.CliRootCommandName,
				CliSubCmdName:  s.workload.GetSubcommandName(),
				CliSubCmdDescr: s.workload.GetSubcommandDescr(),
				SpecFields:     s.apiSpecFields,
			},
			&cli.CliCmdGenerate{
				CliRootCmd: s.project.CliRootCommandName,
			},
			&cli.CliCmdGenerateSub{
				CliRootCmd:     s.project.CliRootCommandName,
				CliSubCmdName:  s.workload.GetSubcommandName(),
				CliSubCmdDescr: s.workload.GetSubcommandDescr(),
				PackageName:    packageName,
			},
		); err != nil {
			return err
		}
	} else if s.workload.GetRootcommandName() != "" {
		// build a subcommand for standalone, e.g. `webappctl init`
		if err = scaffold.Execute(
			&cli.CliCmdInitSub{
				CliRootCmd:     s.project.CliRootCommandName,
				CliSubCmdName:  s.workload.GetSubcommandName(),
				CliSubCmdDescr: s.workload.GetSubcommandDescr(),
				SpecFields:     s.apiSpecFields,
				Component:      s.workload.IsComponent(),
			},
			&cli.CliCmdGenerateSub{
				CliRootCmd:     s.project.CliRootCommandName,
				CliSubCmdName:  s.workload.GetSubcommandName(),
				CliSubCmdDescr: s.workload.GetSubcommandDescr(),
				Component:      s.workload.IsComponent(),
				PackageName:    packageName,
			},
		); err != nil {
			return err
		}
	}

	// API types
	if !s.workload.IsComponent() {
		if err = scaffold.Execute(
			&api.Types{
				SpecFields:    s.apiSpecFields,
				ClusterScoped: s.workload.IsClusterScoped(),
				Dependencies:  s.workload.GetDependencies(),
			},
			&common.Components{},
			&common.Conditions{},
			&resources.Resources{
				PackageName:     packageName,
				CreateFuncNames: createFuncNames,
				SpecFields:      s.apiSpecFields,
			},
			&controller.Controller{
				PackageName: packageName,
				RBACRules:   s.rbacRules,
			},
			&controller.Common{},
			&phases.Types{},
			&phases.Common{},
			&phases.CreateResource{},
			&phases.ResourcePersist{},
			&phases.ResourceCreateInMemory{},
			&samples.CRDSample{
				SpecFields: s.apiSpecFields,
			},
		); err != nil {
			return err
		}
	} else {
		// TODO: build API for collections
		return nil
	}

	// resource definition files
	// these are the resources defined in the static yaml manifests
	for _, sourceFile := range *s.sourceFiles {

		scaffold := machinery.NewScaffold(s.fs,
			machinery.WithConfig(s.config),
			machinery.WithBoilerplate(string(boilerplate)),
			machinery.WithResource(&s.resource),
		)

		if err = scaffold.Execute(
			&resources.Definition{
				ClusterScoped: s.workload.IsClusterScoped(),
				SourceFile:    sourceFile,
				PackageName:   packageName,
			},
		); err != nil {
			return err
		}
	}

	return nil
}
