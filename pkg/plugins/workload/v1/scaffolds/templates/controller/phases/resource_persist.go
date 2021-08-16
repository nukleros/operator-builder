package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &ResourcePersist{}

// ResourcePersist scaffolds the resource persist phase methods.
type ResourcePersist struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *ResourcePersist) SetTemplateDefaults() error {
	f.Path = filepath.Join("controllers", "phases", "resource_persist.go")

	f.TemplateBody = resourcePersistTemplate

	return nil
}

const resourcePersistTemplate = `{{ .Boilerplate }}

package phases

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

func persistExitSuccessCondition(objectName string, objectKind string) *common.Condition {
	return &common.Condition{
		Phase:   common.ConditionPhaseCreateResources,
		Type:    common.ConditionTypeReconciling,
		Status:  common.ConditionStatusTrue,
		Message: "Created " + objectName + " " + objectKind,
	}
}

// PersistResourcePhase.Execute executes persisting resources to the Kubernetes database.
func (phase *PersistResourcePhase) Execute(resource *ComponentResource) (ctrl.Result, bool, error) {
	// if we are skipping resource creation, return immediately
	if resource.Skip {
		return ctrl.Result{}, true, nil
	}

	// if we are replacing resources, use the replaced resources, else use the original resources
	if len(resource.ReplacedResources) > 0 {
		for _, replacedResource := range resource.ReplacedResources {
			if err := persistResource(resource.ComponentReconciler, replacedResource); err != nil {
				return ctrl.Result{}, false, err
			}
		}
	} else {
		if err := persistResource(resource.ComponentReconciler, resource.OriginalResource); err != nil {
			return ctrl.Result{}, false, err
		}
	}

	return ctrl.Result{}, true, nil
}

// persistResource persists a single resource to the Kubernetes database.
func persistResource(
	r common.ComponentReconciler,
	resource metav1.Object,
) error {
	objectName := resource.(metav1.Object).GetName()
	objectKind := resource.(runtime.Object).GetObjectKind().GroupVersionKind().Kind

	// persist resource
	if err := r.CreateOrUpdate(resource); err != nil {
		if isOptimisticLockError(err) {
			return nil
		} else {
			r.GetLogger().V(0).Info("failed persisting object of kind: " + objectKind + " with name: " + objectName)

			return err
		}
	}

	// update the condition to notify that we have created a child resource
	return updateStatusConditions(r, persistExitSuccessCondition(objectName, objectKind))
}
`
