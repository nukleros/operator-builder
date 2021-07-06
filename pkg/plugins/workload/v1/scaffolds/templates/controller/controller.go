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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

// {{ .Resource.Kind }}Reconciler reconciles a {{ .Resource.Kind }} object
type {{ .Resource.Kind }}Reconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	Context    context.Context
	Component  *{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}
	{{- if .IsComponent }}
	Collection *{{ .Collection.Spec.APIGroup }}{{ .Collection.Spec.APIVersion }}.{{ .Collection.Spec.APIKind }}
	{{ end }}
}

// +kubebuilder:rbac:groups={{ .Resource.Group }}.{{ .Resource.Domain }},resources={{ .Resource.Plural }},verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups={{ .Resource.Group }}.{{ .Resource.Domain }},resources={{ .Resource.Plural }}/status,verbs=get;update;patch
{{ range .RBACRules -}}
// +kubebuilder:rbac:groups={{ .Group }},resources={{ .Resource }},verbs=get;list;watch;create;update;patch;delete
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

	{{- if .IsComponent }}
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
		proceed, err := phase.Execute(r)
		result, err := phases.HandlePhaseExit(r, phase.(phases.PhaseHandler), proceed, err)

		// return only if we have an error or are told not to proceed
		if err != nil || !proceed {
			log.V(4).Info("not proceeding - requeuing")
			return result, err
		}
	}

	return phases.DefaultReconcileResult(), nil
}

// GetResources will create and return the resources in memory
func (r *{{ .Resource.Kind }}Reconciler) GetResources(parent common.Component) ([]metav1.Object, error) {
{{ if .HasChildResources }}
	var resourceObjects []metav1.Object

	// create resources in memory
	for _, f := range {{ .PackageName }}.CreateFuncs {
		{{- if .IsComponent }}
		resource, err := f(r.Component, r.Collection)
		{{ else }}
		resource, err := f(r.Component)
		{{ end }}
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

// SetRefAndCreateIfNotPresent creates a resource if does not already exist
func (r *{{ .Resource.Kind }}Reconciler) SetRefAndCreateIfNotPresent(
	resource metav1.Object,
) error {
	if err := ctrl.SetControllerReference(r.Component, resource, r.Scheme); err != nil {
		r.GetLogger().V(0).Info("unable to set owner reference on resource")
		return err
	}

	//objectKey, err := client.ObjectKeyFromObject(resource.(runtime.Object))
	objectKey := client.ObjectKeyFromObject(resource.(client.Object))
	//if err != nil {
	//	return err
	//}
	if err := r.Get(r.Context, objectKey, resource.(client.Object)); err != nil {
		if errors.IsNotFound(err) {
			r.GetLogger().V(0).Info("creating resource with name: [" + resource.GetName() + "] of kind: [" + resource.(runtime.Object).GetObjectKind().GroupVersionKind().Kind + "]")
			if err := r.Create(r.Context, resource.(client.Object)); err != nil {
				r.GetLogger().V(0).Info("unable to create resource")
				return err
			}
		} else {
			r.GetLogger().V(0).Info("updating resource with name: [" + resource.GetName() + "] of kind: [" + resource.(runtime.Object).GetObjectKind().GroupVersionKind().Kind + "]")
			if err := r.Update(r.Context, resource.(client.Object)); err != nil {
				r.GetLogger().V(0).Info("unable to update resource")
				return err
			}
		}
	}

	return nil
}

// GetLogger returns the logger from the reconciler
func (r *{{ .Resource.Kind }}Reconciler) GetLogger() logr.Logger {
	return r.Log
}

// GetClient returns the client from the reconciler
func (r *{{ .Resource.Kind }}Reconciler) GetClient() client.Client {
	return r.Client
}

// GetScheme returns the scheme from the reconciler
func (r *{{ .Resource.Kind }}Reconciler) GetScheme() *runtime.Scheme {
	return r.Scheme
}

// GetContext returns the context from the reconciler
func (r *{{ .Resource.Kind }}Reconciler) GetContext() context.Context {
	return r.Context
}

// GetComponent returns the component the reconciler is operating against
func (r *{{ .Resource.Kind }}Reconciler) GetComponent() common.Component {
	return r.Component
}

// UpdateStatus updates the status for a component
func (r *{{ .Resource.Kind }}Reconciler) UpdateStatus(
	ctx context.Context,
	parent common.Component,
) error {
	if err := r.Status().Update(ctx, r.Component); err != nil {
		return err
	}
	return nil
}

{{- if not .IsStandalone }}
// CheckReady will return whether a component is ready
func (r *{{ .Resource.Kind }}Reconciler) CheckReady() (bool, error) {
	return dependencies.{{ .Resource.Kind }}CheckReady(r)
}

// Mutate will run the mutate phase of a resource
func (r *{{ .Resource.Kind }}Reconciler) Mutate(
	object *metav1.Object,
) ([]metav1.Object, bool, error) {
	return mutate.{{ .Resource.Kind }}Mutate(r, object)
}

// Wait will run the wait phase of a resource
func (r *{{ .Resource.Kind }}Reconciler) Wait(
	object *metav1.Object,
) (bool, error) {
	return wait.{{ .Resource.Kind }}Wait(r, object)
}
{{ end }}

func (r *{{ .Resource.Kind }}Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}{}).
		Complete(r)
}
`
