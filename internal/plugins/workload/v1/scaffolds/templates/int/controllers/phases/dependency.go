// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Dependencies{}

// Dependencies scaffolds the dependency phase methods.
type Dependencies struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Dependencies) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "phases", "dependency.go")

	f.TemplateBody = dependenciesTemplate

	return nil
}

const dependenciesTemplate = `{{ .Boilerplate }}

package phases

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"{{ .Repo }}/apis/common"
)

// DependencyPhase.DefaultRequeue executes checking for a parent components readiness status.
func (phase *DependencyPhase) DefaultRequeue() ctrl.Result {
	return Requeue()
}

// DependencyPhase.Execute executes a dependency check prior to attempting to create resources.
func (phase *DependencyPhase) Execute(r common.ComponentReconciler) (proceedToNextPhase bool, err error) {
	if !r.GetComponent().GetDependencyStatus() {
		satisfied, err := dependenciesSatisfied(r)
		if err != nil {
			return false, fmt.Errorf("unable to list dependencies, %w", err)
		}

		return satisfied, nil
	}

	return true, nil
}

// dependenciesSatisfied will return whether or not all dependencies are satisfied for a component.
func dependenciesSatisfied(r common.ComponentReconciler) (bool, error) {
	for _, dep := range r.GetComponent().GetDependencies() {
		satisfied, err := dependencySatisfied(r, dep)
		if err != nil || !satisfied {
			return false, err
		}
	}

	return true, nil
}

// dependencySatisfied will return whether or not an individual dependency is satisfied.
func dependencySatisfied(r common.ComponentReconciler, dependency common.Component) (bool, error) {
	// get the dependencies by kind that already exist in cluster
	dependencyList := &unstructured.UnstructuredList{}

	dependencyList.SetGroupVersionKind(dependency.GetComponentGVK())

	if err := r.List(r.GetContext(), dependencyList, &client.ListOptions{}); err != nil {
		return false, fmt.Errorf("unable to list dependencies, %w", err)
	}

	// expect only one item returned, otherwise dependencies are considered unsatisfied
	if len(dependencyList.Items) != 1 {
		return false, nil
	}

	// get the status.created field on the object and return the status and any errors found
	status, found, err := unstructured.NestedBool(dependencyList.Items[0].Object, "status", "created")
	if err != nil {
		return false, fmt.Errorf("unable to retrieve status.created field, %w", err)
	}

	if !found {
		return false, nil
	}

	return status, nil
}
`
