package controller

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Common{}

// Common scaffolds controller utilities common to all controllers
type Common struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.RepositoryMixin
	machinery.ResourceMixin
}

func (f *Common) SetTemplateDefaults() error {

	f.Path = filepath.Join("controllers", "common.go")

	f.TemplateBody = controllerCommonTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

var controllerCommonTemplate = `{{ .Boilerplate }}

package controllers

import (
	apierrs "k8s.io/apimachinery/pkg/api/errors"

	apiscommon "{{ .Repo }}/apis/common"
	phases "{{ .Repo }}/controllers/phases"
)

func IgnoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

// CreatePhases defines the phases for create and the order in which they run during the reconcile process
func CreatePhases() []phases.Phase {
	return []phases.Phase{
		//&phases.DependencyPhase{},
		//&phases.PreFlightPhase{},
		&phases.CreateResourcesPhase{},
		//&phases.CheckReadyPhase{},
		//&phases.CompletePhase{},
	}
}

// UpdatePhases defines the phases for update and the order in which they run during the reconcile process
func UpdatePhases() []phases.Phase {
	// we have nothing to do for updating at this point
	return []phases.Phase{}
}

// Phases returns which phases to run given the component
func Phases(component apiscommon.Component) []phases.Phase {
	var phases []phases.Phase
	if !component.GetReadyStatus() {
		phases = CreatePhases()
	} else {
		phases = UpdatePhases()
	}

	return phases
}
`
