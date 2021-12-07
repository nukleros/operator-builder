// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

// CliCommand defines the command name and description for the root command or
// subcommand of a companion CLI.
type CliCommand struct {
	Name        string
	Description string
	VarName     string
	FileName    string
	SubCommands *[]CliCommand
}

func (cli *CliCommand) setCommonValues(kind, descriptionTemplate string) {
	// set the file name and variable name to be used in the generated cli
	// codebase
	cli.FileName = utils.ToFileName(cli.Name)
	cli.VarName = utils.ToPascalCase(cli.Name)

	// provide a default description for the cli help menu if one has not been
	// provided, keying off the the api kind
	if !cli.hasDescription() {
		cli.Description = fmt.Sprintf(
			descriptionTemplate,
			strings.ToLower(kind),
		)
	}
}

func (cli *CliCommand) setSubCommandValues(kind, descriptionTemplate string) {
	// default the sub command name to the defaultSubcommandCollection if
	// it is not specified
	if !cli.hasName() {
		cli.Name = strings.ToLower(kind)
	}

	// set the common cli values
	cli.setCommonValues(kind, descriptionTemplate)
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
