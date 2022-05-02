// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package kinds

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/commands/companion"
)

func Test_ComponentSetNames(t *testing.T) {
	t.Parallel()

	sharedNameInput := WorkloadShared{
		Name: "shared-name",
		Kind: WorkloadKindComponent,
	}

	sharedNameExpected := WorkloadShared{
		Name:        "shared-name",
		PackageName: "sharedname",
		Kind:        WorkloadKindComponent,
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
					API: WorkloadAPISpec{
						Kind: "ComponentWorkloadTest",
					},
					CompanionCliSubcmd: companion.CLI{},
				},
			},
			expected: &ComponentWorkload{
				WorkloadShared: sharedNameExpected,
				Spec: ComponentWorkloadSpec{
					API: WorkloadAPISpec{
						Kind: "ComponentWorkloadTest",
					},
					CompanionCliSubcmd: companion.CLI{
						Name:         "componentworkloadtest",
						Description:  "Manage componentworkloadtest workload",
						VarName:      "Componentworkloadtest",
						FileName:     "componentworkloadtest",
						IsSubcommand: true,
					},
				},
			},
		},
		{
			name: "component workload with subcommand",
			input: &ComponentWorkload{
				WorkloadShared: sharedNameInput,
				Spec: ComponentWorkloadSpec{
					API: WorkloadAPISpec{
						Kind: "ComponentWorkloadTest",
					},
					CompanionCliSubcmd: companion.CLI{
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
					API: WorkloadAPISpec{
						Kind: "ComponentWorkloadTest",
					},
					CompanionCliSubcmd: companion.CLI{
						Name:         "componentworkloadtest",
						Description:  "Manage componentworkloadtest workload custom",
						VarName:      "Componentworkloadtest",
						FileName:     "componentworkloadtest",
						IsSubcommand: true,
					},
				},
			},
		},
		{
			name: "component workload with subcommand but missing description",
			input: &ComponentWorkload{
				WorkloadShared: sharedNameInput,
				Spec: ComponentWorkloadSpec{
					API: WorkloadAPISpec{
						Kind: "ComponentWorkloadTest",
					},
					CompanionCliSubcmd: companion.CLI{
						Name:     "componentworkloadtest",
						VarName:  "Componentworkloadtest",
						FileName: "componentworkloadtest",
					},
				},
			},
			expected: &ComponentWorkload{
				WorkloadShared: sharedNameExpected,
				Spec: ComponentWorkloadSpec{
					API: WorkloadAPISpec{
						Kind: "ComponentWorkloadTest",
					},
					CompanionCliSubcmd: companion.CLI{
						Name:         "componentworkloadtest",
						Description:  "Manage componentworkloadtest workload",
						VarName:      "Componentworkloadtest",
						FileName:     "componentworkloadtest",
						IsSubcommand: true,
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
