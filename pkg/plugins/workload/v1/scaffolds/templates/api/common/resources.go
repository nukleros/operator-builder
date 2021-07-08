package common

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Resources{}

// Resources scaffolds the resources for all workloads.
type Resources struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
}

func (f *Resources) SetTemplateDefaults() error {
	f.Path = filepath.Join("apis", "common", "resources.go")

	f.TemplateBody = resourcesTemplate

	return nil
}

const resourcesTemplate = `{{ .Boilerplate }}

package common

// Resource describes a Kubernetes resource managed by the parent object
type Resource struct {
	Created      bool   ` + "`" + `json:"created"` + "`" + `
	Kind         string ` + "`" + `json:"kind"` + "`" + `
	Name         string ` + "`" + `json:"name"` + "`" + `
	Namespace    string ` + "`" + `json:"namespace"` + "`" + `
	LastRevision string ` + "`" + `json:"lastRevision"` + "`" + `
	LastUpdate   string ` + "`" + `json:"lastUpdate"` + "`" + `
}
`
