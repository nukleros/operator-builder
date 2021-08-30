package controller

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Common{}

// Common scaffolds controller utilities common to all controllers.
type Common struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	IsStandalone bool
}

func (f *Common) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "common.go")

	f.TemplateBody = controllerCommonTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const controllerCommonTemplate = `{{ .Boilerplate }}

package controllers

import (
	"reflect"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"{{ .Repo }}/apis/common"
	"{{ .Repo }}/pkg/resources"
	controllerphases "{{ .Repo }}/controllers/phases"
)

const (
	FieldManager = "reconciler"
)

func IgnoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}

	return err
}

// CreatePhases defines the phases for create and the order in which they run during the reconcile process.
func CreatePhases() []controllerphases.Phase {
	return []controllerphases.Phase{
		&controllerphases.DependencyPhase{},
		&controllerphases.PreFlightPhase{},
		&controllerphases.CreateResourcesPhase{},
		&controllerphases.CheckReadyPhase{},
		&controllerphases.CompletePhase{},
	}
}

// UpdatePhases defines the phases for update and the order in which they run during the reconcile process.
func UpdatePhases() []controllerphases.Phase {
	// at this time create/update are identical; return the create phases
	return CreatePhases()
}

// Phases returns which phases to run given the component.
func Phases(component common.Component) []controllerphases.Phase {
	var phases []controllerphases.Phase
	if !component.GetReadyStatus() {
		phases = CreatePhases()
	} else {
		phases = UpdatePhases()
	}

	return phases
}

// getDesiredObject returns the desired object from a list stored on the
// reconciler.
func getDesiredObject(compared *resources.Resource) (desired *resources.Resource) {
	for _, resource := range compared.Reconciler.GetResources() {
		if resource.EqualGVK(compared) && resource.EqualNamespaceName(compared) {
			return resource.(*resources.Resource)
		}
	}

	return desired
}

// needsReconciliation performs some simple checks and returns whether or not a
// resource needs to be updated.
func needsReconciliation(existing, requested resources.Resource) bool {
	// skip if the resources versions are the same
	if existing.Object.GetResourceVersion() == requested.Object.GetResourceVersion() {
		return false
	}

	// skip if the objects support observed generation and they are equal
	if existing.Object.GetGeneration() > 0 && requested.Object.GetGeneration() > 0 {
		if existing.Object.GetGeneration() == requested.Object.GetGeneration() {
			return false
		}
	}

	// get the desired object from the reconciler and ensure that we both
	// found that desired object and that the desired object fields are equal
	// to the existing object fields
	desired := getDesiredObject(&requested)
	if desired == nil {
		return true
	}

	equal, err := resources.AreEqual(*desired, requested)
	if err != nil {
		requested.Reconciler.GetLogger().V(0).Error(err,
			"unable to determine equality for reconciliation")

		return true
	}

	return !equal
}

// ResourcePredicates returns the filters which are used to filter out the common reconcile events
// prior to reconciling the child resource of a component.
func ResourcePredicates(r common.ComponentReconciler) predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return needsReconciliation(
				*resources.NewResourceFromClient(e.ObjectOld, r),
				*resources.NewResourceFromClient(e.ObjectNew, r),
			)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			// do not run reconciliation on unknown events
			return false
		},
		CreateFunc: func(e event.CreateEvent) bool {
			// do not run reconciliation again when we just created the child resource
			return false
		},
	}
}

// ComponentPredicates returns the filters which are used to filter out the common reconcile events
// prior to reconciling an object for a component.
func ComponentPredicates() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return !e.DeleteStateUnknown
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

// Watch watches a resource.
func Watch(
	r common.ComponentReconciler,
	resource client.Object,
) error {
	// check if the resource is already being watched
	var watched bool

	if len(r.GetWatches()) > 0 {
		for _, watcher := range r.GetWatches() {
			if reflect.DeepEqual(
				resource.GetObjectKind().GroupVersionKind(),
				watcher.GetObjectKind().GroupVersionKind(),
			) {
				watched = true
			}
		}
	}

	// watch the resource if it current is not being watched
	if !watched {
		if err := r.GetController().Watch(
			&source.Kind{Type: resource},
			&handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    r.GetComponent().(runtime.Object),
			},
			ResourcePredicates(r),
		); err != nil {
			return err
		}

		r.SetWatch(resource)
	}

	return nil
}
`
