// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package samples

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"

	"github.com/nukleros/operator-builder/internal/utils"
	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

var _ machinery.Template = &CRDSample{}

// CRDSample scaffolds a file that defines a sample manifest for the CRD.
type CRDSample struct {
	machinery.TemplateMixin
	machinery.ResourceMixin

	SpecFields      *kinds.APIFields
	IsClusterScoped bool
	RequiredOnly    bool
}

func (f *CRDSample) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"config",
		"samples",
		fmt.Sprintf(
			"%s_%s_%s.yaml",
			f.Resource.Group,
			f.Resource.Version,
			utils.ToFileName(f.Resource.Kind)),
	)

	f.RequiredOnly = false
	f.TemplateBody = SampleTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const SampleTemplate = `apiVersion: {{ .Resource.QualifiedGroup }}/{{ .Resource.Version }}
kind: {{ .Resource.Kind }}
metadata:
  name: {{ lower .Resource.Kind }}-sample
{{- if not .IsClusterScoped }}
  namespace: default
{{- end }}
{{ .SpecFields.GenerateSampleSpec false -}}
`

const SampleTemplateRequiredOnly = `apiVersion: {{ .Resource.QualifiedGroup }}/{{ .Resource.Version }}
kind: {{ .Resource.Kind }}
metadata:
  name: {{ lower .Resource.Kind }}-sample
{{- if not .IsClusterScoped }}
  namespace: default
{{- end }}
{{ .SpecFields.GenerateSampleSpec true -}}
`
