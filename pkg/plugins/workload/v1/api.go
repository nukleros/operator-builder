package v1

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/yaml"

	"gitlab.eng.vmware.com/landerr/operator-builder/pkg/plugins/workload/v1/scaffolds"
	workloadv1 "gitlab.eng.vmware.com/landerr/operator-builder/pkg/workload/v1"
)

type createAPISubcommand struct {
	config config.Config

	resource *resource.Resource

	workloadConfigPath string
	workloadConfig     workloadv1.WorkloadConfig
}

var _ plugin.CreateAPISubcommand = &createAPISubcommand{}

func (p *createAPISubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {

	subcmdMeta.Description = `Build a new API that can capture state for workloads
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Add API attributes defined by a workload config file
  %[1]s create api --workload-config .source-manifests/workload.yaml
`, cliMeta.CommandName)
}

func (p *createAPISubcommand) BindFlags(fs *pflag.FlagSet) {

	fs.StringVar(&p.workloadConfigPath, "workload-config", "", "path to workload config file")
}

func (p *createAPISubcommand) InjectConfig(c config.Config) error {

	p.config = c

	return nil
}

func (p *createAPISubcommand) InjectResource(res *resource.Resource) error {

	p.resource = res

	return nil
}

func (p *createAPISubcommand) PreScaffold(machinery.Filesystem) error {

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

func (p *createAPISubcommand) Scaffold(fs machinery.Filesystem) error {

	// The specFields contain all fields to build into the API type spec
	specFields, err := p.workloadConfig.GetSpecFields(p.workloadConfigPath)

	// The sourceFiles contain the information needed to build resource source
	// code files
	sourceFiles, err := p.workloadConfig.GetResources(p.workloadConfigPath)

	scaffolder := scaffolds.NewAPIScaffolder(
		p.config,
		*p.resource,
		&p.workloadConfig,
		specFields,
		sourceFiles,
	)
	scaffolder.InjectFS(fs)
	err = scaffolder.Scaffold()
	if err != nil {
		return err
	}

	return nil
}
