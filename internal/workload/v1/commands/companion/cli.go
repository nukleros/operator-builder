// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package companion

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nukleros/operator-builder/internal/utils"
)

const (
	defaultDescription                      = `Manage %s workload`
	defaultCollectionSubcommandName         = `collection`
	defaultCollectionSubcommandDescription  = `Manage %s workload`
	defaultCollectionRootcommandDescription = `Manage %s collection and components`
)

// companionCLIProcessor is an interface which includes the methods for processing
// a workload for the purpose of generating a companion CLI.
type companionCLIProcessor interface {
	IsCollection() bool
	GetAPIKind() string
}

// CLI defines the command name and description for the root command or
// subcommand of a companion CLI.
type CLI struct {
	Name          string
	Description   string
	VarName       string `json:"-" yaml:"-" validate:"omitempty"`
	FileName      string `json:"-" yaml:"-" validate:"omitempty"`
	IsSubcommand  bool   `json:"-" yaml:"-" validate:"omitempty"`
	IsRootcommand bool   `json:"-" yaml:"-" validate:"omitempty"`
}

// SetDefaults sets the default values for a companion CLI.
func (cli *CLI) SetDefaults(workload companionCLIProcessor, isSubcommand bool) {
	cli.IsSubcommand = isSubcommand
	cli.IsRootcommand = !isSubcommand

	if !cli.HasName() {
		cli.Name = cli.getDefaultName(workload)
	}

	if !cli.HasDescription() {
		cli.Description = cli.getDefaultDescription(workload)
	}
}

// SetCommonValues sets the common values for a companion CLI.
func (cli *CLI) SetCommonValues(workload companionCLIProcessor, isSubcommand bool) {
	// ensure that defaults are properly set
	cli.SetDefaults(workload, isSubcommand)

	// set the file name and variable name to be used in the generated cli
	// codebase
	cli.FileName = utils.ToFileName(cli.Name)
	cli.VarName = utils.ToPascalCase(cli.Name)
}

// HasName is a helper method which determines if a companion CLI has a name set.
func (cli *CLI) HasName() bool {
	return cli.Name != ""
}

// HasDescription is a helper method which determines if a companion CLI has a description set.
func (cli *CLI) HasDescription() bool {
	return cli.Description != ""
}

// GetSubCmdRelativeFileName will generate a path for a subcommand CLI file
// that is relative to the root of the repository.
func (cli *CLI) GetSubCmdRelativeFileName(
	rootCmdName string,
	subCommandFolder string,
	group string,
	fileName string,
) string {
	return filepath.Join("cmd", rootCmdName, "commands", subCommandFolder, group, fileName+".go")
}

// getDefaultName determines the default command name for a companion CLI subcommand.
func (cli *CLI) getDefaultName(workload companionCLIProcessor) string {
	if workload.IsCollection() && cli.IsSubcommand {
		return defaultCollectionSubcommandName
	}

	return strings.ToLower(workload.GetAPIKind())
}

// getDefaultDescription determines the default command description for a companion CLI subcommand.
func (cli *CLI) getDefaultDescription(workload companionCLIProcessor) string {
	kind := strings.ToLower(workload.GetAPIKind())

	if workload.IsCollection() {
		if cli.IsSubcommand {
			return fmt.Sprintf(defaultCollectionSubcommandDescription, kind)
		}

		return fmt.Sprintf(defaultCollectionRootcommandDescription, kind)
	}

	return fmt.Sprintf(defaultDescription, kind)
}
