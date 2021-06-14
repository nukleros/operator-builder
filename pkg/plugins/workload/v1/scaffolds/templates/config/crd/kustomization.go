package crd

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Kustomization{}

// Kustomization scaffolds the CRD kustomization.yaml file
type Kustomization struct {
	machinery.TemplateMixin
	machinery.ResourceMixin

	CRDSampleFilenames []string
}

func (f *Kustomization) SetTemplateDefaults() error {

	f.Path = filepath.Join("config", "crd", "kustomization.yaml")
	f.TemplateBody = kustomizationTemplate
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const kustomizationTemplate = `# This kustomization.yaml is not intended to be run by itself,
# since it depends on service name and namespace that are out of this kustomize package.
# It should be run by config/default
resources:
{{ range .CRDSampleFilenames -}}
- bases/{{ . }}
{{ end }}

patchesStrategicMerge:
# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix.
# patches here are for enabling the conversion webhook for each CRD
#- patches/webhook_in_cloudnativeplatforms.yaml
#+kubebuilder:scaffold:crdkustomizewebhookpatch

# [CERTMANAGER] To enable webhook, uncomment all the sections with [CERTMANAGER] prefix.
# patches here are for enabling the CA injection for each CRD
#- patches/cainjection_in_cloudnativeplatforms.yaml
#+kubebuilder:scaffold:crdkustomizecainjectionpatch

# the following config is for teaching kustomize how to do kustomization for CRDs.
configurations:
- kustomizeconfig.yaml
`
