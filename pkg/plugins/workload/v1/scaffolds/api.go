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
	workloadConfig  *workloadv1.WorkloadConfig
	apiSpecFields   *[]workloadv1.APISpecField
	sourceFiles     *[]workloadv1.SourceFile

	fs machinery.Filesystem
}

// NewAPIScaffolder returns a new Scaffolder for project initialization operations
func NewAPIScaffolder(
	config config.Config,
	res resource.Resource,
	workloadConfig *workloadv1.WorkloadConfig,
	apiSpecFields *[]workloadv1.APISpecField,
	sourceFiles *[]workloadv1.SourceFile,
) plugins.Scaffolder {
	return &apiScaffolder{
		config:          config,
		resource:        res,
		boilerplatePath: "hack/boilerplate.go.txt",
		workloadConfig:  workloadConfig,
		apiSpecFields:   apiSpecFields,
		sourceFiles:     sourceFiles,
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

	packageName := strings.ToLower(strings.Replace(s.workloadConfig.Name, "-", "_", -1))

	var createFuncNames []string
	for _, sourceFile := range *s.sourceFiles {
		for _, childResource := range sourceFile.Children {
			funcName := fmt.Sprintf("Create%s", childResource.UniqueName)
			createFuncNames = append(createFuncNames, funcName)
		}
	}

	// API types
	if !s.workloadConfig.Spec.Collection {
		if err = scaffold.Execute(
			&api.Types{
				SpecFields:    s.apiSpecFields,
				ClusterScoped: s.workloadConfig.Spec.ClusterScoped,
				Dependencies:  s.workloadConfig.Spec.Dependencies,
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
				ClusterScoped: s.workloadConfig.Spec.ClusterScoped,
				SourceFile:    sourceFile,
				PackageName:   packageName,
			},
		); err != nil {
			return err
		}
	}

	return nil
}
