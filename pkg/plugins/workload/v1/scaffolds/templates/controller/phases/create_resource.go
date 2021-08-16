package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &CreateResource{}

// CreateResource scaffolds the create resource phase methods.
type CreateResource struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin

	IsStandalone bool
}

func (f *CreateResource) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "create_resource.go")

	f.TemplateBody = createResourceTemplate

	return nil
}

const createResourceTemplate = `{{ .Boilerplate }}

package phases

import (
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

// GetSuccessCondition defines the success condition for the phase.
func (phase *CreateResourcesPhase) GetSuccessCondition() common.Condition {
	return common.Condition{
		Phase:   common.ConditionPhaseCreateResources,
		Type:    common.ConditionTypeReconciling,
		Status:  common.ConditionStatusTrue,
		Message: "Completed Phase " + string(common.ConditionPhaseCreateResources),
	}
}

// GetPendingCondition defines the pending condition for the phase.
func (phase *CreateResourcesPhase) GetPendingCondition() common.Condition {
	return common.Condition{
		Phase:   common.ConditionPhaseCreateResources,
		Type:    common.ConditionTypePending,
		Status:  common.ConditionStatusTrue,
		Message: "Resources Not Finished Creating",
	}
}

// GetFailCondition defines the fail condition for the phase.
func (phase *CreateResourcesPhase) GetFailCondition() common.Condition {
	return common.Condition{
		Phase:   common.ConditionPhaseCreateResources,
		Type:    common.ConditionTypeFailed,
		Status:  common.ConditionStatusTrue,
		Message: "Failed Phase " + string(common.ConditionPhaseCreateResources),
	}
}

// Requeue defines the result return when a requeue is needed.
func (phase *CreateResourcesPhase) Requeue() ctrl.Result {
	return Requeue()
}

// createResourcePhases defines the phases for resource creation and the order in which they run during the reconcile process.
func createResourcePhases() []ResourcePhase {
	return []ResourcePhase{
		{{- if not .IsStandalone }}
		// wait for other resources before attempting to create
		&WaitForResourcePhase{},

		// update fields within a resource
		&MutateResourcePhase{},
		{{ end }}
		// create the resource in the cluster
		&PersistResourcePhase{},
	}
}

// CreateResourcesPhase.Execute executes executes sub-phases which are required to create the resources.
func (phase *CreateResourcesPhase) Execute(
	r common.ComponentReconciler,
) (proceedToNextPhase bool, err error) {
	r.GetLogger().V(2).Info("constructing resources in memory")

	proceed, err := new(ConstructPhase).Execute(r, phase)
	if err != nil || !proceed {
		return false, err
	}

	// execute the resource phases against each resource
	for _, resource := range phase.Resources {
		componentResource := &ComponentResource{
			ComponentReconciler: r,
			OriginalResource:    resource,
		}

		for _, resourcePhase := range createResourcePhases() {
			r.GetLogger().V(7).Info(fmt.Sprintf("enter resource phase: %T", resourcePhase))
			_, proceed, err := resourcePhase.Execute(componentResource)

			// return the error and result on error
			if err != nil || !proceed {
				return false, err
			}

			r.GetLogger().V(5).Info(fmt.Sprintf("completed resource phase: %T", resourcePhase))
		}
	}

	return true, nil
}
`
