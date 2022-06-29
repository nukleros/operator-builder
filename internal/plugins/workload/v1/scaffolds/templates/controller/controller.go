// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package controller

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/kinds"
)

var _ machinery.Template = &Controller{}

// Controller scaffolds the workload's controller.
type Controller struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	// input fields
	Builder kinds.WorkloadBuilder

	// template fields
	BaseImports     []string
	OtherImports    []string
	InternalImports []string
}

func (f *Controller) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"controllers",
		f.Resource.Group,
		fmt.Sprintf("%s_controller.go", utils.ToFileName(f.Resource.Kind)),
	)

	f.TemplateBody = controllerTemplate
	f.IfExistsAction = machinery.OverwriteFile

	f.setBaseImports()
	f.setOtherImports()
	f.setInternalImports()

	return nil
}

func (f *Controller) setBaseImports() {
	f.BaseImports = []string{`"context"`, `"fmt"`}

	if f.Builder.IsComponent() {
		f.BaseImports = append(f.BaseImports, `"errors"`, `"reflect"`)
	}
}

func (f *Controller) setOtherImports() {
	f.OtherImports = []string{
		`"github.com/go-logr/logr"`,
		`apierrs "k8s.io/apimachinery/pkg/api/errors"`,
		`"k8s.io/client-go/tools/record"`,
		`ctrl "sigs.k8s.io/controller-runtime"`,
		`"sigs.k8s.io/controller-runtime/pkg/client"`,
		`"sigs.k8s.io/controller-runtime/pkg/controller"`,
		`"github.com/nukleros/operator-builder-tools/pkg/controller/phases"`,
		`"github.com/nukleros/operator-builder-tools/pkg/controller/predicates"`,
		`"github.com/nukleros/operator-builder-tools/pkg/controller/workload"`,
	}

	if f.Builder.IsComponent() {
		f.OtherImports = append(f.OtherImports,
			`"github.com/nukleros/operator-builder-tools/pkg/resources"`,
			`"sigs.k8s.io/controller-runtime/pkg/event"`,
			`"sigs.k8s.io/controller-runtime/pkg/handler"`,
			`"sigs.k8s.io/controller-runtime/pkg/predicate"`,
			`"sigs.k8s.io/controller-runtime/pkg/reconcile"`,
			`"sigs.k8s.io/controller-runtime/pkg/source"`,
			`"k8s.io/apimachinery/pkg/types"`,
		)
	}
}

func (f *Controller) setInternalImports() {
	f.InternalImports = []string{
		fmt.Sprintf(`"%s/internal/dependencies"`, f.Repo),
		fmt.Sprintf(`"%s/internal/mutate"`, f.Repo),
		fmt.Sprintf(`%s %q`, f.Resource.ImportAlias(), f.Resource.Path),
	}

	if f.Builder.IsComponent() {
		f.InternalImports = append(f.InternalImports, f.getAPITypesPath(f.Builder.GetCollection()))
	}

	if f.Builder.HasChildResources() {
		f.InternalImports = append(f.InternalImports,
			fmt.Sprintf(`"%s/%s"`,
				f.Resource.Path,
				f.Builder.GetPackageName(),
			),
		)
	}
}

func (f *Controller) getAPITypesPath(builder kinds.WorkloadBuilder) string {
	return fmt.Sprintf(`%s%s "%s/apis/%s/%s"`,
		builder.GetAPIGroup(),
		builder.GetAPIVersion(),
		f.Repo,
		builder.GetAPIGroup(),
		builder.GetAPIVersion(),
	)
}

//nolint: lll
const controllerTemplate = `{{ .Boilerplate }}

package {{ .Resource.Group }}

import (
	{{ range .BaseImports -}}
	{{ . }}
	{{ end }}

	{{ range .OtherImports -}}
	{{ . }}
	{{ end }}

	{{ range .InternalImports -}}
	{{ . }}
	{{ end }}
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

{{ range .Builder.GetRBACRules -}}
{{ .ToMarker }}
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
		{{- if .Builder.IsComponent }}
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

	// get the component from the cluster
	if err := r.Get(ctx, request.NamespacedName, component); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "unable to fetch workload")

			return nil, fmt.Errorf("unable to fetch workload, %w", err)
		}

		return nil, err
	}

	// create the workload request
	workloadRequest := &workload.Request{
		Context:  ctx,
		Workload: component,
		Log:      log,
	}

	{{ if .Builder.IsComponent }}
	// store the collection and return any resulting error
	return workloadRequest, r.SetCollection(component, workloadRequest)
	{{- else }}
	return workloadRequest, nil
	{{- end }}
}

{{- if .Builder.IsComponent }}
// SetCollection sets the collection for a particular workload request.
func (r *{{ .Resource.Kind }}Reconciler) SetCollection(component *{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}, req *workload.Request) error {
	collection, err := r.GetCollection(component, req)
	if err != nil || collection == nil {
		return fmt.Errorf("unable to set collection, %w", err)
	}

	req.Collection = collection

	return r.EnqueueRequestOnCollectionChange(req)
}

// GetCollection gets a collection for a component given a list.
func (r *{{ .Resource.Kind }}Reconciler) GetCollection(
	component *{{ .Resource.ImportAlias }}.{{ .Resource.Kind }},
	req *workload.Request,
) (*{{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }}.{{ .Builder.GetCollection.Spec.API.Kind }}, error) {
	var collectionList {{ .Builder.GetCollection.Spec.API.Group }}{{ .Builder.GetCollection.Spec.API.Version }}.{{ .Builder.GetCollection.Spec.API.Kind }}List

	if err := r.List(req.Context, &collectionList); err != nil {
		return nil, fmt.Errorf("unable to list collection {{ .Builder.GetCollection.Spec.API.Kind }}, %w", err)
	}

	// determine if we have requested a specific collection
	name, namespace := component.Spec.Collection.Name, component.Spec.Collection.Namespace

	var collectionRef {{ .Resource.ImportAlias }}.{{ .Resource.Kind }}CollectionSpec

	hasSpecificCollection := component.Spec.Collection != collectionRef && component.Spec.Collection.Name != ""

	// if a specific collection has not been requested, we ensure only one exists
	if !hasSpecificCollection {
		if len(collectionList.Items) != 1 {
			return nil, fmt.Errorf("expected only 1 {{ .Builder.GetCollection.Spec.API.Kind }} collection, found %v", len(collectionList.Items))
		}

		return &collectionList.Items[0], nil
	}

	// find the collection that was requested and return it
	for _, collection := range collectionList.Items {
		if collection.Name == name && collection.Namespace == namespace {
			return &collection, nil
		}
	}

	return nil, workload.ErrCollectionNotFound
}

// EnqueueRequestOnCollectionChange enqueues a reconcile request when an associated collection object changes.
func (r *{{ .Resource.Kind }}Reconciler) EnqueueRequestOnCollectionChange(req *workload.Request) error {
	if len(r.Watches) > 0 {
		for _, watched := range r.Watches {
			if reflect.DeepEqual(
				req.Collection.GetObjectKind().GroupVersionKind(),
				watched.GetObjectKind().GroupVersionKind(),
			) {
				return nil
			}
		}
	}

	// create a function which maps this specific reconcile request
	mapFn := func(collection client.Object) []reconcile.Request {
		return []reconcile.Request{
			{
				NamespacedName: types.NamespacedName{
					Name:      req.Workload.GetName(),
					Namespace: req.Workload.GetNamespace(),
				},
			},
		}
	}

	// watch the collection and use our map function to enqueue the request
	if err := r.Controller.Watch(
		&source.Kind{Type: req.Collection},
		handler.EnqueueRequestsFromMapFunc(mapFn),
		predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				if !resources.EqualNamespaceName(e.ObjectNew, req.Collection) {
					return false
				}

				return e.ObjectNew != e.ObjectOld
			},
			CreateFunc: func(e event.CreateEvent) bool {
				return false
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return false
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
		},
	); err != nil {
		return err
	}

	r.Watches = append(r.Watches, req.Collection)

	return nil
}
{{- end }}

// GetResources resources runs the methods to properly construct the resources in memory.
func (r *{{ .Resource.Kind }}Reconciler) GetResources(req *workload.Request) ([]client.Object, error) {
	{{- if .Builder.HasChildResources }}
	resourceObjects := []client.Object{}

	component, {{ if .Builder.IsComponent }}collection,{{ end }} err := {{ .Builder.GetPackageName }}.ConvertWorkload(req.Workload{{ if .Builder.IsComponent }}, req.Collection{{ end }})
	if err != nil {
		return nil, err
	}

	// create resources in memory
	resources, err := {{ .Builder.GetPackageName }}.Generate(*component{{ if .Builder.IsComponent }}, *collection{{ end }})
	if err != nil {
		return nil, err
	}

	// run through the mutation functions to mutate the resources
	for _, resource := range resources {
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
