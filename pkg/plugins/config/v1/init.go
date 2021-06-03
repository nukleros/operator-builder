package v1

import (
	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	workloadv1 "gitlab.eng.vmware.com/landerr/operator-builder/pkg/workload/v1"
)

type initSubcommand struct {
	standaloneWorkloadConfigPath string
	workloadCollectionConfigPath string
}

var _ plugin.InitSubcommand = &initSubcommand{}

func (p *initSubcommand) BindFlags(fs *pflag.FlagSet) {

	fs.StringVar(&p.standaloneWorkloadConfigPath, "standalone-workload-config", "", "path to standalone workload config file")
	fs.StringVar(&p.workloadCollectionConfigPath, "workload-collection-config", "", "path to workload collection config file")
}

func (p *initSubcommand) InjectConfig(c config.Config) error {

	taxi := workloadv1.ConfigTaxi{
		StandaloneConfigPath: p.standaloneWorkloadConfigPath,
		CollectionConfigPath: p.workloadCollectionConfigPath,
	}

	if err := c.EncodePluginConfig(workloadv1.ConfigTaxiKey, taxi); err != nil {
		return err
	}

	workload, err := workloadv1.ProcessInitConfig(
		p.standaloneWorkloadConfigPath,
		p.workloadCollectionConfigPath,
	)
	if err != nil {
		return err
	}

	if err := c.SetDomain(workload.GetDomain()); err != nil {
		return err
	}

	return nil
}

func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {
	return nil
}
