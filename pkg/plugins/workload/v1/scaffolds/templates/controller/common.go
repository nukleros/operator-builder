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
	"fmt"
	"time"
	"reflect"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"{{ .Repo }}/apis/common"
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
		{{- if not .IsStandalone }}
		&controllerphases.DependencyPhase{},
		&controllerphases.PreFlightPhase{},
		{{ end -}}
		&controllerphases.CreateResourcesPhase{},
		{{- if not .IsStandalone }}
		&controllerphases.CheckReadyPhase{},
		&controllerphases.CompletePhase{},
		{{ end -}}
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

// reconcileUpdaters is a list of managers which produce a reconciliation on update.
// TODO: this solves for the 95% of cases, but there will inevitably by corner cases in which
// are not addressed.
func reconcileUpdaters() []string {
	return []string{
		"kubectl",
		"kapp",
		"helm",
		"ansible",
		"ansible-playbook",
	}
}

// ResourcePredicates returns the filters which are used to filter out the common reconcile events
// prior to reconciling the child resource of a component.
func ResourcePredicates() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// return immediately if the managed fields are the same
			if reflect.DeepEqual(e.ObjectNew.GetManagedFields(), e.ObjectOld.GetManagedFields()) {
				return false
			}

			// return immediately if both objects are the same
			if reflect.DeepEqual(e.ObjectNew, e.ObjectOld) {
				return false
			}

			// if we have a non-reconciler update return
			var numWhitelisted int
			var justUpdated bool
			for _, new := range e.ObjectNew.GetManagedFields() {
				if new.Operation == v1.ManagedFieldsOperationUpdate {
					// count the number of whitelisted managers which we know we need to update for
					for _, updater := range reconcileUpdaters() {
						if new.Manager == updater {
							numWhitelisted++
						}
					}

					// if our manager is the reconciler, see if it was just updated
					// TODO: time boxing the update is not ideal, however it is much simpler than
					// doing a deep compare at this moment.  we can improve this logic at a later time.
					if new.Manager == FieldManager {
						if time.Now().UTC().Sub(new.Time.Time.UTC()) < 2*time.Second {
							justUpdated = true
						}
					}
				}
			}

			return !justUpdated && numWhitelisted > 0
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return !e.DeleteStateUnknown
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
			ResourcePredicates(),
		); err != nil {
			return err
		}

		r.SetWatch(resource)
	}

	return nil
}

// Create creates a resource.
func Create(
	r common.ComponentReconciler,
	newResource client.Object,
) error {
	r.GetLogger().V(0).Info(fmt.Sprintf("creating resource with name: [%s] in namespace: [%s] of kind: [%s]",
		newResource.GetName(), newResource.GetNamespace(), newResource.GetObjectKind().GroupVersionKind().Kind))

	if err := r.Create(r.GetContext(), newResource, &client.CreateOptions{FieldManager: FieldManager}); err != nil {
		r.GetLogger().V(0).Info("unable to create resource")

		return err
	}

	return nil
}

// Update updates a resource.
func Update(
	r common.ComponentReconciler,
	newResource client.Object,
	oldResource client.Object,
) error {
	r.GetLogger().V(0).Info(fmt.Sprintf("updating resource with name: [%s] in namespace: [%s] of kind: [%s]",
		newResource.GetName(), newResource.GetNamespace(), newResource.GetObjectKind().GroupVersionKind().Kind))

	if err := r.Patch(r.GetContext(), newResource, &client.Merge, &client.PatchOptions{FieldManager: FieldManager}); err != nil {
		r.GetLogger().V(0).Info("unable to update resource")

		return err
	}

	return nil
}
`
