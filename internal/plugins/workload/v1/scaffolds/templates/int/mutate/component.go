// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package mutate

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

var _ machinery.Template = &Component{}

// Component scaffolds the workload's mutate function.
type Component struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin
}

func (f *Component) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"mutate",
		fmt.Sprintf("%s.go", utils.ToFileName(f.Resource.Kind)),
	)

	f.TemplateBody = componentTemplate

	f.IfExistsAction = machinery.SkipFile

	return nil
}

const componentTemplate = `{{ .Boilerplate }}

package mutate

import (
	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// {{ .Resource.Kind }}Mutate performs the logic to mutate resources that belong to the parent.
func {{ .Resource.Kind }}Mutate(
	r workload.Reconciler,
	req *workload.Request,
	object client.Object,
) (replacedObjects []client.Object, skip bool, err error) {
	return []client.Object{object}, false, nil
}
`
