// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

const (
	defaultDescription = `Manage %s workload`
)

// CliCommand defines the command name and description for the root command or
// subcommand of a companion CLI.
type CliCommand struct {
	Name          string
	Description   string
	VarName       string `json:"-" yaml:"-" validate:"omitempty"`
	FileName      string `json:"-" yaml:"-" validate:"omitempty"`
	IsSubcommand  bool   `json:"-" yaml:"-" validate:"omitempty"`
	IsRootcommand bool   `json:"-" yaml:"-" validate:"omitempty"`
}

func (cli *CliCommand) SetDefaults(workload WorkloadAPIBuilder, isSubcommand bool) {
	cli.IsSubcommand = isSubcommand
	cli.IsRootcommand = !isSubcommand

	if !cli.hasName() {
		cli.Name = cli.getDefaultName(workload)
	}

	if !cli.hasDescription() {
		cli.Description = cli.getDefaultDescription(workload)
	}
}

func (cli *CliCommand) getDefaultName(workload WorkloadAPIBuilder) string {
	if workload.IsCollection() && cli.IsSubcommand {
		return defaultCollectionSubcommandName
	}

	return strings.ToLower(workload.GetAPIKind())
}

func (cli *CliCommand) getDefaultDescription(workload WorkloadAPIBuilder) string {
	kind := strings.ToLower(workload.GetAPIKind())

	if workload.IsCollection() {
		if cli.IsSubcommand {
			return fmt.Sprintf(defaultCollectionSubcommandDescription, kind)
		}

		return fmt.Sprintf(defaultCollectionRootcommandDescription, kind)
	}

	return fmt.Sprintf(defaultDescription, kind)
}

func (cli *CliCommand) setCommonValues(workload WorkloadAPIBuilder, isSubcommand bool) {
	// ensure that defaults are properly set
	cli.SetDefaults(workload, isSubcommand)

	// set the file name and variable name to be used in the generated cli
	// codebase
	cli.FileName = utils.ToFileName(cli.Name)
	cli.VarName = utils.ToPascalCase(cli.Name)
}

func (cli *CliCommand) hasName() bool {
	return cli.Name != ""
}

func (cli *CliCommand) hasDescription() bool {
	return cli.Description != ""
}

// GetSubCmdRelativeFileName will generate a path for a subcommand CLI file
// that is relative to the root of the repository.
func (cli *CliCommand) GetSubCmdRelativeFileName(
	rootCmdName string,
	subCommandFolder string,
	group string,
	fileName string,
) string {
	return filepath.Join("cmd", rootCmdName, "commands", subCommandFolder, group, fileName+".go")
}
