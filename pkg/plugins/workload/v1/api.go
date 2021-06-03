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

	standaloneWorkloadConfigPath string
	componentWorkloadConfigPath  string
	workloadConfigPath           string

	workload workloadv1.WorkloadAPIBuilder
	project  workloadv1.Project
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

	fs.StringVar(&p.standaloneWorkloadConfigPath, "standalone-workload-config", "", "path to standalone workload config file")
	fs.StringVar(&p.componentWorkloadConfigPath, "component-workload-config", "", "path to component workload config file")
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

	// process workload config file
	workload, pathInUse, err := workloadv1.ProcessAPIConfig(
		p.standaloneWorkloadConfigPath,
		p.componentWorkloadConfigPath,
	)
	if err != nil {
		return err
	}
	p.workload = workload
	p.workloadConfigPath = pathInUse

	// get project config file
	projectFile, err := ioutil.ReadFile("WORKLOAD")
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(projectFile, &p.project)
	if err != nil {
		return err
	}

	return nil
}

func (p *createAPISubcommand) Scaffold(fs machinery.Filesystem) error {

	// The specFields contain all fields to build into the API type spec
	specFields, err := p.workload.GetSpecFields(p.workloadConfigPath)

	// The sourceFiles contain the information needed to build resource source
	// code files
	sourceFiles, err := p.workload.GetResources(p.workloadConfigPath)

	scaffolder := scaffolds.NewAPIScaffolder(
		p.config,
		*p.resource,
		p.workload,
		p.workloadConfigPath,
		specFields,
		sourceFiles,
		&p.project,
	)
	scaffolder.InjectFS(fs)
	err = scaffolder.Scaffold()
	if err != nil {
		return err
	}

	return nil
}
