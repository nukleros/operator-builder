// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"
	"strings"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

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
