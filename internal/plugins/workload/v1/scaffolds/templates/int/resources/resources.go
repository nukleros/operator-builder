// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package resources

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Resources{}

// The following inherit the ResourceType as they are similar.
type Resources struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin
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

const resourcesTemplate = `{{ .Boilerplate }}

package resources

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"{{ .Repo }}/apis/common"

	rsrcs "github.com/nukleros/operator-builder-tools/pkg/resources"
)

const (
	FieldManager = "reconciler"
)

// Get gets a resource.
func Get(reconciler common.ComponentReconciler, resource client.Object) (metav1.Object, error) {
	// create a stub object to store the current resource in the cluster so that we do not affect
	// the desired state of the resource object in memory
	resourceStore := &unstructured.Unstructured{}
	resourceStore.SetGroupVersionKind(resource.GetObjectKind().GroupVersionKind())

	if err := reconciler.Get(
		reconciler.GetContext(),
		client.ObjectKeyFromObject(resource),
		resourceStore,
	); err != nil {
		if errors.IsNotFound(err) {
			// return nil here so we can easily determine if a resource was not
			// found without having to worry about its type
			return nil, nil
		} else {
			return nil, err
		}
	}

	return resourceStore, nil
}

// Create creates a resource.
func Create(reconciler common.ComponentReconciler, resource client.Object) error {
	reconciler.GetLogger().V(0).Info(
		fmt.Sprintf("creating resource; kind: [%s], name: [%s], namespace: [%s]",
			resource.GetObjectKind().GroupVersionKind().Kind,
			resource.GetName(),
			resource.GetNamespace(),
		),
	)

	if err := reconciler.Create(
		reconciler.GetContext(),
		resource,
		&client.CreateOptions{FieldManager: FieldManager},
	); err != nil {
		return fmt.Errorf("unable to create resource; %v", err)
	}

	return nil
}

// Update updates a resource.
func Update(reconciler common.ComponentReconciler, newResource, oldResource client.Object) error {
	needsUpdate, err := NeedsUpdate(reconciler, newResource, oldResource)
	if err != nil {
		return err
	}

	if needsUpdate {
		reconciler.GetLogger().V(0).Info(
			fmt.Sprintf(
				"updating resource; kind: [%s], name: [%s], namespace: [%s]",
				oldResource.GetObjectKind().GroupVersionKind().Kind,
				oldResource.GetName(),
				oldResource.GetNamespace(),
			),
		)

		if err := reconciler.Patch(
			reconciler.GetContext(),
			newResource,
			client.Merge,
			&client.PatchOptions{FieldManager: FieldManager},
		); err != nil {
			return fmt.Errorf("unable to update resource; %v", err)
		}
	}

	return nil
}

// ToCommonResource converts a resources.Resource into a common API resource.
func ToCommonResource(resource client.Object) *common.Resource {
	resourceCommon := &common.ResourceCommon{
		Group:     resource.GetObjectKind().GroupVersionKind().Group,
		Version:   resource.GetObjectKind().GroupVersionKind().Version,
		Kind:      resource.GetObjectKind().GroupVersionKind().Kind,
		Name:      resource.GetName(),
		Namespace: resource.GetNamespace(),
	}

	resourceObject := &common.Resource{}
	resourceObject.ResourceCommon = *resourceCommon

	return resourceObject
}

// NeedsUpdate determines if a resource needs to be updated.
func NeedsUpdate(reconciler common.ComponentReconciler, desired, actual client.Object) (bool, error) {
	// check for equality first as this will let us avoid spamming user logs
	// when resources that need to be skipped explicitly (e.g. CRDs) are seen
	// as equal anyway
	equal, err := rsrcs.AreEqual(desired, actual)
	if equal || err != nil {
		return !equal, err
	}

	// always skip custom resource updates as they are sensitive to modification
	// e.g. resources provisioned by the resource definition would not
	// understand the update to a spec
	if desired.GetObjectKind().GroupVersionKind().Kind == "CustomResourceDefinition" {
		message := fmt.Sprintf("skipping update of CustomResourceDefinition "+
			"[%s]", desired.GetName())
		messageVerbose := fmt.Sprintf("if updates to CustomResourceDefinition "+
			"[%s] are desired, consider re-deploying the parent "+
			"resource or generating a new api version with the desired "+
			"changes", desired.GetName())
		reconciler.GetLogger().V(4).Info(message)
		reconciler.GetLogger().V(7).Info(messageVerbose)

		return false, nil
	}

	return true, nil
}

// NamespaceForResourceIsReady will check to see if the Namespace of a metadata.namespace
// field of a resource is ready.
func NamespaceForResourceIsReady(r common.ComponentReconciler, resource client.Object) (bool, error) {
	namespace := &v1.Namespace{}
	namespacedName := types.NamespacedName{
		Name:      resource.GetNamespace(),
		Namespace: "",
	}

	if err := r.Get(r.GetContext(), namespacedName, namespace); err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return rsrcs.IsReady(namespace)
}
`
