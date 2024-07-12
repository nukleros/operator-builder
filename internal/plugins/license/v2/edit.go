// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

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

var _ plugin.EditSubcommand = &editSubcommand{}

type editSubcommand struct {
	config config.Config

	// license files
	projectLicensePath string
	sourceHeaderPath   string
}

func (p *editSubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {
	subcmdMeta.Description = `This command will edit the project configuration.
Features supported:
  - Update the project license and license header.
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Add a project license file from a sample on the local filesystem
%[1]s edit --project-license /path/to/sample/LICENSE

# Add the source file header boilerplate based on a sample on the local filesystem
%[1]s edit --source-header-license /path/to/sample/source-header.txt
`, cliMeta.CommandName)
}

func (p *editSubcommand) BindFlags(fs *pflag.FlagSet) {
	licenseplugin.AddFlags(fs, &p.projectLicensePath, &p.sourceHeaderPath)
}

func (p *editSubcommand) InjectConfig(c config.Config) error {
	p.config = c

	return nil
}

func (p *editSubcommand) Scaffold(fs machinery.Filesystem) error {
	// project license
	if p.projectLicensePath != "" {
		if err := license.UpdateProjectLicense(p.projectLicensePath); err != nil {
			return fmt.Errorf("unable to update project license at %s, %w", p.projectLicensePath, err)
		}
	}

	// source header license
	if p.sourceHeaderPath != "" {
		// boilerplate
		if err := license.UpdateSourceHeader(p.sourceHeaderPath); err != nil {
			return fmt.Errorf("unable to update source header file at %s, %w", p.sourceHeaderPath, err)
		}

		// existing source code files
		if err := license.UpdateExistingSourceHeader(p.sourceHeaderPath); err != nil {
			return fmt.Errorf("unable to update source header file at %s, %w", p.sourceHeaderPath, err)
		}
	}

	return nil
}
