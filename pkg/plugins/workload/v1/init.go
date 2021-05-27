package v1

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/yaml"

	"gitlab.eng.vmware.com/landerr/operator-builder/pkg/plugins/workload/v1/scaffolds"
	workloadv1 "gitlab.eng.vmware.com/landerr/operator-builder/pkg/workload/v1"
)

type initSubcommand struct {
	config      config.Config
	commandName string

	workloadConfigPath string
	workloadConfig     workloadv1.WorkloadConfig
}

var _ plugin.InitSubcommand = &initSubcommand{}

func (p *initSubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {

	subcmdMeta.Description = `Add workload management scaffolding to a new project
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Add scaffolding defined by a workload config file
  %[1]s init --workload-config .source-manifests/workload.yaml
`, cliMeta.CommandName)
}

func (p *initSubcommand) BindFlags(fs *pflag.FlagSet) {

	fs.StringVar(&p.workloadConfigPath, "workload-config", "", "path to workload config file")
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

	// unmarshal config file to WorkloadConfig
	config, err := ioutil.ReadFile(p.workloadConfigPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(config, &p.workloadConfig)
	if err != nil {
		return err
	}

	// validate WorkloadConfig
	if err := p.workloadConfig.Validate(); err != nil {
		return err
	}

	return nil
}

func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {

	scaffolder := scaffolds.NewInitScaffolder(p.config, p.workloadConfig)
	scaffolder.InjectFS(fs)
	err := scaffolder.Scaffold()
	if err != nil {
		return err
	}

	return nil
}
