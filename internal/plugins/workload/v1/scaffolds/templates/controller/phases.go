// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package controller

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

var _ machinery.Template = &Controller{}

// Controller scaffolds the workload's controller.
type Phases struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin

	PackageName string
}

func (f *Phases) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"controllers",
		f.Resource.Group,
		fmt.Sprintf("%s_phases.go", utils.ToFileName(f.Resource.Kind)),
	)

	f.TemplateBody = phasesTemplate
	f.IfExistsAction = machinery.SkipFile

	return nil
}

const phasesTemplate = `{{ .Boilerplate }}

package {{ .Resource.Group }}

import (
	"time"

	"github.com/nukleros/operator-builder-tools/pkg/controller/phases"
	ctrl "sigs.k8s.io/controller-runtime"
)

// InitializePhases defines what phases should be run for each event loop. phases are executed
// in the order they are listed.
func (r *{{ .Resource.Kind }}Reconciler) InitializePhases() {
	// Create Phases
	r.Phases.Register(
		"Dependency",
		phases.DependencyPhase,
		phases.CreateEvent,
		phases.WithCustomRequeueResult(ctrl.Result{RequeueAfter: 5 * time.Second }),
	)

	r.Phases.Register(
		"Create-Resources",
		phases.CreateResourcesPhase,
		phases.CreateEvent,
	)

	r.Phases.Register(
		"Check-Ready",
		phases.CheckReadyPhase,
		phases.CreateEvent,
		phases.WithCustomRequeueResult(ctrl.Result{RequeueAfter: 5 * time.Second }),
	)

	r.Phases.Register(
		"Complete",
		phases.CompletePhase,
		phases.CreateEvent,
	)

	// Update Phases
	r.Phases.Register(
		"Dependency",
		phases.DependencyPhase,
		phases.UpdateEvent,
		phases.WithCustomRequeueResult(ctrl.Result{RequeueAfter: 5 * time.Second }),
	)

	r.Phases.Register(
		"Create-Resources",
		phases.CreateResourcesPhase,
		phases.UpdateEvent,
	)

	r.Phases.Register(
		"Check-Ready",
		phases.CheckReadyPhase,
		phases.UpdateEvent,
		phases.WithCustomRequeueResult(ctrl.Result{RequeueAfter: 5 * time.Second }),
	)

	r.Phases.Register(
		"Complete",
		phases.CompletePhase,
		phases.UpdateEvent,
	)

	// Delete Phases
	r.Phases.Register(
		"DeletionComplete",
		phases.DeletionCompletePhase,
		phases.DeleteEvent,
	)
}
`
