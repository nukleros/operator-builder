package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &ResourceMutate{}

// ResourceMutate scaffolds the resource mutate phase method.
type ResourceMutate struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
}

func (f *ResourceMutate) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "resource_mutate.go")

	f.TemplateBody = resourceMutateTemplate

	return nil
}

const resourceMutateTemplate = `{{ .Boilerplate }}

package phases

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

// MutateResourcePhase.Execute executes the mutation of a resource.
func (phase *MutateResourcePhase) Execute(resource *ComponentResource) (ctrl.Result, bool, error) {
	replacedResources, skip, err := resource.ComponentReconciler.Mutate(resource.OriginalResource)
	if err != nil {
		return ctrl.Result{}, false, err
	}

	resource.ReplacedResources = replacedResources
	resource.Skip = skip

	return ctrl.Result{}, true, nil
}
`
