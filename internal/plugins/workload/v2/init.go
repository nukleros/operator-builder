// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v2

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v4/pkg/config"
	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v4/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v4/pkg/plugins/golang"

	"github.com/nukleros/operator-builder/internal/plugins/workload"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds"
	"github.com/nukleros/operator-builder/internal/utils"
	"github.com/nukleros/operator-builder/internal/workload/v1/commands/subcommand"
	workloadconfig "github.com/nukleros/operator-builder/internal/workload/v1/config"
	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

// Variables and function to check Go version requirements.
//
//nolint:gochecknoglobals
var (
	goVerMin = golang.MustParse(fmt.Sprintf("go%s", utils.GeneratedGoVersionMinimum))
	goVerMax = golang.MustParse(fmt.Sprintf("go%s", utils.GeneratedGoVersionPreferred))
)

var (
	ErrDirectoryNotEmpty = errors.New("target directory is not empty")
)

type initSubcommand struct {
	config config.Config

	license            string
	owner              string
	repo               string
	fetchDeps          bool
	skipGoVersionCheck bool

	workloadConfigPath string
	cliRootCommandName string
	controllerImage    string
	enableOlm          bool

	workload kinds.WorkloadBuilder
}

var ErrScaffoldInit = errors.New("unable to scaffold initial config")

var _ plugin.InitSubcommand = &initSubcommand{}

func (p *initSubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {
	subcmdMeta.Description = `Add workload management scaffolding to a new project
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Add project scaffolding defined by a workload config file
  %[1]s init --workload-config .source-manifests/workload.yaml
`, cliMeta.CommandName)
}

func (p *initSubcommand) BindFlags(fs *pflag.FlagSet) {
	workload.AddFlags(fs, &p.workloadConfigPath, &p.enableOlm)
	fs.StringVar(&p.controllerImage, "controller-image", "controller:latest", "controller image")

	fs.BoolVar(&p.skipGoVersionCheck, "skip-go-version-check",
		false, "if specified, skip checking the Go version")

	// dependency args
	fs.BoolVar(&p.fetchDeps, "fetch-deps", true, "ensure dependencies are downloaded")

	// boilerplate args
	fs.StringVar(&p.license, "license", "apache2",
		"license to use to boilerplate, may be one of 'apache2', 'none'")
	fs.StringVar(&p.owner, "owner", "", "owner to add to the copyright")

	// project args
	fs.StringVar(&p.repo, "repo", "", "name to use for go module (e.g., github.com/user/repo), "+
		"defaults to the go package of the current working directory.")
}

func (p *initSubcommand) InjectConfig(c config.Config) error {
	processor, err := workloadconfig.Parse(p.workloadConfigPath)
	if err != nil {
		return fmt.Errorf("unable to inject config into %s, %w", p.workloadConfigPath, err)
	}

	p.config = c

	// operator builder always uses multi-group APIs
	if err := c.SetMultiGroup(); err != nil {
		return fmt.Errorf("unable to enable multigroup apis, %w", err)
	}

	pluginConfig := workloadconfig.Plugin{
		WorkloadConfigPath: p.workloadConfigPath,
		CliRootCommandName: processor.Workload.GetRootCommand().Name,
		ControllerImg:      p.controllerImage,
		EnableOLM:          p.enableOlm,
	}

	if err := c.EncodePluginConfig(workloadconfig.PluginKey, pluginConfig); err != nil {
		return fmt.Errorf("unable to encode operatorbuilder config key at %s, %w", p.workloadConfigPath, err)
	}

	if err := c.SetDomain(processor.Workload.GetDomain()); err != nil {
		return fmt.Errorf("unable to set project domain, %w", err)
	}

	// Try to guess repository if flag is not set.
	if p.repo == "" {
		repoPath, err := golang.FindCurrentRepo()
		if err != nil {
			return fmt.Errorf("error finding current repository, %w", err)
		}
		p.repo = repoPath
	}

	p.cliRootCommandName = pluginConfig.CliRootCommandName

	return p.config.SetRepository(p.repo)
}

func (p *initSubcommand) PreScaffold(machinery.Filesystem) error {
	processor, err := workloadconfig.Parse(p.workloadConfigPath)
	if err != nil {
		return fmt.Errorf("%s for %s, %w", ErrScaffoldInit.Error(), p.workloadConfigPath, err)
	}

	if err := subcommand.Init(processor); err != nil {
		return fmt.Errorf("%s for %s, %w", ErrScaffoldInit.Error(), p.workloadConfigPath, err)
	}

	p.workload = processor.Workload

	// Ensure Go version is in the allowed range if check not turned off.
	if !p.skipGoVersionCheck {
		if err := golang.ValidateGoVersion(goVerMin, goVerMax); err != nil {
			return err
		}
	}

	return checkDir()
}

func (p *initSubcommand) Scaffold(fs machinery.Filesystem) error {
	scaffolder := scaffolds.NewInitScaffolder(
		p.config,
		p.workload,
		p.cliRootCommandName,
		p.controllerImage,
		p.enableOlm,
		p.license,
		p.owner,
	)
	scaffolder.InjectFS(fs)

	if err := scaffolder.Scaffold(); err != nil {
		return fmt.Errorf("%s for %s, %w", ErrScaffoldInit.Error(), p.workloadConfigPath, err)
	}

	return nil
}

// checkDir will return error if the current directory has files which are not allowed.
// Note that, it is expected that the directory to scaffold the project is cleaned.
// Otherwise, it might face issues to do the scaffold.
func checkDir() error {
	err := filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Allow directory trees starting with '.'
			if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}

			// Allow files starting with '.'
			if strings.HasPrefix(info.Name(), ".") {
				return nil
			}

			// Allow files ending with '.md' extension
			if strings.HasSuffix(info.Name(), ".md") && !info.IsDir() {
				return nil
			}

			// Allow capitalized files except PROJECT
			isCapitalized := true
			for _, l := range info.Name() {
				if !unicode.IsUpper(l) {
					isCapitalized = false

					break
				}
			}

			if isCapitalized && info.Name() != "PROJECT" {
				return nil
			}

			// Allow files in the following list
			allowedFiles := []string{
				"go.mod", // user might run `go mod init` instead of providing the `--flag` at init
				"go.sum", // auto-generated file related to go.mod
			}

			for _, allowedFile := range allowedFiles {
				if info.Name() == allowedFile {
					return nil
				}
			}

			// Do not allow any other file
			return fmt.Errorf(
				"%w, (only %s, files and directories with the prefix \".\", "+
					"files with the suffix \".md\" or capitalized files name are allowed); "+
					"found existing file %q", ErrDirectoryNotEmpty, strings.Join(allowedFiles, ", "), path)
		})
	if err != nil {
		return err
	}

	return nil
}
