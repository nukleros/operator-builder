package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &ResourceCreateInMemory{}

// ResourceCreateInMemory scaffolds the create resource in memory phase methods.
type ResourceCreateInMemory struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *ResourceCreateInMemory) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "resource_create_in_memory.go")

	f.TemplateBody = resourceCreateInMemoryTemplate

	return nil
}

const resourceCreateInMemoryTemplate = `{{ .Boilerplate }}

package phases

import (
	"{{ .Repo }}/apis/common"
)

// CreateResourcesInMemoryPhase.Execute executes the creation of resources in memory, prior to mutating or persisting them to the database
func (phase *CreateResourcesInMemoryPhase) Execute(
	r common.ComponentReconciler,
	parentPhase *CreateResourcesPhase,
) (proceedToNextPhase bool, err error) {
	resources, err := r.GetResources(r.GetComponent())
	if err != nil {
		return false, err
	}

	// update the resources on the parent phase object
	setResources(parentPhase, resources)

	return true, nil
}
`
