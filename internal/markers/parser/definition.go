// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

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
