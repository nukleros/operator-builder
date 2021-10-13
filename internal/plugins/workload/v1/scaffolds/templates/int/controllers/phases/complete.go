// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package phases

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Complete{}

// Complete scaffolds the complete phase methods.
type Complete struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
}

func (f *Complete) SetTemplateDefaults() error {
	f.Path = filepath.Join("internal", "controllers", "phases", "complete.go")

	f.TemplateBody = completeTemplate

	return nil
}

const completeTemplate = `{{ .Boilerplate }}

package phases

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"{{ .Repo }}/apis/common"
)

// CompletePhase.DefaultRequeue executes checking for a parent components readiness status.
func (phase *CompletePhase) DefaultRequeue() ctrl.Result {
	return Requeue()
}

// CompletePhase.Execute executes the completion of a reconciliation loop.
func (phase *CompletePhase) Execute(
	r common.ComponentReconciler,
) (proceedToNextPhase bool, err error) {
	r.GetComponent().SetReadyStatus(true)
	r.GetLogger().V(0).Info("successfully reconciled")

	return true, nil
}
`
