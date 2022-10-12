// Copyright 2022 Nukleros
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
		"github.com/go-logr/logr":                    "v1.2.3",
		"github.com/nukleros/operator-builder-tools": "v0.3.0",
		"github.com/onsi/ginkgo":                     "v1.16.5",
		"github.com/onsi/gomega":                     "v1.19.0",
		"github.com/spf13/cobra":                     "v1.4.0",
		"github.com/stretchr/testify":                "v1.7.3",
		"google.golang.org/api":                      "v0.84.0",
		"gopkg.in/yaml.v2":                           "v2.4.0",
		"k8s.io/api":                                 "v0.24.2",
		"k8s.io/apimachinery":                        "v0.24.2",
		"k8s.io/client-go":                           "v0.24.2",
		"sigs.k8s.io/controller-runtime":             "v0.12.1",
		"sigs.k8s.io/kubebuilder/v3":                 "v3.4.1",
		"sigs.k8s.io/yaml":                           "v1.3.0",
	}
}

func goModIndirectDependencyMap() map[string]string {
	return map[string]string{
		"gopkg.in/check.v1": "v1.0.0-20201130134442-10cb98267c6c",
	}
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

require (
	{{ range $package, $version := $.IndirectDependencies }}
	"{{ $package }}" {{ $version }}
	{{- end }}
)
`
