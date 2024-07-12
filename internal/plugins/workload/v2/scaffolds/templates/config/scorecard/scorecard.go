// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

// NOTE: this was copied from operator-sdk in order to
// include support for OpenShift Lifecycle Manager.  It was
// copied because operator-sdk templates are internal to the
// repo and unable to be imported directly.  Including original
// license with this file.

// Copyright 2021 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scorecard

import (
	"errors"
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
)

var _ machinery.Template = &Scorecard{}

var (
	ErrUnknwonScorecardType = errors.New("unknown scorecard type")
)

//nolint:golint
type ScorecardType int

const (
	kustomizeFileSubPath    = "kustomization.yaml"
	operatorSdkImageVersion = "v1.28.0"

	ScorecardTypeUnknown ScorecardType = iota
	ScorecardTypeBase
	ScorecardTypeKustomize
	ScorecardTypePatchesOLM
	ScorecardTypePatchesBasic
)

// Scorecard scaffolds a file which represents an Operator Lifecycle Manager scorecard.
// It is only used when --enable-olm is set to true.
type Scorecard struct {
	machinery.TemplateMixin

	// input variables
	ScorecardTestImage string
	ScorecardType      ScorecardType
}

func (f *Scorecard) SetTemplateDefaults() error {
	if f.ScorecardType == ScorecardTypeUnknown {
		return ErrUnknwonScorecardType
	}

	f.Path = filepath.Join(append([]string{"config", "scorecard"}, getScorecardSubPath(f.ScorecardType)...)...)
	f.TemplateBody = getScorecardTemplate(f.ScorecardType)

	f.ScorecardTestImage = fmt.Sprintf("quay.io/operator-framework/scorecard-test:%s", operatorSdkImageVersion)
	f.IfExistsAction = machinery.OverwriteFile

	return nil
}

const (
	// kustomizationFile is a kustomization.yaml file for the scorecard componentconfig.
	// This should always be written to config/scorecard/kustomization.yaml.
	kustomizeFile = `resources:
  - bases/config.yaml
patchesJson6902:
  - path: patches/basic.config.yaml
    target:
      group: scorecard.operatorframework.io
      version: v1alpha3
      kind: Configuration
      name: config
  - path: patches/olm.config.yaml
    target:
      group: scorecard.operatorframework.io
      version: v1alpha3
      kind: Configuration
      name: config
%[1]s
`

	// YAML file marker to append to kustomization.yaml files.
	patchesJSON6902Marker = "patchesJson6902"

	// Config is an empty scorecard componentconfig with parallel stages.
	config = `apiVersion: scorecard.operatorframework.io/v1alpha3
kind: Configuration
metadata:
  name: config
stages:
  - parallel: true
    tests: []
`

	// PatchBasic contains all default "basic" test configurations.
	patchBasic = `- op: add
  path: /stages/0/tests/-
  value:
    entrypoint:
      - scorecard-test
      - basic-check-spec
    image: {{ .ScorecardTestImage }}
    labels:
      suite: basic
      test: basic-check-spec-test
`

	// PatchOLM contains all default "olm" test configurations.
	patchOLM = `- op: add
  path: /stages/0/tests/-
  value:
    entrypoint:
      - scorecard-test
      - olm-bundle-validation
    image: {{ .ScorecardTestImage }}
    labels:
      suite: olm
      test: olm-bundle-validation-test
- op: add
  path: /stages/0/tests/-
  value:
    entrypoint:
      - scorecard-test
      - olm-crds-have-validation
    image: {{ .ScorecardTestImage }}
    labels:
      suite: olm
      test: olm-crds-have-validation-test
- op: add
  path: /stages/0/tests/-
  value:
    entrypoint:
      - scorecard-test
      - olm-crds-have-resources
    image: {{ .ScorecardTestImage }}
    labels:
      suite: olm
      test: olm-crds-have-resources-test
- op: add
  path: /stages/0/tests/-
  value:
    entrypoint:
      - scorecard-test
      - olm-spec-descriptors
    image: {{ .ScorecardTestImage }}
    labels:
      suite: olm
      test: olm-spec-descriptors-test
- op: add
  path: /stages/0/tests/-
  value:
    entrypoint:
      - scorecard-test
      - olm-status-descriptors
    image: {{ .ScorecardTestImage }}
    labels:
      suite: olm
      test: olm-status-descriptors-test
`
)

func getScorecardSubPath(scorecardType ScorecardType) []string {
	return map[ScorecardType][]string{
		ScorecardTypeBase:         {"bases", "config.yaml"},
		ScorecardTypeKustomize:    {kustomizeFileSubPath},
		ScorecardTypePatchesBasic: {"patches", "basic.config.yaml"},
		ScorecardTypePatchesOLM:   {"patches", "olm.config.yaml"},
	}[scorecardType]
}

func getScorecardTemplate(scorecardType ScorecardType) string {
	return map[ScorecardType]string{
		ScorecardTypeBase:         config,
		ScorecardTypeKustomize:    fmt.Sprintf(kustomizeFile, machinery.NewMarkerFor(kustomizeFileSubPath, patchesJSON6902Marker)),
		ScorecardTypePatchesBasic: patchBasic,
		ScorecardTypePatchesOLM:   patchOLM,
	}[scorecardType]
}
