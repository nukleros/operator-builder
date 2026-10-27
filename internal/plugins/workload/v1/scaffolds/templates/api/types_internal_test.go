// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTypesTemplateStatusFields checks the status bool tags and the list tags.
func TestTypesTemplateStatusFields(t *testing.T) {
	t.Parallel()

	// check bool fields stay optional and keep a false value
	assertStatusTag(t, "Created", "`json:\"created\"`", true)
	assertStatusTag(t, "DependenciesSatisfied", "`json:\"dependenciesSatisfied\"`", true)

	// check list fields still omit an empty value
	assertStatusTag(t, "Conditions", "`json:\"conditions,omitempty\"`", false)
	assertStatusTag(t, "Resources", "`json:\"resources,omitempty\"`", false)
}

// assertStatusTag checks the json tag on one status field in typesTemplate.
func assertStatusTag(t *testing.T, field, tag string, optional bool) {
	t.Helper()

	// find the field line and check its tag
	lines := strings.Split(typesTemplate, "\n")
	for i, line := range lines {
		if !strings.Contains(line, field+" ") {
			continue
		}

		assert.Contains(t, line, tag)
		if optional {
			assert.Contains(t, previousNonBlank(lines, i), "+kubebuilder:validation:Optional")
		}

		return
	}

	t.Errorf("field %s not found in types template", field)
}

// previousNonBlank returns the nearest non-blank line above index.
func previousNonBlank(lines []string, index int) string {
	// walk up to the nearest line that has text
	for i := index - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return lines[i]
		}
	}

	return ""
}
