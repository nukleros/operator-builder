// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package templates

import (
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &GoMod{}

// GoMod scaffolds a file that defines the project dependencies.
type GoMod struct {
	machinery.TemplateMixin
	machinery.RepositoryMixin

	Dependencies map[string]string
}

// goModDependencyMap pins the versions within the go.mod file so that they
// do not get auto-updated.
//
// See https://github.com/vmware-tanzu-labs/operator-builder/issues/250
func goModDependencyMap() map[string]string {
	return map[string]string{
		"github.com/go-logr/logr":                    "v0.4.0",
		"github.com/nukleros/operator-builder-tools": "v0.2.0",
		"github.com/onsi/ginkgo":                     "v1.16.4",
		"github.com/onsi/gomega":                     "v1.15.0",
		"github.com/spf13/cobra":                     "v1.2.1",
		"github.com/stretchr/testify":                "v1.7.0",
		"gopkg.in/yaml.v2":                           "v2.4.0",
		"k8s.io/api":                                 "v0.22.2",
		"k8s.io/apimachinery":                        "v0.22.2",
		"k8s.io/client-go":                           "v0.22.2",
		"sigs.k8s.io/controller-runtime":             "v0.10.2",
		"sigs.k8s.io/kubebuilder/v3":                 "v3.2.0",
		"sigs.k8s.io/yaml":                           "v1.2.0",
	}
}

func (f *GoMod) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = "go.mod"
	}

	f.Dependencies = goModDependencyMap()
	f.TemplateBody = goModTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const goModTemplate = `
module {{ .Repo }}

go 1.15

require (
	{{ range $k, $v := $.Dependencies }}
	"{{ $k }}" {{ $v }}
	{{ end -}}
)
`
