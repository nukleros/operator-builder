// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package inspect

import (
	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/marker"
	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/parser"
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
