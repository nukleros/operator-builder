// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package marker

import "github.com/vmware-tanzu-labs/operator-builder/internal/markers/parser"

type Registry struct {
	registry map[string]*Definition
	Results  chan interface{}
}

func NewRegistry() *Registry {
	return &Registry{
		registry: make(map[string]*Definition),
		Results:  make(chan interface{}),
	}
}

func (r *Registry) Add(marker *Definition) {
	r.registry[marker.Name] = marker
}

func (r *Registry) Lookup(name string) bool {
	_, found := r.registry[name]

	return found
}

func (r *Registry) GetDefinition(name string) parser.Definition {
	m := r.registry[name]

	marker := *m

	marker.Fields = make(map[string]Argument)

	for k, v := range m.Fields {
		marker.Fields[k] = v
	}

	return &marker
}
