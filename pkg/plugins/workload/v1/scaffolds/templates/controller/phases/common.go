package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Common{}

// Common scaffolds common phase operations.
type Common struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Common) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "common.go")

	f.TemplateBody = commonTemplate

	return nil
}

const commonTemplate = `{{ .Boilerplate }}

package phases

import (
	"fmt"
	"reflect"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

const optimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"

// Requeue will return the default result to requeue a reconciler request when needed.
func Requeue() ctrl.Result {
	return ctrl.Result{Requeue: true}
}

// DefaultReconcileResult will return the default reconcile result when requeuing is not needed.
func DefaultReconcileResult() ctrl.Result {
	return ctrl.Result{}
}

// conditionExists will return whether or not a specific condition already exists on the object.
func conditionExists(
	currentConditions []common.Condition,
	condition *common.Condition,
) bool {

	for _, currentCondition := range currentConditions {
		if reflect.DeepEqual(currentCondition, *condition) {
			return true
		}
	}
	return false
}

// updateStatusConditions updates the status.conditions field of the parent custom resource.
func updateStatusConditions(
	r common.ComponentReconciler,
	condition *common.Condition,
) error {
	component := r.GetComponent()

	if !conditionExists(component.GetStatusConditions(), condition) {
		component.SetStatusConditions(*condition)

		if err := r.UpdateStatus(r.GetContext(), component); err != nil {
			return err
		}
	}

	return nil
}

// handlePhaseExit will perform the steps required to exit a phase.
func HandlePhaseExit(
	reconciler common.ComponentReconciler,
	phaseHandler PhaseHandler,
	phaseIsReady bool,
	phaseError error,
) (ctrl.Result, error) {
	var condition common.Condition
	var result ctrl.Result

	switch {
	case phaseError != nil:
		condition = phaseHandler.GetFailCondition()
		result = DefaultReconcileResult()
	case !phaseIsReady:
		condition = phaseHandler.GetPendingCondition()
		result = Requeue()
	default:
		condition = phaseHandler.GetSuccessCondition()
		result = DefaultReconcileResult()
	}

	// update the status conditions and return any errors
	if updateError := updateStatusConditions(reconciler, &condition); updateError != nil {
		// adjust the message if we had both an update error and a phase error
		if phaseError != nil {
			phaseError = fmt.Errorf("failed to update status conditions; %v; %v", updateError, phaseError)
		}
	}

	return result, phaseError
}

// isOptimisticLockError checks to see if the error is a locking error.
func isOptimisticLockError(err error) bool {
	return strings.Contains(err.Error(), optimisticLockErrorMsg)
}

// setResources will set the resources against a CreateResourcePhase.
func setResources(
	parent *CreateResourcesPhase,
	resources []metav1.Object,
) {
	parent.Resources = resources
}

// getResources will get the resources from a CreateResourcePhase.
func getResources(
	parent *CreateResourcesPhase,
) []metav1.Object {
	return parent.Resources
}
`
