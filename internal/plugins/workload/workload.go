// Copyright 2024 Nukleros
// SPDX-License-Identifier: MIT

package workload

import "github.com/spf13/pflag"

// AddFlags adds a consistent set of workload flags across plugin versions and commands.
func AddFlags(fs *pflag.FlagSet, workloadConfigPath *string, enableOlm *bool) {
	fs.StringVar(workloadConfigPath, "workload-config", "", "path to workload config file")
	fs.BoolVar(enableOlm, "enable-olm", false, "enable support for OpenShift Lifecycle Manager")
}
