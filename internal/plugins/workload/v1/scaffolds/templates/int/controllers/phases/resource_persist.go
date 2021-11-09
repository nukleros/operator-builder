// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &ResourcePersist{}

// ResourcePersist scaffolds the resource persist phase methods.
type ResourcePersist struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *ResourcePersist) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "phases", "resource_persist.go")

	f.TemplateBody = resourcePersistTemplate

	return nil
}

const resourcePersistTemplate = `{{ .Boilerplate }}

package phases

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"{{ .Repo }}/apis/common"
	"{{ .Repo }}/internal/resources"
)

// PersistResourcePhase.Execute executes persisting resources to the Kubernetes database.
func (phase *PersistResourcePhase) Execute(
	r common.ComponentReconciler,
	resource client.Object,
	resourceCondition common.ResourceCondition,
) (ctrl.Result, bool, error) {
	// persist the resource
	if err := persistResource(
		r,
		resource,
		resourceCondition,
		phase,
	); err != nil {
		return ctrl.Result{}, false, err
	}

	return ctrl.Result{}, true, nil
}

// persistResource persists a single resource to the Kubernetes database.
func persistResource(
	r common.ComponentReconciler,
	resource client.Object,
	condition common.ResourceCondition,
	phase *PersistResourcePhase,
) error {
	// persist resource
	if err := r.CreateOrUpdate(resource); err != nil {
		if IsOptimisticLockError(err) {
			return nil
		} else {
			r.GetLogger().V(0).Info(err.Error())

			return err
		}
	}

	// set attributes related to the persistence of this child resource
	condition.LastResourcePhase = getResourcePhaseName(phase)
	condition.LastModified = time.Now().UTC().String()
	condition.Message = "resource created successfully"
	condition.Created = true

	// update the condition to notify that we have created a child resource
	return updateResourceConditions(r, *resources.ToCommonResource(resource), &condition)
}
`
