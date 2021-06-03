package scaffolds

import (
	"fmt"
	"strings"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"

	"gitlab.eng.vmware.com/landerr/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/api"
	"gitlab.eng.vmware.com/landerr/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/api/common"
	"gitlab.eng.vmware.com/landerr/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/api/resources"
	"gitlab.eng.vmware.com/landerr/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/cli"
	"gitlab.eng.vmware.com/landerr/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/config/samples"
	"gitlab.eng.vmware.com/landerr/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/controller"
	"gitlab.eng.vmware.com/landerr/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/controller/phases"
	workloadv1 "gitlab.eng.vmware.com/landerr/operator-builder/pkg/workload/v1"
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
	specFields, err := s.workload.GetSpecFields(s.workloadPath)
	if err != nil {
		return err
	}

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
				SpecFields:     specFields,
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
				SpecFields:     specFields,
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
				SpecFields:    specFields,
				ClusterScoped: s.workload.IsClusterScoped(),
				Dependencies:  s.workload.GetDependencies(),
			},
			&common.Components{},
			&common.Conditions{},
			&resources.Resources{
				PackageName:     packageName,
				CreateFuncNames: createFuncNames,
				SpecFields:      specFields,
			},
			&controller.Controller{
				PackageName: packageName,
			},
			&controller.Common{},
			&phases.Types{},
			&phases.Common{},
			&phases.CreateResource{},
			&phases.ResourcePersist{},
			&phases.ResourceCreateInMemory{},
			&samples.CRDSample{
				SpecFields: specFields,
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
