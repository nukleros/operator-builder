// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package controller

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
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
	{{- if .IsComponent }}
	"errors"
	{{- end }}
	"fmt"

	"github.com/go-logr/logr"
	"github.com/nukleros/operator-builder-tools/pkg/controller/phases"
	"github.com/nukleros/operator-builder-tools/pkg/controller/predicates"
	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	{{ if .IsComponent -}}
	{{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }} "{{ .Repo }}/apis/{{ .Collection.Spec.API.Group }}/{{ .Collection.Spec.API.Version }}"
	{{ end }}
	{{- if .HasChildResources -}}
	"{{ .Resource.Path }}/{{ .PackageName }}"
	{{ end -}}
	"{{ .Repo }}/internal/dependencies"
	"{{ .Repo }}/internal/mutate"
)

// {{ .Resource.Kind }}Reconciler reconciles a {{ .Resource.Kind }} object.
type {{ .Resource.Kind }}Reconciler struct {
	client.Client
	Name         string
	Log          logr.Logger
	Controller   controller.Controller
	Events       record.EventRecorder
	FieldManager string
	Watches      []client.Object
	Phases       *phases.Registry
}

func New{{ .Resource.Kind }}Reconciler(mgr ctrl.Manager) *{{ .Resource.Kind }}Reconciler {
	return &{{ .Resource.Kind }}Reconciler{
		Name:         "{{ .Resource.Kind }}",
		Client:       mgr.GetClient(),
		Events:       mgr.GetEventRecorderFor("{{ .Resource.Kind }}-Controller"),
		FieldManager: "{{ .Resource.Kind }}-reconciler",
		Log:          ctrl.Log.WithName("controllers").WithName("{{ .Resource.Group }}").WithName("{{ .Resource.Kind }}"),
		Watches:      []client.Object{},
		Phases:       &phases.Registry{},
	}
}

// +kubebuilder:rbac:groups={{ .Resource.Group }}.{{ .Resource.Domain }},resources={{ .Resource.Plural }},verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups={{ .Resource.Group }}.{{ .Resource.Domain }},resources={{ .Resource.Plural }}/status,verbs=get;update;patch
{{ range .RBACRules -}}
// +kubebuilder:rbac:groups={{ .Group }},resources={{ .Resource }},verbs={{ .VerbString }}
{{ end }}

// Until Webhooks are implemented we need to list and watch namespaces to ensure
// they are available before deploying resources,
// See:
//   - https://github.com/vmware-tanzu-labs/operator-builder/issues/141
//   - https://github.com/vmware-tanzu-labs/operator-builder/issues/162

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *{{ .Resource.Kind }}Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	req, err := r.NewRequest(ctx, request)
	if err != nil {
		{{- if .IsComponent }}
		if errors.Is(err, workload.ErrCollectionNotFound) {
			return ctrl.Result{Requeue: true}, nil
		}
		{{- end }}

		if !apierrs.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		
		return ctrl.Result{}, nil
	}

	if err := phases.RegisterDeleteHooks(r, req); err != nil {
		return ctrl.Result{}, err
	}

	// execute the phases
	return r.Phases.HandleExecution(r, req)
}

func (r *{{ .Resource.Kind }}Reconciler) NewRequest(ctx context.Context, request ctrl.Request) (*workload.Request, error) {
	component := &{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}{}

	log := r.Log.WithValues(
		"kind", component.GetWorkloadGVK().Kind,
		"name", request.Name,
		"namespace", request.Namespace,
	)

	// get and store the component
	if err := r.Get(ctx, request.NamespacedName, component); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "unable to fetch workload")

			return nil, fmt.Errorf("unable to fetch workload, %w", err)
		}

		return nil, err
	}

	{{ if .IsComponent -}}
	// get and store the collection
	var collectionList {{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }}.{{ .Collection.Spec.API.Kind }}List
	
	var collection *{{ .Collection.Spec.API.Group }}{{ .Collection.Spec.API.Version }}.{{ .Collection.Spec.API.Kind }}

	if err := r.List(ctx, &collectionList); err != nil {
		return nil, fmt.Errorf("unable to list collection {{ .Collection.Spec.API.Kind }}, %w", err)
	}

	switch len(collectionList.Items) {
	case 0:
		if component.GetDeletionTimestamp().IsZero() {
			log.Info("no collections available; initiating controller requeue")

			return nil, workload.ErrCollectionNotFound
		}
	case 1:
		collection = &collectionList.Items[0]
	default:
		log.Info("multiple collections found; expected 1; cannot proceed")

		return nil, nil
	}
	{{- end }}

	return &workload.Request{
		Context:    ctx,
		Workload:   component,
		{{- if .IsComponent }}
		Collection: collection,
		{{- end }}
		Log:        log,
	}, nil
}

// GetResources resources runs the methods to properly construct the resources in memory.
func (r *{{ .Resource.Kind }}Reconciler) GetResources(req *workload.Request) ([]client.Object, error) {
	{{- if .HasChildResources }}
	resourceObjects := []client.Object{}

	component, {{ if .IsComponent }}collection,{{ end }} err := {{ .PackageName }}.ConvertWorkload(req.Workload{{ if .IsComponent }}, req.Collection{{ end }})
	if err != nil {
		return nil, err
	}

	// create resources in memory
	for _, f := range {{ .PackageName }}.CreateFuncs {
		resource, err := f(component{{ if .IsComponent }}, collection{{ end }})
		if err != nil {
			return nil, err
		}

		// run through the mutation functions to mutate the resources
		mutatedResources, skip, err := r.Mutate(req, resource)
		if err != nil {
			return []client.Object{}, err
		}

		if skip {
			continue
		}

		resourceObjects = append(resourceObjects, mutatedResources...)
	}

	return resourceObjects, nil
{{- else -}}
	return []client.Object{}, nil
{{ end -}}
}

// GetEventRecorder returns the event recorder for writing kubernetes events.
func (r *{{ .Resource.Kind }}Reconciler) GetEventRecorder() record.EventRecorder {
	return r.Events
}

// GetFieldManager returns the name of the field manager for the controller.
func (r *{{ .Resource.Kind }}Reconciler) GetFieldManager() string {
	return r.FieldManager
}

// GetLogger returns the logger from the reconciler.
func (r *{{ .Resource.Kind }}Reconciler) GetLogger() logr.Logger {
	return r.Log
}

// GetName returns the name of the reconciler.
func (r *{{ .Resource.Kind }}Reconciler) GetName() string {
	return r.Name
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

// CheckReady will return whether a component is ready.
func (r *{{ .Resource.Kind }}Reconciler) CheckReady(req *workload.Request) (bool, error) {
	return dependencies.{{ .Resource.Kind }}CheckReady(r, req)
}

// Mutate will run the mutate function for the workload.
func (r *{{ .Resource.Kind }}Reconciler) Mutate(
	req *workload.Request,
	object client.Object,
) ([]client.Object, bool, error) {
	return mutate.{{ .Resource.Kind }}Mutate(r, req, object)
}

func (r *{{ .Resource.Kind }}Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.InitializePhases()

	baseController, err := ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(predicates.WorkloadPredicates()).
		For(&{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}{}).
		Build(r)
	if err != nil {
		return fmt.Errorf("unable to setup controller, %w", err)
	}

	r.Controller = baseController

	return nil
}
`
