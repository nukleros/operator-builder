package v1

import (
	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

type createAPISubcommand struct {
	workloadConfigPath string
}

var _ plugin.CreateAPISubcommand = &createAPISubcommand{}

func (p *createAPISubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&p.workloadConfigPath, "workload-config", "", "path to workload config file")
}

func (p *createAPISubcommand) InjectConfig(c config.Config) error {
	taxi := workloadv1.ConfigTaxi{
		WorkloadConfigPath: p.workloadConfigPath,
	}

	if err := c.EncodePluginConfig(workloadv1.ConfigTaxiKey, taxi); err != nil {
		return err
	}

	return nil
}

func (p *createAPISubcommand) InjectResource(res *resource.Resource) error {
	workload, err := workloadv1.ProcessAPIConfig(
		p.workloadConfigPath,
	)
	if err != nil {
		return err
	}

	// set from config file if not provided with command line flag
	if res.Group == "" {
		res.Group = workload.GetAPIGroup()
	}

	if res.Version == "" {
		res.Version = workload.GetAPIVersion()
	}

	if res.Kind == "" {
		res.Kind = workload.GetAPIKind()
		res.Plural = resource.RegularPlural(workload.GetAPIKind())
	}

	return nil
}

func (p *createAPISubcommand) Scaffold(fs machinery.Filesystem) error {
	return nil
}
