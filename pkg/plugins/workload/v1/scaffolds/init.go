package scaffolds

import (
	"fmt"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3/scaffolds"

	"gitlab.eng.vmware.com/landerr/operator-builder/pkg/plugins/workload/v1/scaffolds/templates"
	"gitlab.eng.vmware.com/landerr/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/cli"
	workloadv1 "gitlab.eng.vmware.com/landerr/operator-builder/pkg/workload/v1"
)

const CobraVersion = "v1.1.3"

var _ plugins.Scaffolder = &initScaffolder{}

type initScaffolder struct {
	config          config.Config
	boilerplatePath string
	workloadConfig  workloadv1.WorkloadConfig

	fs machinery.Filesystem
}

// NewInitScaffolder returns a new Scaffolder for project initialization operations
func NewInitScaffolder(config config.Config, workloadConfig workloadv1.WorkloadConfig) plugins.Scaffolder {
	return &initScaffolder{
		config:          config,
		boilerplatePath: "hack/boilerplate.go.txt",
		workloadConfig:  workloadConfig,
	}
}

// InjectFS implements cmdutil.Scaffolder
func (s *initScaffolder) InjectFS(fs machinery.Filesystem) {
	s.fs = fs
}

// scaffold implements cmdutil.Scaffolder
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

	return scaffold.Execute(
		&cli.CliMain{
			CliRootCmd: s.workloadConfig.Spec.CompanionCliRootcmd.Name,
		},
		&cli.CliCmdRoot{
			CliRootCmd:         s.workloadConfig.Spec.CompanionCliRootcmd.Name,
			CliRootDescription: s.workloadConfig.Spec.CompanionCliRootcmd.Description,
		},
		&templates.GoMod{
			ControllerRuntimeVersion: scaffolds.ControllerRuntimeVersion,
			CobraVersion:             CobraVersion,
		},
	)

}
