// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package utils

import "strings"

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

// needsUnderscore has logic that determines when a character needs to be replaced
// with an underscore.  It is needed by the ToSnakeCase function.
func needsUnderscore(char rune) bool {
	for _, s := range []string{"/", ".", "-", "~", `\`} {
		//nolint:gocritic
		if char == []rune(s)[0] {
			return true
		}
	}

	return false
}

// ToSnakeCase will convert a string to a snake case.
func ToSnakeCase(name string) string {
	var buff strings.Builder

	diff := 'a' - 'A'

	nameSize := len(name)

	for i, char := range name {
		if needsUnderscore(char) {
			//nolint:gocritic
			buff.WriteRune([]rune("_")[0])

			continue
		}

		// A is 65, a is 97
		if char >= 'a' {
			buff.WriteRune(char)

			continue
		}

		// v is capital letter here
		// irregard first letter
		// add underscore if last letter is capital letter
		// add underscore when previous letter is lowercase
		// add underscore when next letter is lowercase
		if (i != 0 || i == nameSize-1) && (  // head and tail
		(i > 0 && rune(name[i-1]) >= 'a') || // pre
			(i < nameSize-1 && rune(name[i+1]) >= 'a')) { // next
			buff.WriteRune('_')
		}

		buff.WriteRune(char + diff)
	}

	return buff.String()
}

// ToTitle replaces the strings.Title method, which is deprecated in go1.18.  This is a helper
// method to make titling a string much more readable than the new methodology.
// TODO: use commented code below eventually.  It returns different at this time but will
// eventually be deprecated.
//
//nolint:godox
func ToTitle(in string) string {
	//nolint:gocritic
	// return cases.Title(language.Und, cases.NoLower, cases.NoLower).String(in)
	return strings.Title(in)
}
