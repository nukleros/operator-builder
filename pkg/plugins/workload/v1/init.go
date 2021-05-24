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

	// workload
	workloadPath string
	//workload     workloadv1.Workload
	workload workloadv1.Workload
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

	fs.StringVar(&p.workloadPath, "workload-config", "", "path to workload config file")
}

func (p *initSubcommand) InjectConfig(c config.Config) error {
	p.config = c
	if err := c.SetMultiGroup(); err != nil {
		return err
	}

	return nil
}

func (p *initSubcommand) PreScaffold(machinery.Filesystem) error {

	return nil
}

func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {

	// unmarshal config file to Workload
	config, err := ioutil.ReadFile(p.workloadPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(config, &p.workload)
	if err != nil {
		return err
	}

	// validate Workload config
	if err := p.workload.Validate(); err != nil {
		return err
	}

	scaffolder := scaffolds.NewInitScaffolder(p.config, p.workload)
	scaffolder.InjectFS(fs)
	err = scaffolder.Scaffold()
	if err != nil {
		return err
	}

	return nil
}
