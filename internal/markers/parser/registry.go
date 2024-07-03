// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package parser

type Registry interface {
	Lookup(name string) bool
	GetDefinition(name string) Definition
}
