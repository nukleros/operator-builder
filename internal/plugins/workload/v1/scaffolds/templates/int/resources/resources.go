package resources

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &ResourceType{}

// ResourceType scaffolds the workload's resources package.
type ResourceType struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin
}

// The following inherit the ResourceType as they are similar.
type Resources struct{ ResourceType }
type NamespaceType struct{ ResourceType }
type CustomResourceDefinitionType struct{ ResourceType }
type SecretType struct{ ResourceType }
type ConfigMapType struct{ ResourceType }
type DeploymentType struct{ ResourceType }
type DaemonSetType struct{ ResourceType }
type StatefulSetType struct{ ResourceType }
type JobType struct{ ResourceType }
type ServiceType struct{ ResourceType }

func (f *ResourceType) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"resources",
		"types.go",
	)

	f.TemplateBody = resourceTypesTemplate

	return nil
}

func (f *Resources) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"resources",
		"resources.go",
	)

	f.TemplateBody = resourcesTemplate

	return nil
}

func (f *NamespaceType) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"resources",
		"namespace.go",
	)

	f.TemplateBody = namespaceTemplate

	return nil
}

func (f *ConfigMapType) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"resources",
		"configmap.go",
	)

	f.TemplateBody = configMapTemplate

	return nil
}

func (f *CustomResourceDefinitionType) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"resources",
		"crd.go",
	)

	f.TemplateBody = crdTemplate

	return nil
}

func (f *DaemonSetType) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"resources",
		"daemonset.go",
	)

	f.TemplateBody = daemonSetTemplate

	return nil
}

func (f *DeploymentType) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"resources",
		"deployment.go",
	)

	f.TemplateBody = deploymentTemplate

	return nil
}

func (f *JobType) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"resources",
		"job.go",
	)

	f.TemplateBody = jobTemplate

	return nil
}

func (f *SecretType) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"resources",
		"secret.go",
	)

	f.TemplateBody = secretTemplate

	return nil
}

func (f *ServiceType) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"resources",
		"service.go",
	)

	f.TemplateBody = serviceTemplate

	return nil
}

func (f *StatefulSetType) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"internal",
		"resources",
		"statefulset.go",
	)

	f.TemplateBody = statefulSetTemplate

	return nil
}

const resourceTypesTemplate = `{{ .Boilerplate }}

package resources

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"{{ .Repo }}/apis/common"
)

// Resource represents a resource as managed during the reconciliation process.
type Resource struct {
	common.ResourceCommon

	Object     client.Object
	Reconciler common.ComponentReconciler
}
`

const resourcesTemplate = `{{ .Boilerplate }}

package resources

import (
	"fmt"

	"github.com/imdario/mergo"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"{{ .Repo }}/apis/common"
)

const (
	FieldManager = "reconciler"
)

// Create creates a resource.
func (resource *Resource) Create() error {
	resource.Reconciler.GetLogger().V(0).Info(fmt.Sprintf("creating resource; kind: [%s], name: [%s], namespace: [%s]",
		resource.Kind, resource.Name, resource.Namespace))

	if err := resource.Reconciler.Create(
		resource.Reconciler.GetContext(),
		resource.Object,
		&client.CreateOptions{FieldManager: FieldManager},
	); err != nil {
		return fmt.Errorf("unable to create resource; %v", err)
	}

	return nil
}

// Update updates a resource.
func (resource *Resource) Update(oldResource *Resource) error {
	equal, err := AreEqual(*resource, *oldResource)
	if err != nil {
		return err
	}

	if !equal {
		resource.Reconciler.GetLogger().V(0).Info(fmt.Sprintf("updating resource; kind: [%s], name: [%s], namespace: [%s]",
			resource.Kind, resource.Name, resource.Namespace))

		if err := resource.Reconciler.Patch(
			resource.Reconciler.GetContext(),
			resource.Object,
			client.Merge,
			&client.PatchOptions{FieldManager: FieldManager},
		); err != nil {
			return fmt.Errorf("unable to update resource; %v", err)
		}
	}

	return nil
}

// NewResourceFromClient returns a new resource given a client object.  It optionally will take in
// a reconciler and set it.
func NewResourceFromClient(resource client.Object, reconciler ...common.ComponentReconciler) *Resource {
	newResource := &Resource{
		Object: resource,
	}

	// set the inherited fields
	newResource.Group = resource.GetObjectKind().GroupVersionKind().Group
	newResource.Version = resource.GetObjectKind().GroupVersionKind().Version
	newResource.Kind = resource.GetObjectKind().GroupVersionKind().Kind
	newResource.Name = resource.GetName()
	newResource.Namespace = resource.GetNamespace()

	if len(reconciler) > 0 {
		newResource.Reconciler = reconciler[0]
	}

	return newResource
}

// ToUnstructured returns an unstructured representation of a Resource.
func (resource *Resource) ToUnstructured() (*unstructured.Unstructured, error) {
	innerObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&resource.Object)
	if err != nil {
		return nil, err
	}

	return &unstructured.Unstructured{Object: innerObject}, nil
}

// ToCommonResource converts a resources.Resource into a common API resource.
func (resource *Resource) ToCommonResource() *common.Resource {
	commonResource := &common.Resource{}

	// set the inherited fields
	commonResource.Group = resource.Group
	commonResource.Version = resource.Version
	commonResource.Kind = resource.Kind
	commonResource.Name = resource.Name
	commonResource.Namespace = resource.Namespace

	return commonResource
}

// IsReady returns whether a specific known resource is ready.  Always returns true for unknown resources
// so that dependency checks will not fail and reconciliation of resources can happen with errors rather
// than stopping entirely.
func (resource *Resource) IsReady() (bool, error) {
	switch resource.Kind {
	case NamespaceKind:
		return NamespaceIsReady(resource)
	case CustomResourceDefinitionKind:
		return CustomResourceDefinitionIsReady(resource)
	case SecretKind:
		return SecretIsReady(resource)
	case ConfigMapKind:
		return ConfigMapIsReady(resource)
	case DeploymentKind:
		return DeploymentIsReady(resource)
	case DaemonSetKind:
		return DaemonSetIsReady(resource)
	case StatefulSetKind:
		return StatefulSetIsReady(resource)
	case JobKind:
		return JobIsReady(resource)
	case ServiceKind:
		return ServiceIsReady(resource)
	}

	return true, nil
}

// AreReady returns whether resources are ready.  All resources must be ready in order
// to satisfy the requirement that resources are ready.
func AreReady(resources ...common.ComponentResource) (bool, error) {
	for _, resource := range resources {
		ready, err := resource.IsReady()
		if !ready || err != nil {
			return false, err
		}
	}

	return true, nil
}

// AreEqual determines if two resources are equal.
func AreEqual(desired, actual Resource) (bool, error) {
	mergedResource, err := actual.ToUnstructured()
	if err != nil {
		return false, err
	}

	actualResource, err := actual.ToUnstructured()
	if err != nil {
		return false, err
	}

	desiredResource, err := desired.ToUnstructured()
	if err != nil {
		return false, err
	}

	// ensure that resource versions and observed generation do not interfere
	// with calculating equality
	desiredResource.SetResourceVersion(actualResource.GetResourceVersion())
	desiredResource.SetGeneration(actualResource.GetGeneration())

	// ensure that a current cluster-scoped resource is not evaluated against
	// a manifest which may include a namespace
	if actualResource.GetNamespace() == "" {
		desiredResource.SetNamespace(actualResource.GetNamespace())
	}

	// merge the overrides from the desired resource into the actual resource
	mergo.Merge(
		&mergedResource.Object,
		desiredResource.Object,
		mergo.WithOverride,
		mergo.WithSliceDeepCopy,
	)

	// calculate the actual differences
	diffOptions := []patch.CalculateOption{
		reconciler.IgnoreManagedFields(),
		patch.IgnoreStatusFields(),
		patch.IgnoreVolumeClaimTemplateTypeMetaAndStatus(),
		patch.IgnorePDBSelector(),
	}

	diffResults, err := patch.DefaultPatchMaker.Calculate(
		actualResource,
		mergedResource,
		diffOptions...,
	)
	if err != nil {
		return false, err
	}

	return diffResults.IsEmpty(), nil
}

// EqualNamespaceName will compare the namespace and name of two resource objects for equality.
func (resource *Resource) EqualNamespaceName(compared common.ComponentResource) bool {
	comparedResource := compared.(*Resource)
	return (resource.Name == comparedResource.Name) && (resource.Namespace == comparedResource.Namespace)
}

// EqualGVK will compare the GVK of two resource objects for equality.
func (resource *Resource) EqualGVK(compared common.ComponentResource) bool {
	comparedResource := compared.(*Resource)
	return resource.Group == comparedResource.Group &&
		resource.Version == comparedResource.Version &&
		resource.Kind == comparedResource.Kind
}

// GetObject returns the Object field of a Resource.
func (resource *Resource) GetObject() client.Object {
	return resource.Object
}

// GetReconciler returns the Reconciler field of a Resource.
func (resource *Resource) GetReconciler() common.ComponentReconciler {
	return resource.Reconciler
}

// GetGroup returns the Group field of a Resource.
func (resource *Resource) GetGroup() string {
	return resource.Group
}

// GetVersion returns the Version field of a Resource.
func (resource *Resource) GetVersion() string {
	return resource.Version
}

// GetKind returns the Kind field of a Resource.
func (resource *Resource) GetKind() string {
	return resource.Kind
}

// GetName returns the Name field of a Resource.
func (resource *Resource) GetName() string {
	return resource.Name
}

// GetNamespace returns the Name field of a Resource.
func (resource *Resource) GetNamespace() string {
	return resource.Namespace
}

// getObject returns an object based on an input object, and a destination object.
// TODO: move to controller utils as this is not specific to resources.
func getObject(source common.ComponentResource, destination client.Object, allowMissing bool) error {
	namespacedName := types.NamespacedName{
		Name:      source.GetName(),
		Namespace: source.GetNamespace(),
	}
	if err := source.GetReconciler().Get(source.GetReconciler().GetContext(), namespacedName, destination); err != nil {
		if allowMissing {
			if errors.IsNotFound(err) {
				return nil
			}
		} else {
			return err
		}
	}

	return nil
}
`

const namespaceTemplate = `{{ .Boilerplate }}

package resources

import (
	v1 "k8s.io/api/core/v1"

	"{{ .Repo }}/apis/common"
)

const (
	NamespaceKind = "Namespace"
)

// NamespaceIsReady defines the criteria for a namespace to be condsidered ready.
func NamespaceIsReady(resource common.ComponentResource) (bool, error) {
	var namespace v1.Namespace
	if err := getObject(resource, &namespace, true); err != nil {
		return false, err
	}

	// if we have a name that is empty, we know we did not find the object
	if namespace.Name == "" {
		return false, nil
	}

	// if the namespace is terminating, it is not considered ready
	if namespace.Status.Phase == v1.NamespaceTerminating {
		return false, nil
	}

	// finally, rely on the active field to determine if this namespace is ready
	return namespace.Status.Phase == v1.NamespaceActive, nil
}

// NamespaceForResourceIsReady checks to see if the namespace of a resource is ready.
func NamespaceForResourceIsReady(resource common.ComponentResource) (bool, error) {
	// create a stub namespace resource to pass to the NamespaceIsReady method
	namespace := &Resource{
		Reconciler: resource.GetReconciler(),
	}

	// insert the inherited fields
	namespace.Name = resource.GetNamespace()
	namespace.Group = ""
	namespace.Version = "v1"
	namespace.Kind = NamespaceKind

	return NamespaceIsReady(namespace)
}
`

const configMapTemplate = `{{ .Boilerplate }}

package resources

import (
	v1 "k8s.io/api/core/v1"

	"{{ .Repo }}/apis/common"
)

const (
	ConfigMapKind = "ConfigMap"
)

// ConfigMapIsReady performs the logic to determine if a secret is ready.
func ConfigMapIsReady(resource common.ComponentResource, expectedKeys ...string) (bool, error) {
	var configMap v1.ConfigMap
	if err := getObject(resource, &configMap, true); err != nil {
		return false, err
	}

	// if we have a name that is empty, we know we did not find the object
	if configMap.Name == "" {
		return false, nil
	}

	// check the status for a ready ca keypair secret
	for _, key := range expectedKeys {
		if string(configMap.Data[key]) == "" {
			return false, nil
		}
	}

	return true, nil
}
`

const crdTemplate = `{{ .Boilerplate }}

package resources

import (
	"k8s.io/apimachinery/pkg/api/errors"

	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"{{ .Repo }}/apis/common"
)

const (
	CustomResourceDefinitionKind = "CustomResourceDefinition"
)

// CustomResourceDefinitionIsReady performs the logic to determine if a custom resource definition is ready.
func CustomResourceDefinitionIsReady(resource common.ComponentResource) (bool, error) {
	var crd extensionsv1.CustomResourceDefinition
	if err := getObject(resource, &crd, false); err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
	}

	return true, nil
}
`

const daemonSetTemplate = `{{ .Boilerplate }}

package resources

import (
	appsv1 "k8s.io/api/apps/v1"

	"{{ .Repo }}/apis/common"
)

const (
	DaemonSetKind = "DaemonSet"
)

// DaemonSetIsReady checks to see if a daemonset is ready.
func DaemonSetIsReady(resource common.ComponentResource) (bool, error) {
	var daemonSet appsv1.DaemonSet
	if err := getObject(resource, &daemonSet, true); err != nil {
		return false, err
	}

	// if we have a name that is empty, we know we did not find the object
	if daemonSet.Name == "" {
		return false, nil
	}

	// ensure the desired number is scheduled and ready
	if daemonSet.Status.DesiredNumberScheduled == daemonSet.Status.NumberReady {
		if daemonSet.Status.NumberReady > 0 && daemonSet.Status.NumberUnavailable < 1 {
			return true, nil
		} else {
			return false, nil
		}
	}

	return false, nil
}
`

const deploymentTemplate = `{{ .Boilerplate }}

package resources

import (
	appsv1 "k8s.io/api/apps/v1"

	"{{ .Repo }}/apis/common"
)

const (
	DeploymentKind = "Deployment"
)

// DeploymentIsReady performs the logic to determine if a deployment is ready.
func DeploymentIsReady(resource common.ComponentResource) (bool, error) {
	var deployment appsv1.Deployment
	if err := getObject(resource, &deployment, true); err != nil {
		return false, err
	}

	// if we have a name that is empty, we know we did not find the object
	if deployment.Name == "" {
		return false, nil
	}

	// rely on observed generation to give us a proper status
	if deployment.Generation != deployment.Status.ObservedGeneration {
		return false, nil
	}

	// check the status for a ready deployment
	if deployment.Status.ReadyReplicas != deployment.Status.Replicas {
		return false, nil
	}

	return true, nil
}
`

const jobTemplate = `{{ .Boilerplate }}

package resources

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"

	"{{ .Repo }}/apis/common"
)

const (
	JobKind = "Job"
)

// JobIsReady checks to see if a job is ready.
func JobIsReady(resource common.ComponentResource) (bool, error) {
	var job batchv1.Job
	if err := getObject(resource, &job, true); err != nil {
		return false, err
	}

	// if we have a name that is empty, we know we did not find the object
	if job.Name == "" {
		return false, nil
	}

	// return immediately if the job is active or has no completion time
	if job.Status.Active == 1 || job.Status.CompletionTime == nil {
		return false, nil
	}

	// ensure the completion is actually successful
	if job.Status.Succeeded != 1 {
		return false, fmt.Errorf("job " + job.GetName() + " was not successful")
	}

	return true, nil
}
`

const secretTemplate = `{{ .Boilerplate }}

package resources

import (
	v1 "k8s.io/api/core/v1"

	"{{ .Repo }}/apis/common"
)

const (
	SecretKind = "Secret"
)

// SecretIsReady performs the logic to determine if a secret is ready.
func SecretIsReady(resource common.ComponentResource, expectedKeys ...string) (bool, error) {
	var secret v1.Secret
	if err := getObject(resource, &secret, true); err != nil {
		return false, err
	}

	// if we have a name that is empty, we know we did not find the object
	if secret.Name == "" {
		return false, nil
	}

	// check the status for a ready secret if we expect certain fields to exist
	for _, key := range expectedKeys {
		if string(secret.Data[key]) == "" {
			return false, nil
		}
	}

	return true, nil
}
`

const serviceTemplate = `{{ .Boilerplate }}

package resources

import (
	corev1 "k8s.io/api/core/v1"

	"{{ .Repo }}/apis/common"
)

const (
	ServiceKind = "Service"
)

// ServiceIsReady checks to see if a job is ready.
func ServiceIsReady(resource common.ComponentResource) (bool, error) {
	var service corev1.Service
	if err := getObject(resource, &service, true); err != nil {
		return false, err
	}

	// if we have a name that is empty, we know we did not find the object
	if service.Name == "" {
		return false, nil
	}

	// return if we have an external service type
	if service.Spec.Type == corev1.ServiceTypeExternalName {
		return true, nil
	}

	// ensure a cluster ip address exists for cluster ip types
	if service.Spec.ClusterIP != corev1.ClusterIPNone && len(service.Spec.ClusterIP) == 0 {
		return false, nil
	}

	// ensure a load balancer ip or hostname is present
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
		if len(service.Status.LoadBalancer.Ingress) == 0 {
			return false, nil
		}
	}

	return true, nil
}
`

const statefulSetTemplate = `{{ .Boilerplate }}

package resources

import (
	appsv1 "k8s.io/api/apps/v1"

	"{{ .Repo }}/apis/common"
)

const (
	StatefulSetKind = "StatefulSet"
)

// StatefulSetIsReady performs the logic to determine if a secret is ready.
func StatefulSetIsReady(resource common.ComponentResource, expectedKeys ...string) (bool, error) {
	var statefulSet appsv1.StatefulSet
	if err := getObject(resource, &statefulSet, true); err != nil {
		return false, err
	}

	// if we have a name that is empty, we know we did not find the object
	if statefulSet.Name == "" {
		return false, nil
	}

	// rely on observed generation to give us a proper status
	if statefulSet.Generation != statefulSet.Status.ObservedGeneration {
		return false, nil
	}

	// check for valid replicas
	replicas := statefulSet.Spec.Replicas
	if replicas == nil {
		return false, nil
	}

	// check to see if replicas have been updated
	var needsUpdate int32
	if statefulSet.Spec.UpdateStrategy.RollingUpdate != nil &&
		statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition != nil &&
		*statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition > 0 {

		needsUpdate -= *statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition
	}
	notUpdated := needsUpdate - statefulSet.Status.UpdatedReplicas
	if notUpdated > 0 {
		return false, nil
	}

	// check to see if replicas are available
	notReady := *replicas - statefulSet.Status.ReadyReplicas
	if notReady > 0 {
		return false, nil
	}

	// check to see if a scale down operation is complete
	notDeleted := statefulSet.Status.Replicas - *replicas
	if notDeleted > 0 {
		return false, nil
	}

	return true, nil
}
`
