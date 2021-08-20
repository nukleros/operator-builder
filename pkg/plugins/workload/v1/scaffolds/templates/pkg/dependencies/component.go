package dependencies

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/utils"
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
		"pkg",
		"dependencies",
		fmt.Sprintf("%s.go", utils.ToFileName(f.Resource.Kind)),
	)

	f.TemplateBody = componentTemplate

	return nil
}

const componentTemplate = `{{ .Boilerplate }}

package dependencies

import (
	"{{ .Repo }}/apis/common"
)

// {{ .Resource.Kind }}CheckReady performs the logic to determine if a {{ .Resource.Kind }} object is ready.
func {{ .Resource.Kind }}CheckReady(reconciler common.ComponentReconciler) (bool, error) {
	return true, nil
}
`
