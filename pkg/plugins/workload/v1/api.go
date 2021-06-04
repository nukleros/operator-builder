package v1

import (
	"fmt"
	"io/ioutil"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/plugins/workload/v1/scaffolds"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

type createAPISubcommand struct {
	config config.Config

	resource *resource.Resource

	workloadConfigPath string
	workload           workloadv1.WorkloadAPIBuilder
	project            workloadv1.Project
}

var _ plugin.CreateAPISubcommand = &createAPISubcommand{}

func (p *createAPISubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {

	subcmdMeta.Description = `Build a new API that can capture state for workloads
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Add API attributes defined by a workload config file
  %[1]s create api --workload-config .source-manifests/workload.yaml
`, cliMeta.CommandName)
}

func (p *createAPISubcommand) InjectConfig(c config.Config) error {

	p.config = c

	var taxi workloadv1.ConfigTaxi
	if err := c.DecodePluginConfig(workloadv1.ConfigTaxiKey, &taxi); err != nil {
		return err
	}

	p.workloadConfigPath = taxi.WorkloadConfigPath

	return nil
}

func (p *createAPISubcommand) InjectResource(res *resource.Resource) error {

	p.resource = res

	return nil
}

func (p *createAPISubcommand) PreScaffold(machinery.Filesystem) error {

	// load the workload config
	workload, err := workloadv1.ProcessAPIConfig(
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

	// get WORKLOAD project config file
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
