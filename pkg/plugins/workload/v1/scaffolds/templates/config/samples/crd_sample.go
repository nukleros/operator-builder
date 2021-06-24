package samples

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/utils"
	workloadv1 "github.com/vmware-tanzu-labs/operator-builder/pkg/workload/v1"
)

var _ machinery.Template = &CRDSample{}

// CRDSample scaffolds a file that defines a sample manifest for the CRD.
type CRDSample struct {
	machinery.TemplateMixin
	machinery.ResourceMixin

	SpecFields *[]workloadv1.APISpecField
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

	f.TemplateBody = crdSampleTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const crdSampleTemplate = `apiVersion: {{ .Resource.QualifiedGroup }}/{{ .Resource.Version }}
kind: {{ .Resource.Kind }}
metadata:
  name: {{ lower .Resource.Kind }}-sample
spec:
{{- range .SpecFields }}
  {{ .SampleField -}}
{{ end }}
`
