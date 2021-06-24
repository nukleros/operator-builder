package v1

import (
	"fmt"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/license"
)

var _ plugin.InitSubcommand = &initSubcommand{}

type initSubcommand struct {
	config      config.Config
	commandName string

	// license files
	projectLicensePath string
	sourceHeaderPath   string
}

func (p *initSubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {
	subcmdMeta.Description = `Add a project license file and license headers to source code files
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Add a project license file from a sample on the local filesystem
  %[1]s init --project-license /path/to/sample/LICENSE

  # Add the source file header boilerplate based on a sample on the local filesystem
  %[1]s init --source-header-license /path/to/sample/source-header.txt
`, cliMeta.CommandName)
}

func (p *initSubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&p.projectLicensePath, "project-license", "", "path to project license file")
	fs.StringVar(&p.sourceHeaderPath, "source-header-license", "", "path to file with source code header license text")
}

func (p *initSubcommand) InjectConfig(c config.Config) {
	_ = c.SetPluginChain([]string{pluginKey})
	p.config = c
}

func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {
	// project license
	if len(p.projectLicensePath) != 0 {
		if err := license.UpdateProjectLicense(p.projectLicensePath); err != nil {
			return err
		}
	}

	// source header license
	if len(p.sourceHeaderPath) != 0 {
		if err := license.UpdateSourceHeader(p.sourceHeaderPath); err != nil {
			return err
		}
	}

	return nil
}
