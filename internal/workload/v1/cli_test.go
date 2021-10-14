// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

//nolint:testpackage
package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_setCommonValues(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name                string
		kind                string
		descriptionTemplate string
		input               CliCommand
		expected            CliCommand
	}{
		{
			name:                "command without description",
			kind:                "MissingDescription",
			descriptionTemplate: `Manage %s test`,
			input: CliCommand{
				Name: "MissingDescriptionTest",
			},
			expected: CliCommand{
				Name:        "MissingDescriptionTest",
				VarName:     "MissingDescriptionTest",
				FileName:    "missingdescriptiontest",
				Description: "Manage missingdescription test",
			},
		},
		{
			name:                "command with description",
			kind:                "HasDescription",
			descriptionTemplate: `Manage %s test`,
			input: CliCommand{
				Name:        "HasDescriptionTest",
				Description: "This is my command description",
			},
			expected: CliCommand{
				Name:        "HasDescriptionTest",
				VarName:     "HasDescriptionTest",
				FileName:    "hasdescriptiontest",
				Description: "This is my command description",
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.input.setCommonValues(tt.kind, tt.descriptionTemplate)
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

func Test_setSubCommandValues(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name                string
		kind                string
		descriptionTemplate string
		input               CliCommand
		expected            CliCommand
	}{
		{
			name:                "subcommand without name",
			kind:                "MissingName",
			descriptionTemplate: `Manage %s test`,
			input:               CliCommand{},
			expected: CliCommand{
				Name:        "missingname",
				VarName:     "Missingname",
				FileName:    "missingname",
				Description: "Manage missingname test",
			},
		},
		{
			name:                "subcommand with name",
			kind:                "HasName",
			descriptionTemplate: `Manage %s test`,
			input: CliCommand{
				Name: "hasname",
			},
			expected: CliCommand{
				Name:        "hasname",
				VarName:     "Hasname",
				FileName:    "hasname",
				Description: "Manage hasname test",
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.input.setSubCommandValues(tt.kind, tt.descriptionTemplate)
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

func Test_hasName(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		input    CliCommand
		expected bool
	}{
		{
			name: "command has a name field",
			input: CliCommand{
				Name: "HasNameField",
			},
			expected: true,
		},
		{
			name:     "command does not have a name field",
			input:    CliCommand{},
			expected: false,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hasName := tt.input.hasName()
			assert.Equal(t, tt.expected, hasName)
		})
	}
}

func Test_hasDescription(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		input    CliCommand
		expected bool
	}{
		{
			name: "command has a description field",
			input: CliCommand{
				Description: "HasDescriptionField",
			},
			expected: true,
		},
		{
			name:     "command does not have a description field",
			input:    CliCommand{},
			expected: false,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hasDescription := tt.input.hasDescription()
			assert.Equal(t, tt.expected, hasDescription)
		})
	}
}
