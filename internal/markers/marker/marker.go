// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package marker

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrTypeMustBeStruct = errors.New("type must be a struct")
	ErrArgNotFound      = errors.New("argument not found")
	ErrMissingArguments = errors.New("missing arguments")
)

type Definition struct {
	Name   string
	Output reflect.Type
	Fields map[string]Argument
}

func (m Definition) String() string {
	return m.Name
}

func Define(name string, outputType interface{}) (*Definition, error) {
	m := &Definition{
		Name:   name,
		Output: reflect.TypeOf(outputType),
	}

	if err := m.loadFields(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Definition) GetName() string {
	return m.Name
}

func (m *Definition) LookupArgument(argName string) bool {
	_, found := m.Fields[argName]

	return found
}

func (m *Definition) SetArgument(argName string, value interface{}) error {
	if arg, found := m.Fields[argName]; found {
		if err := arg.SetValue(value); err != nil {
			return fmt.Errorf("%w on arg %s", err, argName)
		}

		m.Fields[argName] = arg

		return nil
	}

	return fmt.Errorf("%w %q for marker %s", ErrArgNotFound, argName, m.Name)
}

func (m *Definition) InflateObject() (interface{}, error) {
	o := reflect.Indirect(reflect.New(m.Output))

	var missingFields []string

	for argName, arg := range m.Fields {
		field := o.FieldByName(arg.FieldName)

		if !arg.isSet {
			if !arg.Optional {
				missingFields = append(missingFields, argName)

				continue
			}

			if !arg.Pointer {
				arg.InitializeValue()
			}
		}

		if arg.Value.IsValid() {
			field.Set(arg.Value)
		}
	}

	if len(missingFields) > 0 {
		return nil, fmt.Errorf("%w: %q", ErrMissingArguments, missingFields)
	}

	return o.Interface(), nil
}

func (m *Definition) loadFields() error {
	if m.Fields == nil {
		m.Fields = make(map[string]Argument)
	}

	if m.Output.Kind() != reflect.Struct {
		return ErrTypeMustBeStruct
	}

	for i := 0; i < m.Output.NumField(); i++ {
		field := m.Output.Field(i)
		if field.PkgPath != "" {
			// as per the reflect package docs, pkgpath is empty for exported fields,
			// so non-empty package path means a private field, which we should skip
			continue
		}

		arg, err := ArgumentFromField(&field)
		if err != nil {
			return fmt.Errorf("unable to extract type information for field %q: %w", field.Name, err)
		}

		m.Fields[arg.Name] = arg
	}

	return nil
}
