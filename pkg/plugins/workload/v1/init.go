package v1

import (
	"fmt"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	"gitlab.eng.vmware.com/landerr/operator-builder/pkg/plugins/workload/v1/scaffolds"
	workloadv1 "gitlab.eng.vmware.com/landerr/operator-builder/pkg/workload/v1"
)

type initSubcommand struct {
	config      config.Config
	commandName string

	standaloneWorkloadConfigPath string
	workloadCollectionConfigPath string

	workload workloadv1.WorkloadInitializer
}

var _ plugin.InitSubcommand = &initSubcommand{}

func (p *initSubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {

	subcmdMeta.Description = `Add workload management scaffolding to a new project
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Add scaffolding defined by a standalone workload config file
  %[1]s init --standalone-workload-config .source-manifests/workload.yaml
`, cliMeta.CommandName)
}

func (p *initSubcommand) BindFlags(fs *pflag.FlagSet) {

	fs.StringVar(&p.standaloneWorkloadConfigPath, "standalone-workload-config", "", "path to standalone workload config file")
	fs.StringVar(&p.workloadCollectionConfigPath, "workload-collection-config", "", "path to workload collection config file")
}

func (p *initSubcommand) InjectConfig(c config.Config) error {
	p.config = c

	// operator builder always uses multi-group APIs
	if err := c.SetMultiGroup(); err != nil {
		return err
	}

	return nil
}

func (p *initSubcommand) PreScaffold(machinery.Filesystem) error {

	// process workload config file
	workload, err := workloadv1.ProcessInitConfig(
		p.standaloneWorkloadConfigPath,
		p.workloadCollectionConfigPath,
	)
	if err != nil {
		return err
	}
	p.workload = workload

	return nil
}

func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {

	scaffolder := scaffolds.NewInitScaffolder(p.config, p.workload)
	scaffolder.InjectFS(fs)
	err := scaffolder.Scaffold()
	if err != nil {
		return err
	}

	return nil
}
