// Copyright 2006-2021 VMware, Inc.
// SPDX-License-Identifier: MIT
package cli

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kubebuilder/v3/pkg/cli"
	cfgv3 "sigs.k8s.io/kubebuilder/v3/pkg/config/v3"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang"
	declarativev1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/declarative/v1"
	golangv3 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3"

	"gitlab.eng.vmware.com/landerr/kb-license-plugin/pkg/license"
)

var (
	commands = []*cobra.Command{
		license.LicenseCmd,
	}
)

// GetPluginsCLI returns the plugins based CLI configured to be used in your CLI
// binary
func GetPluginsCLI() *cli.CLI {
	gov3Bundle, _ := plugin.NewBundle(golang.DefaultNameQualifier, plugin.Version{Number: 3},
		golangv3.Plugin{},
	)
	c, err := cli.New(
		cli.WithCommandName("kbl"),
		cli.WithVersion(versionString()),
		cli.WithPlugins(
			gov3Bundle,
			&declarativev1.Plugin{},
		),
		cli.WithDefaultPlugins(cfgv3.Version, gov3Bundle),
		cli.WithDefaultProjectVersion(cfgv3.Version),
		cli.WithExtraCommands(commands...),
		cli.WithCompletion(),
	)
	if err != nil {
		log.Fatal(err)
	}

	return c
}

func versionString() string {
	return "v1"
}
