package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &CheckReady{}

// CheckReady scaffolds the check ready phase methods
type CheckReady struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *CheckReady) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "check_ready.go")

	f.TemplateBody = checkReadyTemplate

	return nil
}

var checkReadyTemplate = `{{ .Boilerplate }}

package phases

import (
	ctrl "sigs.k8s.io/controller-runtime"

	common "{{ .Repo }}/apis/common"
)

// GetSuccessCondition defines the success condition for the phase
func (phase *CheckReadyPhase) GetSuccessCondition() common.Condition {
	return common.Condition{
		Phase:   common.ConditionPhaseCheckReady,
		Type:    common.ConditionTypeReconciling,
		Status:  common.ConditionStatusTrue,
		Message: "Completed Phase " + string(common.ConditionPhaseCheckReady),
	}
}

// GetPendingCondition defines the pending condition for the phase
func (phase *CheckReadyPhase) GetPendingCondition() common.Condition {
	return common.Condition{
		Phase:   common.ConditionPhaseCheckReady,
		Type:    common.ConditionTypePending,
		Status:  common.ConditionStatusTrue,
		Message: "Component is Not Ready",
	}
}

// GetFailCondition defines the fail condition for the phase
func (phase *CheckReadyPhase) GetFailCondition() common.Condition {
	return common.Condition{
		Phase:   common.ConditionPhasePreFlight,
		Type:    common.ConditionTypeFailed,
		Status:  common.ConditionStatusTrue,
		Message: "Failed Phase " + string(common.ConditionPhaseCheckReady),
	}
}

// GetDefaultRequeueResult defines the result return when a requeue is needed
func (phase *CheckReadyPhase) GetDefaultRequeueResult() ctrl.Result {
	return DefaultRequeueResult()
}

// CheckReadyPhase.Execute executes checking for a parent components readiness status
func (phase *CheckReadyPhase) Execute(
	r common.ComponentReconciler,
) (proceedToNextPhase bool, err error) {
	// mark the resource as ready and created
	ready, err := r.CheckReady()
	if err != nil || !ready {
		return false, err
	}

	return true, nil
}
`
