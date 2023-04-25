// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package markers

import (
	"errors"
	"fmt"
)

var ErrUnableToParseFieldType = errors.New("unable to parse field")

// FieldType defines the types of fields for a field marker that are accepted
// during parsing of a manifest.
type FieldType int

const (
	FieldUnknownType FieldType = iota
	FieldString
	FieldInt
	FieldBool
	FieldStruct
)

// UnmarshalMarkerArg will convert the type argument within a field or collection
// field marker into its underlying FieldType object.
func (f *FieldType) UnmarshalMarkerArg(in string) error {
	types := map[string]FieldType{
		"":       FieldUnknownType,
		"string": FieldString,
		"int":    FieldInt,
		"bool":   FieldBool,
	}

	if t, ok := types[in]; ok {
		if t == FieldUnknownType {
			return fmt.Errorf("%w, %s into FieldType", ErrUnableToParseFieldType, in)
		}

		*f = t

		return nil
	}

	return fmt.Errorf("%w, %s into FieldType", ErrUnableToParseFieldType, in)
}

// String simply returns a FieldType in string format.
func (f FieldType) String() string {
	types := map[FieldType]string{
		FieldUnknownType: "",
		FieldString:      "string",
		FieldInt:         "int",
		FieldBool:        "bool",
		FieldStruct:      "struct",
	}

	return types[f]
}
