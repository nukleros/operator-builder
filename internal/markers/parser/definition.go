// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package parser

type Definition interface {
	GetName() string
	LookupArgument(name string) bool
	SetArgument(name string, value interface{}) error
	InflateObject() (interface{}, error)
}

func (p *Parser) loadDefinition() (found bool) {
	if ok := p.registry.Lookup(p.scopeBuffer[:len(p.scopeBuffer)-1]); ok {
		p.currentDefinition = p.registry.GetDefinition(p.scopeBuffer[:len(p.scopeBuffer)-1])

		return true
	}

	return found
}
