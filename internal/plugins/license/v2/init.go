// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v2

import (
	"fmt"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v4/pkg/config"
	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v4/pkg/plugin"

	"github.com/nukleros/operator-builder/internal/license"
	licenseplugin "github.com/nukleros/operator-builder/internal/plugins/license"
)

var _ plugin.InitSubcommand = &initSubcommand{}

type initSubcommand struct {
	config config.Config

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
	licenseplugin.AddFlags(fs, &p.projectLicensePath, &p.sourceHeaderPath)
}

func (p *initSubcommand) InjectConfig(c config.Config) {
	_ = c.SetPluginChain([]string{pluginKey})
	p.config = c
}

func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {
	// project license
	if p.projectLicensePath != "" {
		if err := license.UpdateProjectLicense(p.projectLicensePath); err != nil {
			return fmt.Errorf("unable to update project license at %s, %w", p.projectLicensePath, err)
		}
	}

	// source header license
	if p.sourceHeaderPath != "" {
		if err := license.UpdateSourceHeader(p.sourceHeaderPath); err != nil {
			return fmt.Errorf("unable to update source header file at %s, %w", p.sourceHeaderPath, err)
		}
	}

	return nil
}
