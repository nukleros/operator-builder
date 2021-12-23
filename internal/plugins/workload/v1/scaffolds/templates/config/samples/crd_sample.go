// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package samples

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1"
)

var _ machinery.Template = &CRDSample{}

// CRDSample scaffolds a file that defines a sample manifest for the CRD.
type CRDSample struct {
	machinery.TemplateMixin
	machinery.ResourceMixin

	SpecFields      *workloadv1.APIFields
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
