// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

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
	f.Path = filepath.Join("internal", "controllers", "phases", "phases.go")

	f.TemplateBody = commonTemplate

	return nil
}

const commonTemplate = `{{ .Boilerplate }}

package phases

import (
	"fmt"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

const optimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"

// Requeue will return the default result to requeue a reconciler request when needed.
func Requeue() ctrl.Result {
	return ctrl.Result{Requeue: true}
}

// IsOptimisticLockError checks to see if the error is a locking error.
func IsOptimisticLockError(err error) bool {
	return strings.Contains(err.Error(), optimisticLockErrorMsg)
}

// DefaultReconcileResult will return the default reconcile result when requeuing is not needed.
func DefaultReconcileResult() ctrl.Result {
	return ctrl.Result{}
}

// updatePhaseConditions updates the status.conditions field of the parent custom resource.
func updatePhaseConditions(
	r common.ComponentReconciler,
	condition *common.PhaseCondition,
) error {
	r.GetComponent().SetPhaseCondition(*condition)

	return r.UpdateStatus()
}

// updateResourceConditions updates the status.resourceConditions field of the parent custom resource.
func updateResourceConditions(
	r common.ComponentReconciler,
	resource common.Resource,
	condition *common.ResourceCondition,
) error {
	resource.ResourceCondition = *condition
	r.GetComponent().SetResource(resource)

	return r.UpdateStatus()
}

// HandlePhaseExit will perform the steps required to exit a phase.
func HandlePhaseExit(
	reconciler common.ComponentReconciler,
	phase Phase,
	phaseIsReady bool,
	phaseError error,
) (ctrl.Result, error) {

	var condition common.PhaseCondition
	var result ctrl.Result

	switch {
	case phaseError != nil:
		if IsOptimisticLockError(phaseError) {
			phaseError = nil
			condition = GetSuccessCondition(phase)
		} else {
			condition = GetFailCondition(phase, phaseError)
		}
		result = DefaultReconcileResult()
	case !phaseIsReady:
		condition = GetPendingCondition(phase)
		result = phase.DefaultRequeue()
	default:
		condition = GetSuccessCondition(phase)
		result = DefaultReconcileResult()
	}

	// update the status conditions and return any errors
	if updateError := updatePhaseConditions(reconciler, &condition); updateError != nil {
		// adjust the message if we had both an update error and a phase error
		if !IsOptimisticLockError(updateError) {
			if phaseError != nil {
				phaseError = fmt.Errorf("failed to update status conditions; %v; %v", updateError, phaseError)
			} else {
				phaseError = updateError
			}
		}
	}

	return result, phaseError
}

// handleResourcePhaseExit will perform the steps required to exit a phase.
func handleResourcePhaseExit(
	reconciler common.ComponentReconciler,
	resource common.Resource,
	condition common.ResourceCondition,
	phase ResourcePhase,
	phaseIsReady bool,
	phaseError error,
) (bool, error) {

	switch {
	case phaseError != nil:
		if IsOptimisticLockError(phaseError) {
			phaseError = nil
		}
	case !phaseIsReady:
		condition.Message = fmt.Sprintf("unable to proceed with resource creation; phase %v is not ready", getResourcePhaseName(phase))
	}

	// update the status conditions and return any errors
	if updateError := updateResourceConditions(reconciler, resource, &condition); updateError != nil {
		// adjust the message if we had both an update error and a phase error
		if !IsOptimisticLockError(updateError) {
			if phaseError != nil {
				phaseError = fmt.Errorf("failed to update resource conditions; %v; %v", updateError, phaseError)
			} else {
				phaseError = updateError
			}
		}
	} else {
		condition.Message = "resource creation successful"
	}

	return (phaseError == nil && phaseIsReady), phaseError
}
`
