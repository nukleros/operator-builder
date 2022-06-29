// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package crd

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var (
	_ machinery.Template = &Kustomization{}
	_ machinery.Inserter = &Kustomization{}
)

// Kustomization scaffolds a file that defines the kustomization scheme for the crd folder.
type Kustomization struct {
	machinery.TemplateMixin
	machinery.ResourceMixin
}

// SetTemplateDefaults implements file.Template.
func (f *Kustomization) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("config", "crd", "kustomization.yaml")
	}

	f.Path = f.Resource.Replacer().Replace(f.Path)

	f.TemplateBody = fmt.Sprintf(kustomizationTemplate,
		machinery.NewMarkerFor(f.Path, resourceMarker),
		machinery.NewMarkerFor(f.Path, webhookPatchMarker),
		machinery.NewMarkerFor(f.Path, caInjectionPatchMarker),
	)

	return nil
}

const (
	resourceMarker         = "crdkustomizeresource"
	webhookPatchMarker     = "crdkustomizewebhookpatch"
	caInjectionPatchMarker = "crdkustomizecainjectionpatch"
)

// GetMarkers implements file.Inserter.
func (f *Kustomization) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.Path, resourceMarker),
		machinery.NewMarkerFor(f.Path, webhookPatchMarker),
		machinery.NewMarkerFor(f.Path, caInjectionPatchMarker),
	}
}

const (
	resourceCodeFragment = `- bases/%s_%s.yaml
`
	webhookPatchCodeFragment = `#- patches/webhook_in_%s.yaml
`
	caInjectionPatchCodeFragment = `#- patches/cainjection_in_%s.yaml
`
)

// GetCodeFragments implements file.Inserter.
func (f *Kustomization) GetCodeFragments() machinery.CodeFragmentsMap {
	const codeFragmentsLen = 3
	fragments := make(machinery.CodeFragmentsMap, codeFragmentsLen)

	// Generate resource code fragments
	res := make([]string, 0)
	res = append(res, fmt.Sprintf(resourceCodeFragment, f.Resource.QualifiedGroup(), f.Resource.Plural))

	// Generate resource code fragments
	webhookPatch := make([]string, 0)
	webhookPatch = append(webhookPatch, fmt.Sprintf(webhookPatchCodeFragment, f.Resource.Plural))

	// Generate resource code fragments
	caInjectionPatch := make([]string, 0)
	caInjectionPatch = append(caInjectionPatch, fmt.Sprintf(caInjectionPatchCodeFragment, f.Resource.Plural))

	// Only store code fragments in the map if the slices are non-empty
	if len(res) != 0 {
		fragments[machinery.NewMarkerFor(f.Path, resourceMarker)] = res
	}

	if len(webhookPatch) != 0 {
		fragments[machinery.NewMarkerFor(f.Path, webhookPatchMarker)] = webhookPatch
	}

	if len(caInjectionPatch) != 0 {
		fragments[machinery.NewMarkerFor(f.Path, caInjectionPatchMarker)] = caInjectionPatch
	}

	return fragments
}

const kustomizationTemplate = `# This kustomization.yaml is not intended to be run by itself,
# since it depends on service name and namespace that are out of this kustomize package.
# It should be run by config/default
resources:
%s

patchesStrategicMerge:
# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix.
# patches here are for enabling the conversion webhook for each CRD
%s

# [CERTMANAGER] To enable cert-manager, uncomment all the sections with [CERTMANAGER] prefix.
# patches here are for enabling the CA injection for each CRD
%s

# the following config is for teaching kustomize how to do kustomization for CRDs.
configurations:
- kustomizeconfig.yaml
`
