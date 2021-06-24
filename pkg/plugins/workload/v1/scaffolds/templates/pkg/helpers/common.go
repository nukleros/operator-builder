package helpers

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Common{}

// Common scaffolds the helper functions
type Common struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin
	machinery.DomainMixin
}

func (f *Common) SetTemplateDefaults() error {
	f.Path = filepath.Join("pkg", "helpers", "common.go")
	f.TemplateBody = commonTemplate

	return nil
}

const commonTemplate = `{{ .Boilerplate }}

package helpers

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	common "{{ .Repo }}/apis/common"
)

const (
	Domain               = "{{ .Domain }}"
	CollectionAPIGroup   = "{{ .Resource.Group }}"
	CollectionAPIVersion = "{{ .Resource.Version }}"
	CollectionAPIKind    = "{{ .Resource.Kind }}"
)

// SkipResourceCreation skips the resource creation during the mutate phase
func SkipResourceCreation(
	err error,
) ([]metav1.Object, bool, error) {
	return []metav1.Object{}, true, err
}

// GetProperObject gets a proper object type given an existing source metav1.object
func GetProperObject(destination metav1.Object, source metav1.Object) error {
	// convert the source object to an unstructured type
	unstructuredObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(source)
	if err != nil {
		return err
	}

	// return the outcome of converting the unstructured type to its proper type
	return runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObject, destination)
}

// getValueFromCollection gets a specific value from the {{ .Resource.Kind }} resource
func getValueFromCollection(reconciler common.ComponentReconciler, path ...string) (string, error) {
	// retrieve a list of platform configs
	collectionConfigs, err := GetCollectionConfigs(reconciler)
	if err != nil {
		return "", err
	}

	if len(collectionConfigs.Items) > 1 {
		return "", errors.New("Must have only one collection resource present in the cluster")
	}

	// get the first platform config
	collectionConfig := collectionConfigs.Items[0]

	// get the value from the platform config
	collectionConfigValue, found, err := unstructured.NestedString(collectionConfig.Object, path...)
	if !found || err != nil {
		return "", fmt.Errorf("unable to get path %s from platform configuration; %v\n", path, err)
	}

	return collectionConfigValue, nil
}

// GetCollectionConfigs returns all of the collection resources from the cluster
func GetCollectionConfigs(
	r common.ComponentReconciler,
) (unstructured.UnstructuredList, error) {
	collectionConfigs := unstructured.UnstructuredList{}
	collectionConfigGVK := schema.GroupVersionKind{
		Group:   fmt.Sprintf("%s.%s", CollectionAPIGroup, Domain),
		Version: CollectionAPIVersion,
		Kind:    CollectionAPIKind,
	}

	// get a list of configurations from the cluster
	collectionConfigs.SetGroupVersionKind(collectionConfigGVK)
	if err := r.List(r.GetContext(), &collectionConfigs, &client.ListOptions{}); err != nil {
		return collectionConfigs, err
	}

	return collectionConfigs, nil
}

// GetCollectionName returns the name of the platform from the {{ .Resource.Kind }} resource
func GetCollectionName(r common.ComponentReconciler) (string, error) {
	return getValueFromCollection(r, "metadata", "name")
}
`
