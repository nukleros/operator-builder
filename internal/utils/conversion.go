// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package utils

import (
	"errors"
	"fmt"
)

var (
	ErrConvertArrayInterface     = errors.New("unable to convert to []interface{}")
	ErrConvertArrayString        = errors.New("unable to convert to []string")
	ErrConvertMapStringInterface = errors.New("unable to convert to map[string]interface{}")
	ErrConvertString             = errors.New("unable to convert to string")
)

// ToArrayInterface attempts a conversion from an interface to an underlying array of
// interface type.  Returns an error if the conversion is not possible.
func ToArrayInterface(in interface{}) ([]interface{}, error) {
	out, ok := in.([]interface{})
	if !ok {
		return nil, ErrConvertArrayInterface
	}

	return out, nil
}

// ToArrayString attempts a conversion from an interface to an underlying array of
// string type.  Returns an error if the conversion is not possible.
func ToArrayString(in interface{}) ([]string, error) {
	// attempt direct conversion
	out, ok := in.([]string)
	if !ok {
		// attempt conversion for each item
		outInterfaces, err := ToArrayInterface(in)
		if err != nil {
			return nil, fmt.Errorf("%w; %s", err, ErrConvertArrayString)
		}

		outStrings := make([]string, len(outInterfaces))

		for i := range outInterfaces {
			outString, err := ToString(outInterfaces[i])
			if err != nil {
				return nil, fmt.Errorf("%w; %s", err, ErrConvertArrayString)
			}

			outStrings[i] = outString
		}

		return outStrings, nil
	}

	return out, nil
}

// ToMapStringInterface attempts a conversion from an interface to an underlying map
// string interface type.  Returns an error if the conversion is not possible.
func ToMapStringInterface(in interface{}) (map[string]interface{}, error) {
	out, ok := in.(map[string]interface{})
	if !ok {
		return nil, ErrConvertMapStringInterface
	}

	return out, nil
}

// ToArrayInterface attempts a conversion from an interface to an underlying
// string type.  Returns an error if the conversion is not possible.
func ToString(in interface{}) (string, error) {
	out, ok := in.(string)
	if !ok {
		return "", ErrConvertString
	}

	return out, nil
}
