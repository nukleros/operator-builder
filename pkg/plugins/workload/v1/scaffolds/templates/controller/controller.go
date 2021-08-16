package controller

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

var _ machinery.Template = &Controller{}

// Controller scaffolds the workload's controller.
type Controller struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	PackageName       string
	RBACRules         *[]workloadv1.RBACRule
	OwnershipRules    *[]workloadv1.OwnershipRule
	HasChildResources bool
	IsStandalone      bool
	IsComponent       bool
	Collection        *workloadv1.WorkloadCollection
}

func (f *Controller) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"controllers",
		f.Resource.Group,
		fmt.Sprintf("%s_controller.go", utils.ToFileName(f.Resource.Kind)),
	)

	f.TemplateBody = controllerTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

//nolint: lll
const controllerTemplate = `{{ .Boilerplate }}

package {{ .Resource.Group }}

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"{{ .Repo }}/apis/common"
	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	{{ if .IsComponent -}}
	{{ .Collection.Spec.APIGroup }}{{ .Collection.Spec.APIVersion }} "{{ .Repo }}/apis/{{ .Collection.Spec.APIGroup }}/{{ .Collection.Spec.APIVersion }}"
	{{ end }}
	{{- if .HasChildResources -}}
	"{{ .Resource.Path }}/{{ .PackageName }}"
	{{ end -}}
	"{{ .Repo }}/controllers"
	"{{ .Repo }}/controllers/phases"
	{{- if not .IsStandalone }}
	"{{ .Repo }}/pkg/dependencies"
	"{{ .Repo }}/pkg/mutate"
	"{{ .Repo }}/pkg/wait"
	{{ end }}
)

// {{ .Resource.Kind }}Reconciler reconciles a {{ .Resource.Kind }} object.
type {{ .Resource.Kind }}Reconciler struct {
	client.Client
	Name       string
	Log        logr.Logger
	Scheme     *runtime.Scheme
	Context    context.Context
	Controller controller.Controller
	Watches    []client.Object
	Component  *{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}
	{{- if .IsComponent }}
	Collection *{{ .Collection.Spec.APIGroup }}{{ .Collection.Spec.APIVersion }}.{{ .Collection.Spec.APIKind }}
	{{ end }}
}

// +kubebuilder:rbac:groups={{ .Resource.Group }}.{{ .Resource.Domain }},resources={{ .Resource.Plural }},verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups={{ .Resource.Group }}.{{ .Resource.Domain }},resources={{ .Resource.Plural }}/status,verbs=get;update;patch
{{ range .RBACRules -}}
// +kubebuilder:rbac:groups={{ .Group }},resources={{ .Resource }},verbs={{ .VerbString }}
{{ end }}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WebApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *{{ .Resource.Kind }}Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Context = ctx
	log := r.Log.WithValues("{{ .Resource.Kind | lower }}", req.NamespacedName)

	r.Component = &{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}{}
	if err := r.Get(r.Context, req.NamespacedName, r.Component); err != nil {
		log.V(0).Info("unable to fetch {{ .Resource.Kind }}")

		return ctrl.Result{}, controllers.IgnoreNotFound(err)
	}

	{{ if .IsComponent }}
	var collectionList {{ .Collection.Spec.APIGroup }}{{ .Collection.Spec.APIVersion }}.{{ .Collection.Spec.APIKind }}List

	if err := r.List(r.Context, &collectionList); err != nil {
		return ctrl.Result{}, err
	}

	if len(collectionList.Items) == 0 {
		log.V(0).Info("no collections available - will try again in 10 seconds")

		return ctrl.Result{Requeue: true}, nil
	} else if len(collectionList.Items) > 1 {
		log.V(0).Info("multiple collections found - cannot proceed")

		return ctrl.Result{}, nil
	}

	r.Collection = &collectionList.Items[0]
	{{ end }}

	// execute the phases
	for _, phase := range controllers.Phases(r.Component) {
		r.GetLogger().V(7).Info(fmt.Sprintf("enter phase: %T", phase))
		proceed, err := phase.Execute(r)
		result, err := phases.HandlePhaseExit(r, phase.(phases.PhaseHandler), proceed, err)

		// return only if we have an error or are told not to proceed
		if err != nil || !proceed {
			log.V(4).Info(fmt.Sprintf("not ready; requeuing phase: %T", phase))

			return result, err
		}

		r.GetLogger().V(5).Info(fmt.Sprintf("completed phase: %T", phase))
	}

	return phases.DefaultReconcileResult(), nil
}

// GetResources will create and return the resources in memory.
func (r *{{ .Resource.Kind }}Reconciler) GetResources(parent common.Component) ([]metav1.Object, error) {
{{ if .HasChildResources }}
	var resourceObjects []metav1.Object

	// create resources in memory
	for _, f := range {{ .PackageName }}.CreateFuncs {
		resource, err := f(r.Component{{ if .IsComponent }}, r.Collection){{ else }}){{ end }}
		if err != nil {
			return nil, err
		}

		resourceObjects = append(resourceObjects, resource)
	}

	return resourceObjects, nil
{{- else -}}
	return []metav1.Object{}, nil
{{ end -}}
}

// CreateOrUpdate creates a resource if it does not already exist or updates a resource
// if it does already exist.
func (r *{{ .Resource.Kind }}Reconciler) CreateOrUpdate(
	resource metav1.Object,
) error {
	// set ownership on the underlying resource being created or updated
	if err := ctrl.SetControllerReference(r.Component, resource, r.Scheme); err != nil {
		r.GetLogger().V(0).Info("unable to set owner reference on resource")

		return err
	}

	// create a stub object to store the current resource in the cluster so that we do not affect
	// the desired state of the resource object in memory
	newResource := resource.(client.Object)
	oldResource := &unstructured.Unstructured{}
	oldResource.SetGroupVersionKind(newResource.GetObjectKind().GroupVersionKind())

	objectKey := client.ObjectKeyFromObject(newResource)
	if err := r.Get(r.Context, objectKey, oldResource); err != nil {
		if errors.IsNotFound(err) {
			if err := controllers.Create(r, newResource); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if err := controllers.Update(r, newResource, oldResource); err != nil {
			return err
		}
	}

	return controllers.Watch(r, newResource)
}

// GetLogger returns the logger from the reconciler.
func (r *{{ .Resource.Kind }}Reconciler) GetLogger() logr.Logger {
	return r.Log
}

// GetClient returns the client from the reconciler.
func (r *{{ .Resource.Kind }}Reconciler) GetClient() client.Client {
	return r.Client
}

// GetScheme returns the scheme from the reconciler.
func (r *{{ .Resource.Kind }}Reconciler) GetScheme() *runtime.Scheme {
	return r.Scheme
}

// GetContext returns the context from the reconciler.
func (r *{{ .Resource.Kind }}Reconciler) GetContext() context.Context {
	return r.Context
}

// GetName returns the name of the reconciler.
func (r *{{ .Resource.Kind }}Reconciler) GetName() string {
	return r.Name
}

// GetComponent returns the component the reconciler is operating against.
func (r *{{ .Resource.Kind }}Reconciler) GetComponent() common.Component {
	return r.Component
}

// GetController returns the controller object associated with the reconciler.
func (r *{{ .Resource.Kind }}Reconciler) GetController() controller.Controller {
	return r.Controller
}

// GetWatches returns the objects which are current being watched by the reconciler.
func (r *{{ .Resource.Kind }}Reconciler) GetWatches() []client.Object {
	return r.Watches
}

// SetWatch appends a watch to the list of currently watched objects.
func (r *{{ .Resource.Kind }}Reconciler) SetWatch(watch client.Object) {
	r.Watches = append(r.Watches, watch)
}

// UpdateStatus updates the status for a component.
func (r *{{ .Resource.Kind }}Reconciler) UpdateStatus() error {
	return r.Status().Update(r.Context, r.Component)
}

{{- if not .IsStandalone }}
// CheckReady will return whether a component is ready.
func (r *{{ .Resource.Kind }}Reconciler) CheckReady() (bool, error) {
	return dependencies.{{ .Resource.Kind }}CheckReady(r)
}

// Mutate will run the mutate phase of a resource.
func (r *{{ .Resource.Kind }}Reconciler) Mutate(
	object *metav1.Object,
) ([]metav1.Object, bool, error) {
	return mutate.{{ .Resource.Kind }}Mutate(r, object)
}

// Wait will run the wait phase of a resource.
func (r *{{ .Resource.Kind }}Reconciler) Wait(
	object *metav1.Object,
) (bool, error) {
	return wait.{{ .Resource.Kind }}Wait(r, object)
}
{{ end }}

func (r *{{ .Resource.Kind }}Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	options := controller.Options{
		RateLimiter: controllers.NewDefaultRateLimiter(5*time.Microsecond, 5*time.Minute),
	}

	baseController, err := ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		WithEventFilter(controllers.ComponentPredicates()).
		For(&{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}{}).
		Build(r)
	if err != nil {
		return err
	}

	r.Controller = baseController

	return nil
}
`
