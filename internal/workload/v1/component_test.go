// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

//nolint:testpackage
package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ComponentSetNames(t *testing.T) {
	t.Parallel()

	sharedNameInput := WorkloadShared{
		Name: "shared-name",
		Kind: "ComponentWorkload",
	}

	sharedNameExpected := WorkloadShared{
		Name:        "shared-name",
		PackageName: "sharedname",
		Kind:        "ComponentWorkload",
	}

	for _, tt := range []struct {
		name     string
		input    *ComponentWorkload
		expected *ComponentWorkload
	}{
		{
			name: "component workload missing subcommand",
			input: &ComponentWorkload{
				WorkloadShared: sharedNameInput,
				Spec: ComponentWorkloadSpec{
					API: APISpec{
						Kind: "ComponentWorkloadTest",
					},
					CompanionCliSubcmd: CliCommand{},
				},
			},
			expected: &ComponentWorkload{
				WorkloadShared: sharedNameExpected,
				Spec: ComponentWorkloadSpec{
					API: APISpec{
						Kind: "ComponentWorkloadTest",
					},
					CompanionCliSubcmd: CliCommand{
						Name:        "componentworkloadtest",
						Description: "Manage componentworkloadtest workload",
						VarName:     "Componentworkloadtest",
						FileName:    "componentworkloadtest",
					},
				},
			},
		},
		{
			name: "component workload with subcommand",
			input: &ComponentWorkload{
				WorkloadShared: sharedNameInput,
				Spec: ComponentWorkloadSpec{
					API: APISpec{
						Kind: "ComponentWorkloadTest",
					},
					CompanionCliSubcmd: CliCommand{
						Name:        "componentworkloadtest",
						Description: "Manage componentworkloadtest workload custom",
						VarName:     "Componentworkloadtest",
						FileName:    "componentworkloadtest",
					},
				},
			},
			expected: &ComponentWorkload{
				WorkloadShared: sharedNameExpected,
				Spec: ComponentWorkloadSpec{
					API: APISpec{
						Kind: "ComponentWorkloadTest",
					},
					CompanionCliSubcmd: CliCommand{
						Name:        "componentworkloadtest",
						Description: "Manage componentworkloadtest workload custom",
						VarName:     "Componentworkloadtest",
						FileName:    "componentworkloadtest",
					},
				},
			},
		},
		{
			name: "component workload with subcommand but missing description",
			input: &ComponentWorkload{
				WorkloadShared: sharedNameInput,
				Spec: ComponentWorkloadSpec{
					API: APISpec{
						Kind: "ComponentWorkloadTest",
					},
					CompanionCliSubcmd: CliCommand{
						Name:     "componentworkloadtest",
						VarName:  "Componentworkloadtest",
						FileName: "componentworkloadtest",
					},
				},
			},
			expected: &ComponentWorkload{
				WorkloadShared: sharedNameExpected,
				Spec: ComponentWorkloadSpec{
					API: APISpec{
						Kind: "ComponentWorkloadTest",
					},
					CompanionCliSubcmd: CliCommand{
						Name:        "componentworkloadtest",
						Description: "Manage componentworkloadtest workload",
						VarName:     "Componentworkloadtest",
						FileName:    "componentworkloadtest",
					},
				},
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.input.SetNames()
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}
