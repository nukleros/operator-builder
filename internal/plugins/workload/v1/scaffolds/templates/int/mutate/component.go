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

	return nil
}

const componentTemplate = `{{ .Boilerplate }}

package mutate

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"{{ .Repo }}/apis/common"
)

// {{ .Resource.Kind }}Mutate performs the logic to mutate resources that belong to the parent.
func {{ .Resource.Kind }}Mutate(
	reconciler common.ComponentReconciler,
	object client.Object,
) (replacedObjects []client.Object, skip bool, err error) {
	return []client.Object{object}, false, nil
}
`
