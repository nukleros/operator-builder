// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package marker

import (
	"strings"
	"unicode"
)

// lowerCamelCase converts PascalCase string to
// a camelCase string (by lowering the first rune).
func lowerCamelCase(in string) string {
	isFirst := true

	return strings.Map(
		func(inRune rune) rune {
			if isFirst {
				isFirst = false

				return unicode.ToLower(inRune)
			}

			return inRune
		},
		in,
	)
}
