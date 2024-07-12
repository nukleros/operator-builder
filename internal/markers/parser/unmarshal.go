// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package parser

type Unmarshaler interface {
	UnmarshalMarkerArg(in string) error
}
