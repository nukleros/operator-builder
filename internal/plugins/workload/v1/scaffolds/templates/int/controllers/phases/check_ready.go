// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &CheckReady{}

// CheckReady scaffolds the check ready phase methods.
type CheckReady struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *CheckReady) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "phases", "check_ready.go")

	f.TemplateBody = checkReadyTemplate

	return nil
}

const checkReadyTemplate = `{{ .Boilerplate }}

package phases

import (
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"{{ .Repo }}/apis/common"
	"{{ .Repo }}/internal/resources"

	rsrcs "github.com/nukleros/operator-builder-tools/pkg/resources"
)

// CheckReadyPhase.DefaultRequeue executes checking for a parent components readiness status.
func (phase *CheckReadyPhase) DefaultRequeue() ctrl.Result {
	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: 5 * time.Second,
	}
}

// CheckReadyPhase.Execute executes checking for a parent components readiness status.
func (phase *CheckReadyPhase) Execute(
	r common.ComponentReconciler,
) (proceedToNextPhase bool, err error) {
	// check to see if known types are ready
	knownReady, err := resourcesAreReady(r)
	if err != nil {
		return false, fmt.Errorf("unable to determine if resources are ready, %w", err)
	}

	// check to see if the custom methods return ready
	customReady, err := r.CheckReady()
	if err != nil {
		return false, fmt.Errorf("unable to determine if resources are ready, %w", err)
	}

	return (knownReady && customReady), nil
}

// resourcesAreReady gets the resources in memory, pulls the current state from the
// clusters and determines if they are in a ready condition.
func resourcesAreReady(r common.ComponentReconciler) (bool, error) {
	// get resources in memory
	desiredResources, err := r.GetResources()
	if err != nil {
		return false, fmt.Errorf("unable to retrieve resources, %w", err)
	}

	// get resources from cluster
	clusterResources := make([]metav1.Object, len(desiredResources))
	
	for i, rsrc := range desiredResources {
		clusterResource, err := resources.Get(r, rsrc)
		if err != nil {
			return false, fmt.Errorf("unable to retrieve resource %s, %w", rsrc.GetName(), err)
		}

		clusterResources[i] = clusterResource
	}

	// check to see if known types are ready
	return rsrcs.AreReady(clusterResources...)
}
`
