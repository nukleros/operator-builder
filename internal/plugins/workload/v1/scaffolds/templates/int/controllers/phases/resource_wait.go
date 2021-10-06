// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &ResourceWait{}

// ResourceWait scaffolds the resource wait phase methods.
type ResourceWait struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *ResourceWait) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "phases", "resource_wait.go")

	f.TemplateBody = resourceWaitTemplate

	return nil
}

const resourceWaitTemplate = `{{ .Boilerplate }}

package phases

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
	"{{ .Repo }}/internal/resources"
)

// WaitForResourcePhase.Execute executes waiting for a resource to be ready before continuing.
func (phase *WaitForResourcePhase) Execute(
	resource common.ComponentResource,
	resourceCondition common.ResourceCondition,
) (ctrl.Result, bool, error) {
	// TODO: loop through functions instead of repeating logic
	// common wait logic for a resource
	ready, err := commonWait(resource.GetReconciler(), resource)
	if err != nil {
		return ctrl.Result{}, false, err
	}

	// return the result if the object is not ready
	if !ready {
		return Requeue(), false, nil
	}

	// specific wait logic for a resource
	meta := resource.GetObject().(metav1.Object)
	ready, err = resource.GetReconciler().Wait(&meta)
	if err != nil {
		return ctrl.Result{}, false, err
	}

	// return the result if the object is not ready
	if !ready {
		return Requeue(), false, nil
	}

	return ctrl.Result{}, true, nil
}

// TODO: the following allows all controllers to list all namespaces,
// regardless of whether or not the controller manages namespaces.
//
// This will eventually be moved into a validating webhook so that the user
// will get a message outlining their mistake rather than buried in the
// reconciliation loop, causing pain when having to sift through logs to
// determine a problem.
//
// See:
//   - https://github.com/vmware-tanzu-labs/operator-builder/issues/141
//   - https://github.com/vmware-tanzu-labs/operator-builder/issues/162

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=list;watch

// commonWait applies all common waiting functions for known resources.
func commonWait(
	r common.ComponentReconciler,
	resource common.ComponentResource,
) (bool, error) {
	// Namespace
	if resource.GetObject().GetNamespace() != "" {
		return resources.NamespaceForResourceIsReady(resource)
	}

	return true, nil
}
`
