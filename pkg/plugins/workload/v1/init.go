package v1

import (
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

type initSubcommand struct {
	config      config.Config
	commandName string

	workloadConfigPath string

	workload workloadv1.WorkloadInitializer
}

var _ plugin.InitSubcommand = &initSubcommand{}

func (p *initSubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {

	subcmdMeta.Description = `Add workload management scaffolding to a new project
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Add project scaffolding defined by a workload config file
  %[1]s init --workload-config .source-manifests/workload.yaml
`, cliMeta.CommandName)
}

func (p *initSubcommand) InjectConfig(c config.Config) error {

	p.config = c

	// operator builder always uses multi-group APIs
	if err := c.SetMultiGroup(); err != nil {
		return err
	}

	var taxi workloadv1.ConfigTaxi
	if err := c.DecodePluginConfig(workloadv1.ConfigTaxiKey, &taxi); err != nil {
		return err
	}

	p.workloadConfigPath = taxi.WorkloadConfigPath

	return nil
}

func (p *initSubcommand) PreScaffold(machinery.Filesystem) error {

	// load the workload config
	workload, err := workloadv1.ProcessInitConfig(
		p.workloadConfigPath,
	)
	if err != nil {
		return err
	}

	// validate the workload config
	if err := workload.Validate(); err != nil {
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
