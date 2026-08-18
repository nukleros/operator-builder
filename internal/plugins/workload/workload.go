// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package workload

import (
	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v4/pkg/model/resource"

	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

// AddFlags adds a consistent set of workload flags across plugin versions and commands.
func AddFlags(fs *pflag.FlagSet, workloadConfigPath *string, enableOlm *bool) {
	fs.StringVar(workloadConfigPath, "workload-config", "", "path to workload config file")

	if enableOlm != nil {
		fs.BoolVar(enableOlm, "enable-olm", false, "enable support for OpenShift Lifecycle Manager")
	}
}

// InjectResourceGVK injects the resource group version and kind.  It adds them from a workload
// if they are explicitly missing from the resource.
func InjectResourceGVK(res *resource.Resource, workload kinds.WorkloadBuilder) {
	if res.Group == "" {
		res.Group = workload.GetAPIGroup()
	}

	if res.Version == "" {
		res.Version = workload.GetAPIVersion()
	}

	if res.Kind == "" {
		res.Kind = workload.GetAPIKind()
		res.Plural = resource.RegularPlural(workload.GetAPIKind())
	}
}
