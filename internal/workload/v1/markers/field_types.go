// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

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
	FieldStringSlice
	FieldStruct
)

// UnmarshalMarkerArg will convert the type argument within a field or collection
// field marker into its underlying FieldType object.
func (f *FieldType) UnmarshalMarkerArg(in string) error {
	types := map[string]FieldType{
		"":            FieldUnknownType,
		"string":      FieldString,
		"int":         FieldInt,
		"bool":        FieldBool,
		"stringArray": FieldStringSlice,
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

// String returns the marker keyword for a FieldType (used in marker arguments).
func (f FieldType) String() string {
	types := map[FieldType]string{
		FieldUnknownType: "",
		FieldString:      "string",
		FieldInt:         "int",
		FieldBool:        "bool",
		FieldStringSlice: "stringArray",
		FieldStruct:      "struct",
	}

	return types[f]
}

// GoTypeName returns the Go type name for a FieldType (used in generated source code).
func (f FieldType) GoTypeName() string {
	types := map[FieldType]string{
		FieldUnknownType: "",
		FieldString:      "string",
		FieldInt:         "int",
		FieldBool:        "bool",
		FieldStringSlice: "[]string",
		FieldStruct:      "struct",
	}

	return types[f]
}
