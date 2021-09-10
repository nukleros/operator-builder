// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package wait

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

var _ machinery.Template = &Component{}

// Component scaffolds the workload's wait function.
type Component struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin
}

func (f *Component) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"wait",
		fmt.Sprintf("%s.go", utils.ToFileName(f.Resource.Kind)),
	)

	f.TemplateBody = componentTemplate

	return nil
}

const componentTemplate = `{{ .Boilerplate }}

package wait

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"{{ .Repo }}/apis/common"
)

// {{ .Resource.Kind }}Wait performs the logic to wait for resources that belong to the parent.
func {{ .Resource.Kind }}Wait(reconciler common.ComponentReconciler,
	object *metav1.Object,
) (ready bool, err error) {
	return true, nil
}
`

