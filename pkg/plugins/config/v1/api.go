package v1

import (
	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	workloadv1 "gitlab.eng.vmware.com/landerr/operator-builder/pkg/workload/v1"
)

type createAPISubcommand struct {
	standaloneWorkloadConfigPath string
	componentWorkloadConfigPath  string
}

var _ plugin.CreateAPISubcommand = &createAPISubcommand{}

func (p *createAPISubcommand) BindFlags(fs *pflag.FlagSet) {

	fs.StringVar(&p.standaloneWorkloadConfigPath, "standalone-workload-config", "", "path to standalone workload config file")
	fs.StringVar(&p.componentWorkloadConfigPath, "component-workload-config", "", "path to component workload config file")
}

func (p *createAPISubcommand) InjectConfig(c config.Config) error {

	taxi := workloadv1.ConfigTaxi{
		StandaloneConfigPath: p.standaloneWorkloadConfigPath,
		ComponentConfigPath:  p.componentWorkloadConfigPath,
	}

	if err := c.EncodePluginConfig(workloadv1.ConfigTaxiKey, taxi); err != nil {
		return err
	}

	return nil
}

func (p *createAPISubcommand) InjectResource(res *resource.Resource) error {

	workload, _, err := workloadv1.ProcessAPIConfig(
		p.standaloneWorkloadConfigPath,
		p.componentWorkloadConfigPath,
	)
	if err != nil {
		return err
	}

	// set from config file if not provided with command line flag
	if res.Group == "" {
		res.Group = workload.GetGroup()
	}
	if res.Version == "" {
		res.Version = workload.GetVersion()
	}
	if res.Kind == "" {
		res.Kind = workload.GetKind()
		res.Plural = resource.RegularPlural(workload.GetKind())
	}

	return nil
}

func (p *createAPISubcommand) Scaffold(fs machinery.Filesystem) error {
	return nil
}
