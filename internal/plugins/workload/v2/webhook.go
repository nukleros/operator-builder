// Copyright 2026 Nukleros
// SPDX-License-Identifier: Apache-2.0

/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v2

import (
	"errors"
	"fmt"
	log "log/slog"
	"path"
	"strings"

	"github.com/spf13/pflag"

	"sigs.k8s.io/kubebuilder/v4/pkg/config"
	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v4/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v4/pkg/plugin"
	pluginutil "sigs.k8s.io/kubebuilder/v4/pkg/plugin/util"
	goPlugin "sigs.k8s.io/kubebuilder/v4/pkg/plugins/golang"

	"github.com/nukleros/operator-builder/internal/plugins/workload"
	"github.com/nukleros/operator-builder/internal/plugins/workload/v2/scaffolds"
	workloadconfig "github.com/nukleros/operator-builder/internal/workload/v1/config"
	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

var _ plugin.CreateWebhookSubcommand = &createWebhookSubcommand{}

type createWebhookSubcommand struct {
	config config.Config
	// For help text.
	commandName string

	options *goPlugin.Options

	resource *resource.Resource

	// force indicates that the resource should be created even if it already exists
	force bool

	// runMake indicates whether to run make or not after scaffolding APIs
	runMake bool

	workloadConfigPath string
	workload           kinds.WorkloadBuilder
}

func (p *createWebhookSubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {
	p.commandName = cliMeta.CommandName

	subcmdMeta.Description = `Scaffold a webhook for an API resource. You can choose to scaffold defaulting,
validating and/or conversion webhooks.
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Create defaulting and validating webhooks for Group: ship, Version: v1beta1
  # and Kind: Frigate
  %[1]s create webhook --group ship --version v1beta1 --kind Frigate --defaulting --programmatic-validation

  # Create conversion webhook for Group: ship, Version: v1beta1
  # and Kind: Frigate
  %[1]s create webhook --group ship --version v1beta1 --kind Frigate --conversion --spoke v1

  # Create defaulting webhook with custom path for Group: ship, Version: v1beta1
  # and Kind: Frigate
  %[1]s create webhook --group ship --version v1beta1 --kind Frigate --defaulting \
    --defaulting-path=/my-custom-mutate-path

  # Create validation webhook with custom path for Group: ship, Version: v1beta1
  # and Kind: Frigate
  %[1]s create webhook --group ship --version v1beta1 --kind Frigate \
    --programmatic-validation --validation-path=/my-custom-validate-path

  # Create both defaulting and validation webhooks with different custom paths
  %[1]s create webhook --group ship --version v1beta1 --kind Frigate \
    --defaulting --programmatic-validation \
    --defaulting-path=/custom-mutate --validation-path=/custom-validate
`, cliMeta.CommandName)
}

func (p *createWebhookSubcommand) BindFlags(fs *pflag.FlagSet) {
	workload.AddFlags(fs, &p.workloadConfigPath, nil)

	p.options = &goPlugin.Options{}

	fs.BoolVar(&p.runMake, "make", true,
		"Run 'make generate' after generating files (enabled by default; use --make=false to disable)")

	fs.StringVar(&p.options.Plural, "plural", "",
		"Resource irregular plural form (e.g., 'people' for 'Person'); auto-detected from resource kind if not provided")

	fs.BoolVar(&p.options.DoDefaulting, "defaulting", false,
		"If set, scaffold the defaulting webhook")
	fs.BoolVar(&p.options.DoValidation, "programmatic-validation", false,
		"If set, scaffold the validating webhook")
	fs.BoolVar(&p.options.DoConversion, "conversion", false,
		"If set, scaffold the conversion webhook")

	fs.StringSliceVar(&p.options.Spoke, "spoke",
		nil,
		"Comma-separated list of spoke versions to be added to the conversion webhook (e.g., --spoke v1,v2)")

	fs.StringVar(&p.options.DefaultingPath, "defaulting-path", "",
		"[Optional] Custom path for the defaulting/mutating webhook (e.g., /my-custom-mutate-path). "+
			"Only valid with --defaulting")

	fs.StringVar(&p.options.ValidationPath, "validation-path", "",
		"[Optional] Custom path for the validation webhook (e.g., /my-custom-validate-path). "+
			"Only valid with --programmatic-validation")

	fs.StringVar(&p.options.ExternalAPIPath, "external-api-path", "",
		"Go package import path for the external API (e.g., github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1). "+
			"Used to scaffold webhooks for resources defined outside this project")

	fs.StringVar(&p.options.ExternalAPIDomain, "external-api-domain", "",
		"Domain name for the external API (e.g., cert-manager.io). "+
			"Used to generate accurate RBAC markers and permissions for the external resources")

	fs.StringVar(&p.options.ExternalAPIModule, "external-api-module", "",
		"External API module with optional version (e.g., github.com/cert-manager/cert-manager@v1.18.2)")

	fs.BoolVar(&p.force, "force", false,
		"If set, attempt to create resource even if it already exists")
}

func (p *createWebhookSubcommand) InjectConfig(c config.Config) error {
	processor, err := workloadconfig.Parse(p.workloadConfigPath)
	if err != nil {
		return fmt.Errorf("unable to inject config into %s, %w", p.workloadConfigPath, err)
	}

	p.workload = processor.Workload

	pluginConfig := workloadconfig.Plugin{WorkloadConfigPath: p.workloadConfigPath}

	if err := c.EncodePluginConfig(workloadconfig.PluginKey, pluginConfig); err != nil {
		return fmt.Errorf("unable to encode plugin config at key %s, %w", workloadconfig.PluginKey, err)
	}

	p.config = c

	return nil
}

func (p *createWebhookSubcommand) InjectResource(res *resource.Resource) error {
	// set from config file if not provided with command line flag
	workload.InjectResourceGVK(res, p.workload)

	p.resource = res

	if err := p.validateFlagCombinations(); err != nil {
		return err
	}

	// Normalize and validate the spokes before UpdateResource copies p.options.Spoke
	// onto the resource wholesale; appending to res.Webhooks.Spoke afterwards would
	// only duplicate every entry.
	if err := p.validateSpokes(res); err != nil {
		return err
	}

	p.options.UpdateResource(res, p.config)

	if err := p.updateResourceFromConfig(res); err != nil {
		return err
	}

	// goPlugin.Options.UpdateResource hardcodes api/ via resource.APIPackagePath, so
	// operator-builder's apis/ layout has to be applied after the last call to it.
	// This resource pointer is shared with every other plugin in the bundle, all of
	// which reconcile it against the path already recorded in PROJECT. Leave external
	// and core types alone; UpdateResource sets their paths to a foreign package.
	if !res.External && !res.Core {
		res.Path = path.Join(p.config.GetRepository(), "apis", res.Group, res.Version)
	}

	if err := p.resource.Validate(); err != nil {
		return fmt.Errorf("error validating resource: %w", err)
	}

	if err := p.validateWebhookTypes(); err != nil {
		return err
	}

	return p.reconcileExistingResource()
}

func (p *createWebhookSubcommand) Scaffold(fs machinery.Filesystem) error {
	scaffolder := scaffolds.NewWebhookScaffolder(p.config, p.resource, p.force)
	scaffolder.InjectFS(fs)
	if err := scaffolder.Scaffold(); err != nil {
		return fmt.Errorf("failed to scaffold webhook: %w", err)
	}

	return nil
}

func (p *createWebhookSubcommand) PostScaffold() error {
	// If external API with module specified, add it using go get
	if p.resource.IsExternal() && p.resource.Module != "" {
		log.Info("Adding external API dependency", "module", p.resource.Module)
		// Use go get to add the dependency cleanly as a direct requirement
		err := pluginutil.RunCmd("Add external API dependency", "go", "get", p.resource.Module)
		if err != nil {
			return fmt.Errorf("error adding external API dependency: %w", err)
		}
	}

	err := pluginutil.RunCmd("Update dependencies", "go", "mod", "tidy")
	if err != nil {
		return fmt.Errorf("error updating go dependencies: %w", err)
	}

	if p.runMake {
		err = pluginutil.RunCmd("Running make", "make", "generate")
		if err != nil {
			return fmt.Errorf("error running make generate: %w", err)
		}
	}

	log.Info("Next: implement your new Webhook and generate the manifests with: $ make manifests")

	return nil
}

func (p *createWebhookSubcommand) validateFlagCombinations() error {
	if p.options.DefaultingPath != "" && !p.options.DoDefaulting {
		return fmt.Errorf("--defaulting-path can only be used with --defaulting")
	}

	if p.options.ValidationPath != "" && !p.options.DoValidation {
		return fmt.Errorf("--validation-path can only be used with --programmatic-validation")
	}

	if p.options.ExternalAPIModule != "" && p.options.ExternalAPIPath == "" {
		return errors.New("'--external-api-module' requires '--external-api-path' to be specified")
	}

	return nil
}

func (p *createWebhookSubcommand) validateSpokes(res *resource.Resource) error {
	for i, spoke := range p.options.Spoke {
		spoke = strings.TrimSpace(spoke)
		if !isValidVersion(spoke, res, p.config) {
			return fmt.Errorf("invalid spoke version %q", spoke)
		}

		p.options.Spoke[i] = spoke
	}

	return nil
}

func (p *createWebhookSubcommand) validateWebhookTypes() error {
	if !p.resource.HasDefaultingWebhook() && !p.resource.HasValidationWebhook() && !p.resource.HasConversionWebhook() {
		return fmt.Errorf("%s create webhook requires at least one of --defaulting,"+
			" --programmatic-validation and --conversion to be true", p.commandName)
	}

	return nil
}

func (p *createWebhookSubcommand) reconcileExistingResource() error {
	existing, err := p.config.GetResource(p.resource.GVK)
	if err != nil && !p.resource.External && !p.resource.Core {
		return fmt.Errorf(
			"no API found for %s/%s, Kind %s: run 'create api' first, "+
				"or pass --external-api-path for an external type",
			p.resource.QualifiedGroup(),
			p.resource.Version,
			p.resource.Kind,
		)
	}

	if err == nil && existing.Webhooks != nil && !existing.Webhooks.IsEmpty() && !p.force {
		if err := p.checkWebhookConflicts(&existing); err != nil {
			return err
		}

		if err := p.resource.Webhooks.Update(existing.Webhooks); err != nil {
			return fmt.Errorf("error merging webhook configurations: %w", err)
		}
	}

	return nil
}

func (p *createWebhookSubcommand) checkWebhookConflicts(existing *resource.Resource) error {
	if p.resource.HasDefaultingWebhook() && existing.Webhooks.Defaulting {
		return fmt.Errorf("defaulting webhook already exists for this resource")
	}

	if p.resource.HasValidationWebhook() && existing.Webhooks.Validation {
		return fmt.Errorf("validation webhook already exists for this resource")
	}

	if p.resource.HasConversionWebhook() && existing.Webhooks.Conversion {
		return fmt.Errorf("conversion webhook already exists for this resource")
	}

	return nil
}

// updateResourceFromConfig copies existing resource configuration from PROJECT file.
func (p *createWebhookSubcommand) updateResourceFromConfig(res *resource.Resource) error {
	// Match by Group, Version, and Kind because external APIs may have
	// a different domain than the project domain.
	resources, err := p.config.GetResources()
	if err != nil {
		return fmt.Errorf("failed to load resources from project configuration: %w", err)
	}

	for i := range resources {
		existingRes := &resources[i]
		if existingRes.Group != res.Group ||
			existingRes.Version != res.Version ||
			existingRes.Kind != res.Kind {
			continue
		}

		p.resource.Domain = existingRes.Domain
		p.resource.Path = existingRes.Path
		p.resource.Plural = existingRes.Plural
		p.resource.External = existingRes.External
		p.resource.Core = existingRes.Core
		p.resource.Module = existingRes.Module

		break
	}

	return nil
}

// isValidVersion is a helper function to validate spoke versions.
func isValidVersion(version string, res *resource.Resource, cfg config.Config) bool {
	// Fetch all resources in the config
	resources, err := cfg.GetResources()
	if err != nil {
		return false
	}

	// Iterate through resources and validate if the given version exists for the same Group and Kind
	for i := range resources {
		r := &resources[i]
		if r.Group == res.Group && r.Kind == res.Kind && r.Version == version {
			return true
		}
	}

	// If no matching version is found, return false
	return false
}
