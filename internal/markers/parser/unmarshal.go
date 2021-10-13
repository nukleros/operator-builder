// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package parser

type Unmarshaler interface {
	UnmarshalMarkerArg(in string) error
}
