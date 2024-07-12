// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package parser

type Registry interface {
	Lookup(name string) bool
	GetDefinition(name string) Definition
}
