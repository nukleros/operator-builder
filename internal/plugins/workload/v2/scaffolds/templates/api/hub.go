// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"fmt"
	log "log/slog"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
)

var _ machinery.Template = &Hub{}

// Hub scaffolds the file that defines hub.
type Hub struct {
	machinery.TemplateMixin
	machinery.MultiGroupMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin

	Force bool
}

// SetTemplateDefaults implements file.Template.
func (f *Hub) SetTemplateDefaults() error {
	f.Path = filepath.Join(
		"apis",
		f.Resource.Group,
		f.Resource.Version,
		fmt.Sprintf("%s_conversion.go", strings.ToLower(f.Resource.Kind)),
	)

	f.Path = f.Resource.Replacer().Replace(f.Path)
	log.Info(f.Path)

	f.TemplateBody = hubTemplate

	if f.Force {
		f.IfExistsAction = machinery.OverwriteFile
	} else {
		f.IfExistsAction = machinery.SkipFile
	}

	return nil
}

const hubTemplate = `{{ .Boilerplate }}

package {{ .Resource.Version }}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// Hub marks this type as a conversion hub.
func (*{{ .Resource.Kind }}) Hub() {}
`
