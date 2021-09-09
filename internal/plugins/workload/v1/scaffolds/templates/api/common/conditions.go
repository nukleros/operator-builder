package common

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Conditions{}

// Conditions scaffolds the conditions for all workloads.
type Conditions struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
}

func (f *Conditions) SetTemplateDefaults() error {
	f.Path = filepath.Join("apis", "common", "conditions.go")

	f.TemplateBody = conditionsTemplate

	return nil
}

const conditionsTemplate = `{{ .Boilerplate }}

package common

// PhaseState defines the current state of the phase.
// +kubebuilder:validation:Enum=Complete;Reconciling;Failed;Pending
type PhaseState string

const (
	PhaseStatePending     PhaseState = "Pending"
	PhaseStateReconciling PhaseState = "Reconciling"
	PhaseStateFailed      PhaseState = "Failed"
	PhaseStateComplete    PhaseState = "Complete"
)

// PhaseCondition describes an event that has occurred during a phase
// of the controller reconciliation loop.
type PhaseCondition struct {
	State PhaseState ` + "`" + `json:"state"` + "`" + `

	// Phase defines the phase in which the condition was set.
	Phase string ` + "`" + `json:"phase"` + "`" + `

	// Message defines a helpful message from the phase.
	Message string ` + "`" + `json:"message"` + "`" + `

	// LastModified defines the time in which this component was updated.
	LastModified string ` + "`" + `json:"lastModified"` + "`" + `
}

// ResourceCondition describes the condition of a Kubernetes resource managed by the parent object.
type ResourceCondition struct {
	// Created defines whether this object has been successfully created or not.
	Created bool ` + "`" + `json:"created"` + "`" + `

	// LastResourcePhase defines the last successfully completed resource phase.
	LastResourcePhase string ` + "`" + `json:"lastResourcePhase,omitempty"` + "`" + `

	// LastModified defines the time in which this resource was updated.
	LastModified string ` + "`" + `json:"lastModified,omitempty"` + "`" + `

	// Message defines a helpful message from the resource phase.
	Message string ` + "`" + `json:"message,omitempty"` + "`" + `
}

// GetPhaseConditionIndex returns the index of a matching phase condition.  Any integer which is 0
// or greater indicates that the phase condition was found.  Anything lower indicates that an
// associated condition is not found.
func (condition *PhaseCondition) GetPhaseConditionIndex(component Component) int {
	for i, currentCondition := range component.GetPhaseConditions() {
		if currentCondition.Phase == condition.Phase {
			return i
		}
	}

	return -1
}
`
