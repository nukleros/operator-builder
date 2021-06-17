package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &PreFlight{}

// PreFlight scaffolds the pre-flight phase methods
type PreFlight struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *PreFlight) SetTemplateDefaults() error {

	f.Path = filepath.Join("controllers", "phases", "pre_flight.go")

	f.TemplateBody = preFlightTemplate

	return nil
}

var preFlightTemplate = `{{ .Boilerplate }}

package phases

import (
	ctrl "sigs.k8s.io/controller-runtime"

	common "{{ .Repo }}/apis/common"
)

// GetSuccessCondition defines the success condition for the phase
func (phase *PreFlightPhase) GetSuccessCondition() common.Condition {
	return common.Condition{
		Phase:   common.ConditionPhasePreFlight,
		Type:    common.ConditionTypeReconciling,
		Status:  common.ConditionStatusTrue,
		Message: "Completed Phase " + string(common.ConditionPhasePreFlight),
	}
}

// GetPendingCondition defines the pending condition for the phase
func (phase *PreFlightPhase) GetPendingCondition() common.Condition {
	return common.Condition{
		Phase:   common.ConditionPhasePreFlight,
		Type:    common.ConditionTypePending,
		Status:  common.ConditionStatusTrue,
		Message: "Unable to Continue Phase " + string(common.ConditionPhasePreFlight),
	}
}

// GetFailCondition defines the fail condition for the phase
func (phase *PreFlightPhase) GetFailCondition() common.Condition {
	return common.Condition{
		Phase:   common.ConditionPhasePreFlight,
		Type:    common.ConditionTypeFailed,
		Status:  common.ConditionStatusTrue,
		Message: "Failed Phase " + string(common.ConditionPhasePreFlight),
	}
}

// GetDefaultRequeueResult defines the result return when a requeue is needed
func (phase *PreFlightPhase) GetDefaultRequeueResult() ctrl.Result {
	return DefaultRequeueResult()
}

// PreFlightPhase.Execute executes pre-flight and fail-fast conditions prior to attempting resource creation
func (phase *PreFlightPhase) Execute(
	r common.ComponentReconciler,
) (proceedToNextPhase bool, err error) {
	return true, nil
}
`
