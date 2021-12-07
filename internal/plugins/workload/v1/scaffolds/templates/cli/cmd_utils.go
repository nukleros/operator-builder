// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

var _ machinery.Template = &CmdUtils{}

// CmdUtils scaffolds the companion CLI's common utility code for the
// workload.  This where the generic logic for a companion CLI lives.
type CmdUtils struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin

	Builder workloadv1.WorkloadAPIBuilder
}

func (f *CmdUtils) SetTemplateDefaults() error {
	// set interface variables
	f.Path = filepath.Join("cmd", f.Builder.GetRootCommand().Name, "commands", "utils", "utils.go")
	f.TemplateBody = cliCmdUtilsTemplate

	return nil
}

const cliCmdUtilsTemplate = `{{ .Boilerplate }}

package utils

import (
	"errors"
	"fmt"

	"{{ .Repo }}/apis/common"
)

var ErrInvalidResource = errors.New("supplied resource is incorrect")

// ValidateWorkload validates the unmarshaled version of the workload resource
// manifest.
func ValidateWorkload(workload common.Component) error {
	defaultWorkloadGVK := workload.GetComponentGVK()

	if defaultWorkloadGVK != workload.GetObjectKind().GroupVersionKind() {
		return fmt.Errorf(
			"%w, expected resource of kind: '%s', with group '%s' and version '%s'; "+
				"found resource of kind '%s', with group '%s' and version '%s'",
			ErrInvalidResource,
			defaultWorkloadGVK.Kind,
			defaultWorkloadGVK.Group,
			defaultWorkloadGVK.Version,
			workload.GetObjectKind().GroupVersionKind().Kind,
			workload.GetObjectKind().GroupVersionKind().Group,
			workload.GetObjectKind().GroupVersionKind().Version,
		)
	}

	return nil
}
`
