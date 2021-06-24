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

// ConditionPhase defines the phase in which the condition was set
// +kubebuilder:validation:Enum=Dependency;PreFlight;CreateResources;Mutate;Persist;Wait;CheckReady;Complete
type ConditionPhase string

const (
	ConditionPhaseDependency      ConditionPhase = "Dependency"
	ConditionPhasePreFlight       ConditionPhase = "PreFlight"
	ConditionPhaseCreateResources ConditionPhase = "CreateResources"
	ConditionPhaseMutate          ConditionPhase = "Mutate"
	ConditionPhasePersist         ConditionPhase = "Persist"
	ConditionPhaseWait            ConditionPhase = "Wait"
	ConditionPhaseCheckReady      ConditionPhase = "CheckReady"
	ConditionPhaseComplete        ConditionPhase = "Complete"
)

// ConditionType defines the type of condition
// +kubebuilder:validation:Enum=Ready;Reconciling;Failed;Pending
type ConditionType string

const (
	ConditionTypeReady       ConditionType = "Ready"
	ConditionTypeReconciling ConditionType = "Reconciling"
	ConditionTypeFailed      ConditionType = "Failed"
	ConditionTypePending     ConditionType = "Pending"
)

// ConditionStatus defines the status of the condition
// +kubebuilder:validation:Enum=True;False
type ConditionStatus string

const (
	ConditionStatusTrue  ConditionStatus = "True"
	ConditionStatusFalse ConditionStatus = "False"
)

// Condition describes an event that has occurred against the object
type Condition struct {
	Type    ConditionType   ` + "`" + `json:"type"` + "`" + `
	Status  ConditionStatus ` + "`" + `json:"status"` + "`" + `
	Phase   ConditionPhase  ` + "`" + `json:"phase"` + "`" + `
	Message string          ` + "`" + `json:"message"` + "`" + `
}
`
