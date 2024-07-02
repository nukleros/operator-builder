// Copyright 2024 Nukleros
// SPDX-License-Identifier: MIT

package license

import "github.com/spf13/pflag"

// AddFlags adds a consistent set of license flags across plugin versions and commands.
func AddFlags(fs *pflag.FlagSet, projectLicensePath, sourceHeaderPath *string) {
	fs.StringVar(projectLicensePath, "project-license", "", "path to project license file")
	fs.StringVar(sourceHeaderPath, "source-header-license", "", "path to file with source code header license text")
}
