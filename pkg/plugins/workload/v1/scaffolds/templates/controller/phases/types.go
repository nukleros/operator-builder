package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Types{}

// Types scaffolds the phase interfaces and types.
type Types struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Types) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "types.go")

	f.TemplateBody = typesTemplate

	return nil
}

const typesTemplate = `{{ .Boilerplate }}

package phases

import (
	"context"

	"{{ .Repo }}/apis/common"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Phase defines a phase of the reconciliation process
type Phase interface {
	Execute(common.ComponentReconciler) (bool, error)
}

// PhaseHandler defines an object which can handle the outcome of a phase execution
type PhaseHandler interface {
	GetSuccessCondition() common.Condition
	GetPendingCondition() common.Condition
	GetFailCondition() common.Condition
	Requeue() ctrl.Result
}

// ResourcePhase defines the specific phase of reconcilication associated with creating resources
type ResourcePhase interface {
	Execute(*ComponentResource) (ctrl.Result, bool, error)
}

// ComponentResource defines a resource which is created by the parent Component custom resource
type ComponentResource struct {
	Component           *common.Component
	ComponentReconciler common.ComponentReconciler
	Context             context.Context
	OriginalResource    *metav1.Object
	ReplacedResources   []metav1.Object
	Skip                bool
}

// DependencyPhase defines an object specific to the depenency phase of reconciliation
type DependencyPhase struct{}

// DependencyPhase defines an object specific to the preflight phase of reconciliation
type PreFlightPhase struct{}

// CreateResourcesPhase defines an object specific to the create resources phase of reconciliation
type CreateResourcesPhase struct {
	Resources []metav1.Object
	ReplacedResources []metav1.Object
}

// ConstructPhase defines an object specific to the in memory resource creation phase of reconciliation
type ConstructPhase struct{}

// MutateResourcePhase defines an object specific to the resource mutation phase of reconciliation
type MutateResourcePhase struct{}

// PersistResourcePhase defines an object specific to the resource persistence phase of reconciliation
type PersistResourcePhase struct{}

// WaitForResourcePhase defines an object specific to the resource waiting phase of reconciliation
type WaitForResourcePhase struct{}

// CheckReadyPhase defines an object specific to the resource checking to see if the object is ready
type CheckReadyPhase struct{}

// CompletePhase defines an object specific to the completion phase of reconciliation
type CompletePhase struct{}
`
