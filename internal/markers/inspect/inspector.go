// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package inspect

import (
	"github.com/nukleros/operator-builder/internal/markers/marker"
	"github.com/nukleros/operator-builder/internal/markers/parser"
)

type Inspector struct {
	Registry *marker.Registry
}

func NewInspector(registry *marker.Registry) *Inspector {
	return &Inspector{
		Registry: registry,
	}
}

func (s *Inspector) parse(input string) (results []*parser.Result) {
	p := parser.NewParser(input, s.Registry)

	return p.Parse()
}
