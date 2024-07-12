// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package dependencies

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/nukleros/operator-builder/internal/utils"
)

var _ machinery.Template = &Component{}

// Component scaffolds the workload's check ready function that is called by
// components with a dependency on the workload.
type Component struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin
}

func (f *Component) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"dependencies",
		fmt.Sprintf("%s.go", utils.ToFileName(f.Resource.Kind)),
	)

	f.TemplateBody = componentTemplate

	f.IfExistsAction = machinery.SkipFile

	return nil
}

const componentTemplate = `{{ .Boilerplate }}

package dependencies

import (
	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"
)

// {{ .Resource.Kind }}CheckReady performs the logic to determine if a {{ .Resource.Kind }} object is ready.
func {{ .Resource.Kind }}CheckReady(r workload.Reconciler, req *workload.Request) (bool, error) {
	return true, nil
}
`
