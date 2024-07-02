// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v2

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v4/pkg/config"
	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v4/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v4/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v4/pkg/plugin/util"
	goplugin "sigs.k8s.io/kubebuilder/v4/pkg/plugins/golang"

	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds"
	"github.com/nukleros/operator-builder/internal/workload/v1/commands/subcommand"
	workloadconfig "github.com/nukleros/operator-builder/internal/workload/v1/config"
	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

// DefaultMainPath is default file path of main.go
const DefaultMainPath = "main.go"

type createAPISubcommand struct {
	config   config.Config
	options  *goplugin.Options
	resource *resource.Resource

	// resourceFlag   *pflag.Flag
	// controllerFlag *pflag.Flag
	force   bool
	runMake bool

	workloadConfigPath string
	cliRootCommandName string
	workload           kinds.WorkloadBuilder
	enableOlm          bool
}

var (
	ErrScaffoldCreateAPI = errors.New("unable to scaffold api")
	ErrAPIResourceExists = errors.New("API resource already exists")
	ErrMissingRootFile   = errors.New("file should present in the root directory")
)

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
	fs.BoolVar(&p.enableOlm, "enable-olm", false, "enable support for OpenShift Lifecycle Manager")

	fs.BoolVar(&p.runMake, "make", true, "if true, run `make generate` after generating files")

	fs.BoolVar(&p.force, "force", false,
		"attempt to create resource even if it already exists")

	// always scaffold a controller and an api
	p.options = &goplugin.Options{
		DoAPI:        true,
		DoController: true,
	}
	// fs.BoolVar(&p.options.DoAPI, "resource", true,
	// 	"if set, generate the resource without prompting the user")
	// fs.BoolVar(&p.options.DoController, "controller", true,
	// 	"if set, generate the controller without prompting the user")
	// p.resourceFlag = fs.Lookup("resource")

	fs.StringVar(&p.options.Plural, "plural", "", "resource irregular plural form")
	fs.BoolVar(&p.options.Namespaced, "namespaced", true, "resource is namespaced")

	// p.controllerFlag = fs.Lookup("controller")
}

func (p *createAPISubcommand) InjectConfig(c config.Config) error {
	processor, err := workloadconfig.Parse(p.workloadConfigPath)
	if err != nil {
		return fmt.Errorf("unable to inject config into %s, %w", p.workloadConfigPath, err)
	}

	p.workload = processor.Workload
	p.cliRootCommandName = p.workload.GetRootCommand().Name

	pluginConfig := workloadconfig.Plugin{
		WorkloadConfigPath: p.workloadConfigPath,
		CliRootCommandName: p.cliRootCommandName,
		EnableOLM:          p.enableOlm,
	}

	if err := c.EncodePluginConfig(workloadconfig.PluginKey, pluginConfig); err != nil {
		return fmt.Errorf("unable to encode plugin config at key %s, %w", workloadconfig.PluginKey, err)
	}

	p.config = c

	return nil
}

func (p *createAPISubcommand) InjectResource(res *resource.Resource) error {
	// set from config file if not provided with command line flag
	if res.Group == "" {
		res.Group = p.workload.GetAPIGroup()
	}

	if res.Version == "" {
		res.Version = p.workload.GetAPIVersion()
	}

	if res.Kind == "" {
		res.Kind = p.workload.GetAPIKind()
		res.Plural = resource.RegularPlural(p.workload.GetAPIKind())
	}

	// TODO: re-evaluate whether y/n input still makes sense. We should probably always
	//       scaffold the resource and controller.
	// Ask for API and Controller if not specified
	// reader := bufio.NewReader(os.Stdin)
	// if !p.resourceFlag.Changed {
	// 	log.Println("Create Resource [y/n]")
	// 	p.options.DoAPI = util.YesNo(reader)
	// }
	// if !p.controllerFlag.Changed {
	// 	log.Println("Create Controller [y/n]")
	// 	p.options.DoController = util.YesNo(reader)
	// }

	p.options.UpdateResource(res, p.config)
	res.Path = path.Join(p.config.GetRepository(), "apis", res.Group, res.Version)

	if err := res.Validate(); err != nil {
		return err
	}

	// In case we want to scaffold a resource API we need to do some checks
	if p.options.DoAPI {
		// Check that resource doesn't have the API scaffolded or flag force was set
		if r, err := p.config.GetResource(res.GVK); err == nil && r.HasAPI() && !p.force {
			return ErrAPIResourceExists
		}
	}

	if !p.config.HasResource(res.GVK) {
		if err := p.config.AddResource(*res); err != nil {
			return fmt.Errorf("unable to add resource to config, %w", err)
		}
	}

	p.resource = res

	return nil
}

func (p *createAPISubcommand) PreScaffold(machinery.Filesystem) error {
	processor, err := workloadconfig.Parse(p.workloadConfigPath)
	if err != nil {
		return fmt.Errorf("%s for %s, %w", ErrScaffoldCreateAPI.Error(), p.workloadConfigPath, err)
	}

	if err := subcommand.CreateAPI(processor); err != nil {
		return fmt.Errorf("%s for %s, %w", ErrScaffoldCreateAPI.Error(), p.workloadConfigPath, err)
	}

	// check if main.go is present in the root directory
	if _, err := os.Stat(utils.DefaultMainPath); os.IsNotExist(err) {
		return fmt.Errorf("missing file [%s], %w", utils.DefaultMainPath, ErrMissingRootFile)
	}

	p.workload = processor.Workload

	return nil
}

func (p *createAPISubcommand) Scaffold(fs machinery.Filesystem) error {
	scaffolder := scaffolds.NewAPIScaffolder(
		p.config,
		p.resource,
		p.workload,
		p.cliRootCommandName,
		p.enableOlm,
	)
	scaffolder.InjectFS(fs)

	if err := scaffolder.Scaffold(); err != nil {
		return fmt.Errorf("%s for %s, %w", ErrScaffoldInit.Error(), p.workloadConfigPath, err)
	}

	return nil
}

func (p *createAPISubcommand) PostScaffold() error {
	err := util.RunCmd("Update dependencies", "go", "mod", "tidy")
	if err != nil {
		return err
	}

	if p.runMake && p.resource.HasAPI() {
		err = util.RunCmd("Running make", "make", "generate")
		if err != nil {
			return err
		}
		fmt.Print("Next: implement your new API and generate the manifests (e.g. CRDs,CRs) with:\n$ make manifests\n")
	}

	return nil
}
