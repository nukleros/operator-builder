package scaffolds

import (
	"fmt"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"

	"gitlab.eng.vmware.com/landerr/operator-builder/pkg/plugins/workload/v1/scaffolds/templates/cli"
	workloadv1 "gitlab.eng.vmware.com/landerr/operator-builder/pkg/workload/v1"
)

var _ plugins.Scaffolder = &initScaffolder{}

type initScaffolder struct {
	config          config.Config
	boilerplatePath string
	workload        workloadv1.Workload

	fs machinery.Filesystem
}

// NewInitScaffolder returns a new Scaffolder for project initialization operations
func NewInitScaffolder(
	config config.Config,
	workload workloadv1.Workload,
) plugins.Scaffolder {
	return &initScaffolder{
		config:          config,
		boilerplatePath: "hack/boilerplate.go.txt",
		workload:        workload,
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
			CliRootCmd: s.workload.Spec.CompanionCliRootcmd.Name,
		},
		&cli.CliCmdRoot{
			CliRootCmd:         s.workload.Spec.CompanionCliRootcmd.Name,
			CliRootDescription: s.workload.Spec.CompanionCliRootcmd.Description,
		},
	)

}
