package helpers

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/utils"
)

var _ machinery.Template = &Common{}

// Component scaffolds the workload's helper functions.
type Component struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin
	machinery.DomainMixin
}

func (f *Component) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"pkg",
		"helpers",
		fmt.Sprintf("%s.go", utils.ToFileName(f.Resource.Kind)),
	)
	f.TemplateBody = componentTemplate

	return nil
}

const componentTemplate = `{{ .Boilerplate }}

package helpers

import (
	"fmt"

	common "{{ .Repo }}/apis/common"
	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
)

// {{ .Resource.Kind }}Unique returns only one {{ .Resource.Kind }} and returns an error if more than one are found.
func {{ .Resource.Kind }}Unique(
	reconciler common.ComponentReconciler,
) (
	*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	error,
) {
	components, err := {{ .Resource.Kind }}List(reconciler)
	if err != nil {
		return nil, err
	}

	if len(components.Items) != 1 {
		return nil, fmt.Errorf("expected only 1 {{ .Resource.Kind }}; found %v\n", len(components.Items))
	}

	component := components.Items[0]

	return &component, nil
}

// {{ .Resource.Kind }}List gets a {{ .Resource.Kind }}List from the cluster.
func {{ .Resource.Kind }}List(
	reconciler common.ComponentReconciler,
) (
	*{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}List,
	error,
) {
	components := &{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}List{}
	if err := reconciler.List(reconciler.GetContext(), components); err != nil {
		reconciler.GetLogger().V(0).Info("unable to retrieve {{ .Resource.Kind }}List from cluster")

		return nil, err
	}

	return components, nil
}
`
