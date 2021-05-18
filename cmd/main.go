package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kubebuilder/v3/pkg/cli"
	cfgv2 "sigs.k8s.io/kubebuilder/v3/pkg/config/v2"
	cfgv3 "sigs.k8s.io/kubebuilder/v3/pkg/config/v3"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	kustomizecommonv1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/kustomize/v1"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang"
	declarativev1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/declarative/v1"
	golangv2 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v2"
	golangv3 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3"

	kblcli "gitlab.eng.vmware.com/landerr/kb-license-plugin/pkg/cli"
	licensev1 "gitlab.eng.vmware.com/landerr/kb-license-plugin/pkg/plugins/license/v1"
)

var (
	commands = []*cobra.Command{
		kblcli.NewUpdateCmd(),
	}
)

func main() {

	gov3Bundle, _ := plugin.NewBundle(golang.DefaultNameQualifier, plugin.Version{Number: 3},
		licensev1.Plugin{},
		kustomizecommonv1.Plugin{},
		golangv3.Plugin{},
	)

	c, err := cli.New(
		cli.WithCommandName("kbl"),
		cli.WithVersion(versionString()),
		cli.WithPlugins(
			golangv2.Plugin{},
			gov3Bundle,
			&licensev1.Plugin{},
			&kustomizecommonv1.Plugin{},
			&declarativev1.Plugin{},
		),
		cli.WithDefaultPlugins(cfgv2.Version, golangv2.Plugin{}),
		cli.WithDefaultPlugins(cfgv3.Version, gov3Bundle),
		cli.WithDefaultProjectVersion(cfgv3.Version),
		cli.WithExtraCommands(commands...),
		cli.WithCompletion(),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
}

func versionString() string {
	return "v1"
}
