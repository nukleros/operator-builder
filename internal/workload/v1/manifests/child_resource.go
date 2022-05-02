// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package manifests

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/markers"
)

var (
	ErrChildResourceResourceMarkerInspect = errors.New("error inspecting resource markers for child resource")
	ErrChildResourceResourceMarkerProcess = errors.New("error processing resource markers for child resource")
)

// ChildResource contains attributes for resources created by the custom resource.
// These definitions are inferred from the resource manifests.  They can be thought
// of as the individual resources which are managed by the controller during
// reconciliation and represent all resources which are passed in via the `spec.resources`
// field of the workload configuration.
type ChildResource struct {
	Name          string
	UniqueName    string
	Group         string
	Version       string
	Kind          string
	StaticContent string
	SourceCode    string
	IncludeCode   string
}

// NewChildResource returns a representation of a ChildResource object given an unstructured
// Kubernetes object.
func NewChildResource(object unstructured.Unstructured) *ChildResource {
	return &ChildResource{
		Name:       object.GetName(),
		UniqueName: uniqueName(object),
		Group:      object.GetObjectKind().GroupVersionKind().Group,
		Version:    object.GetObjectKind().GroupVersionKind().Version,
		Kind:       object.GetKind(),
	}
}

//nolint:gocritic // needed to satisfy the stringer interface
func (resource ChildResource) String() string {
	return fmt.Sprintf(
		"{Group: %s, Version: %s, Kind: %s, Name: %s}",
		resource.Group, resource.Version, resource.Kind, resource.Name,
	)
}

func (resource *ChildResource) ProcessResourceMarkers(markerCollection *markers.MarkerCollection) error {
	// obtain the marker results from the child resource input yaml
	_, markerResults, err := markers.InspectForYAML([]byte(resource.StaticContent), markers.ResourceMarkerType)
	if err != nil {
		return fmt.Errorf("%w; %s for child resource %s", err, ErrChildResourceResourceMarkerInspect, resource)
	}

	// ensure we have the expected number of resource markers
	//   - 0: return immediately as resource markers are not required
	//   - 1: continue processing normally
	//   - 2: return an error notifying the user that we only expect 1
	//        resource marker
	if len(markerResults) == 0 {
		return nil
	}

	//nolint: godox // depends on https://github.com/vmware-tanzu-labs/operator-builder/issues/271
	// TODO: we need to ensure only one marker is found and return an error if we find more than one.
	// this becomes difficult as the results are returned as yaml nodes.  for now, we just focus on the
	// first result and all others are ignored but we should notify the user.
	result := markerResults[0]

	// process the marker
	marker, ok := result.Object.(markers.ResourceMarker)
	if !ok {
		return ErrChildResourceResourceMarkerProcess
	}

	if err := marker.Process(markerCollection); err != nil {
		return fmt.Errorf("%w; %s for child resource %s", err, ErrChildResourceResourceMarkerProcess, resource)
	}

	if marker.GetIncludeCode() != "" {
		resource.IncludeCode = marker.GetIncludeCode()
	}

	return nil
}

// CreateFuncName returns the create func name for a child resource.
func (resource *ChildResource) CreateFuncName() string {
	return fmt.Sprintf("Create%s", resource.UniqueName)
}

// InitFuncName returns the init func name for a child resource.
func (resource *ChildResource) InitFuncName() string {
	if strings.EqualFold(resource.Kind, "customresourcedefinition") {
		return resource.CreateFuncName()
	}

	return ""
}

// uniqueName returns the unique name of a resource.  This combines the name, namespace, and kind
// into a name that is unique.
//
// NOTE: because resource includes/excludes are now allowd via resource markers, it is possible
// that these names are no longer unique.  Because of this, we dedeuplicate the function names
// when we generate the function names to avoid collisions.
func uniqueName(object unstructured.Unstructured) string {
	// get the resource name taking into account appropriate yaml tags
	resourceName := strings.ReplaceAll(strings.Title(object.GetName()), "-", "")
	resourceName = strings.ReplaceAll(resourceName, ".", "")
	resourceName = strings.ReplaceAll(resourceName, ":", "")
	resourceName = strings.ReplaceAll(resourceName, "!!Start", "")
	resourceName = strings.ReplaceAll(resourceName, "!!End", "")
	resourceName = strings.ReplaceAll(resourceName, "ParentSpec", "")
	resourceName = strings.ReplaceAll(resourceName, "CollectionSpec", "")
	resourceName = strings.ReplaceAll(resourceName, " ", "")

	// get the namespace name taking into account appropriate yaml tags
	namespaceName := strings.ReplaceAll(strings.Title(object.GetNamespace()), "-", "")
	namespaceName = strings.ReplaceAll(namespaceName, ".", "")
	namespaceName = strings.ReplaceAll(namespaceName, ":", "")
	namespaceName = strings.ReplaceAll(namespaceName, "!!Start", "")
	namespaceName = strings.ReplaceAll(namespaceName, "!!End", "")
	namespaceName = strings.ReplaceAll(namespaceName, "ParentSpec", "")
	namespaceName = strings.ReplaceAll(namespaceName, "CollectionSpec", "")
	namespaceName = strings.ReplaceAll(namespaceName, " ", "")

	resourceName = fmt.Sprintf("%s%s%s", object.GetKind(), namespaceName, resourceName)

	return resourceName
}
