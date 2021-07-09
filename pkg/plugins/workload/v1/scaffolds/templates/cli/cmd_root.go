package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &CliCmdRoot{}

// CliCmdRoot scaffolds the root command file for the companion CLI.
type CliCmdRoot struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin

	// CliRootCmd is the root command for the companion CLI
	CliRootCmd string
	// CliRootDescription is the command description given by the CLI help info
	CliRootDescription string
}

func (f *CliCmdRoot) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.CliRootCmd, "commands", "root.go")

	f.TemplateBody = cliCmdRootTemplate

	return nil
}

const cliCmdRootTemplate = `{{ .Boilerplate }}

package commands

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "{{ .CliRootCmd }}",
	Short: "{{ .CliRootDescription }}",
	Long:  "{{ .CliRootDescription }}",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.{{ .CliRootCmd }}.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".{{ .CliRootCmd }}" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".{{ .CliRootCmd }}")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
`
