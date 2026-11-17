// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package markers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructMarker_GetName(t *testing.T) {
	t.Parallel()

	name := "subscriptionManager"
	emptyName := ""

	tests := []struct {
		name   string
		marker *StructMarker
		want   string
	}{
		{
			name:   "non-nil name is returned",
			marker: &StructMarker{Name: &name},
			want:   "subscriptionManager",
		},
		{
			name:   "nil name returns empty string",
			marker: &StructMarker{},
			want:   "",
		},
		{
			name:   "empty name returns empty string",
			marker: &StructMarker{Name: &emptyName},
			want:   "",
		},
		{
			name: "root sentinel is returned as-is",
			marker: &StructMarker{Name: func() *string {
				s := StructRootName

				return &s
			}()},
			want: ".",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.marker.GetName())
		})
	}
}

func TestStructMarker_GetDescription(t *testing.T) {
	t.Parallel()

	desc := "Manages subscription configuration"

	tests := []struct {
		name   string
		marker *StructMarker
		want   string
	}{
		{
			name:   "non-nil description is returned",
			marker: &StructMarker{Description: &desc},
			want:   "Manages subscription configuration",
		},
		{
			name:   "nil description returns empty string",
			marker: &StructMarker{},
			want:   "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.marker.GetDescription())
		})
	}
}

func TestStructMarker_GetComments(t *testing.T) {
	t.Parallel()

	desc := "Manages subscription configuration"
	emptyDesc := ""

	kubebuilderLine := `+kubebuilder:validation:XValidation:rule="!(self.a=='x'&&self.b=='y')",` +
		`message="a=x is not allowed with b=y; set b explicitly (e.g. z)"`
	kubebuilderDesc := "Manages subscription configuration\n" + kubebuilderLine

	tests := []struct {
		name        string
		marker      *StructMarker
		exceptions  []string
		wantNil     bool
		wantContain string
	}{
		{
			name:        "description is word-wrapped into comment lines",
			marker:      &StructMarker{Description: &desc},
			wantContain: "Manages subscription configuration",
		},
		{
			name:        "kubebuilder line longer than wrap width is not wrapped",
			marker:      &StructMarker{Description: &kubebuilderDesc},
			exceptions:  []string{"+kubebuilder:"},
			wantContain: kubebuilderLine,
		},
		{
			name:    "nil description returns nil",
			marker:  &StructMarker{},
			wantNil: true,
		},
		{
			name:    "empty description returns nil",
			marker:  &StructMarker{Description: &emptyDesc},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.marker.GetComments(tt.exceptions...)
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotEmpty(t, got)
				var found bool

				for _, line := range got {
					if line == tt.wantContain {
						found = true

						break
					}
				}
				assert.True(t, found, "expected %q in comments %v", tt.wantContain, got)
			}
		})
	}
}
