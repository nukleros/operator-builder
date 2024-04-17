// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package templates

import (
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/nukleros/operator-builder/internal/utils"
)

var _ machinery.Template = &GoMod{}

// GoMod scaffolds a file that defines the project dependencies.
type GoMod struct {
	machinery.TemplateMixin
	machinery.RepositoryMixin

	GoVersionMinimum     string
	Dependencies         map[string]string
	IndirectDependencies map[string]string
}

// goModDependencyMap pins the versions within the go.mod file so that they
// do not get auto-updated.
//
// See https://github.com/vmware-tanzu-labs/operator-builder/issues/250
func goModDependencyMap() map[string]string {
	return map[string]string{
		"github.com/go-logr/logr":                    "v1.4.1",
		"github.com/nukleros/operator-builder-tools": "v0.4.0",
		"github.com/onsi/ginkgo/v2":                  "v2.17.1",
		"github.com/onsi/gomega":                     "v1.32.0",
		"github.com/spf13/cobra":                     "v1.8.0",
		"github.com/stretchr/testify":                "v1.9.0",
		"gopkg.in/yaml.v2":                           "v2.4.0",
		"k8s.io/api":                                 "v0.29.4",
		"k8s.io/apimachinery":                        "v0.29.4",
		"k8s.io/client-go":                           "v0.29.4",
		"sigs.k8s.io/controller-runtime":             "v0.17.3",
		"sigs.k8s.io/kubebuilder/v3":                 "v3.7.0",
		"sigs.k8s.io/yaml":                           "v1.4.0",
	}
}

// NOTE: there are no indirect dependencies to manage at this time, but we will leave
// it in place when the time comes to use it.
func goModIndirectDependencyMap() map[string]string {
	return map[string]string{}
}

func (f *GoMod) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = "go.mod"
	}

	f.GoVersionMinimum = utils.GeneratedGoVersionMinimum
	f.Dependencies = goModDependencyMap()
	f.IndirectDependencies = goModIndirectDependencyMap()
	f.TemplateBody = goModTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const goModTemplate = `
module {{ .Repo }}

go {{ .GoVersionMinimum }}

require (
	{{ range $package, $version := $.Dependencies }}
	"{{ $package }}" {{ $version }}
	{{- end }}
)

{{ if gt (len $.IndirectDependencies) 0 }}
require (
	{{ range $package, $version := $.IndirectDependencies }}
	"{{ $package }}" {{ $version }}
	{{- end }}
)
{{- end }}
`
