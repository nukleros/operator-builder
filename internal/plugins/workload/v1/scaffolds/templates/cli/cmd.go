// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &CmdCommon{}

// CmdCommon scaffolds the companion CLI's common code for the
// workload.  This where the actual generate logic lives.
type CmdCommon struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	RootCmd string
}

func (f *CmdCommon) SetTemplateDefaults() error {
	f.Path = filepath.Join("cmd", f.RootCmd, "commands", "commands.go")

	f.TemplateBody = cliCmdCommonTemplate

	return nil
}

//nolint: lll
const cliCmdCommonTemplate = `{{ .Boilerplate }}

package commands

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"{{ .Repo }}/apis/common"
)

// validateWorkload validates the unmarshaled version of the workload resource
// manifest.
func validateWorkload(
	workload common.Component,
) error {
	defaultWorkloadGVK := workload.GetComponentGVK()
	component := workload.(runtime.Object)

	if defaultWorkloadGVK != component.GetObjectKind().GroupVersionKind() {
		return fmt.Errorf(
			"expected resource of kind: '%s', with group '%s' and version '%s'; "+
				"found resource of kind '%s', with group '%s' and version '%s'",
			defaultWorkloadGVK.Kind,
			defaultWorkloadGVK.Group,
			defaultWorkloadGVK.Version,
			component.GetObjectKind().GroupVersionKind().Kind,
			component.GetObjectKind().GroupVersionKind().Group,
			component.GetObjectKind().GroupVersionKind().Version,
		)
	}

	return nil
}
`
