// Copyright 2024 Nukleros
// SPDX-License-Identifier: MIT

package workload

import (
	"os"
	"testing"
)

func TestFromEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want PluginVersion
		env  map[string]string
	}{
		{
			name: "ensure empty environment returns default version",
			want: DefaultPluginVersion,
		},
		{
			name: "ensure v1 environment returns v1 version",
			want: PluginVersionV1,
			env: map[string]string{
				EnvPluginVersionVariable: EnvPluginVersionV1,
			},
		},
		{
			name: "ensure v2 environment returns v2 version",
			want: PluginVersionV2,
			env: map[string]string{
				EnvPluginVersionVariable: EnvPluginVersionV2,
			},
		},
		{
			name: "ensure invalid environment returns unknown version",
			want: PluginVersionUnknown,
			env: map[string]string{
				EnvPluginVersionVariable: "fake",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tt.env {
				_ = os.Setenv(k, v)
			}

			if got := FromEnv(); got != tt.want {
				t.Errorf("FromEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
