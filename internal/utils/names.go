// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package utils

import (
	"strings"
)

// ToPascalCase will convert a kebab-case string to a PascalCase name appropriate to
// use as a go variable name.
func ToPascalCase(name string) string {
	var output string

	makeUpper := true

	for _, letter := range name {
		if makeUpper {
			output += strings.ToUpper(string(letter))
			makeUpper = false
		} else {
			if letter == '-' {
				makeUpper = true
			} else {
				output += string(letter)
			}
		}
	}

	return output
}

// ToFileName will convert a kebab-case string to a snake_case name appropriate to
// use in a go filename.
func ToFileName(name string) string {
	return strings.ToLower(strings.Replace(name, "-", "_", -1))
}

// ToPackageName will convert a kebab-case string to an all lower name
// appropriate for directory and package names.
func ToPackageName(name string) string {
	return strings.ToLower(strings.Replace(name, "-", "", -1))
}

// ToTitle replaces the strings.Title method, which is deprecated in go1.18.  This is a helper
// method to make titling a string much more readable than the new methodology.
//nolint:godox
// TODO: use commented code below eventually.  It returns different at this time but will
// eventually be deprecated.
func ToTitle(in string) string {
	//nolint:gocritic
	// return cases.Title(language.Und, cases.NoLower, cases.NoLower).String(in)
	return strings.Title(in)
}
