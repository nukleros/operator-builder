// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package kinds

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nukleros/operator-builder/internal/workload/v1/markers"
)

func strPtr(s string) *string {
	return &s
}

func TestWorkloadSpec_applyStructMarkers(t *testing.T) {
	t.Parallel()

	buildAPISpecFields := func() *APIFields {
		root := &APIFields{
			Name:   "Spec",
			Type:   markers.FieldStruct,
			Sample: "spec:",
		}
		_ = root.AddField("webstore.image", markers.FieldString, nil, "nginx", true)

		return root
	}

	tests := []struct {
		name          string
		structMarkers []*markers.StructMarker
		wantErr       bool
	}{
		{
			name: "single struct marker sets comments",
			structMarkers: []*markers.StructMarker{
				{Name: strPtr("webstore"), Description: strPtr("Manages the webstore configuration")},
			},
			wantErr: false,
		},
		{
			name: "single root struct marker sets comments",
			structMarkers: []*markers.StructMarker{
				{Name: strPtr(markers.StructRootName), Description: strPtr("Root description")},
			},
			wantErr: false,
		},
		{
			name: "two struct markers for different paths both apply",
			structMarkers: []*markers.StructMarker{
				{Name: strPtr("webstore"), Description: strPtr("First description")},
				{Name: strPtr(markers.StructRootName), Description: strPtr("Root description")},
			},
			wantErr: false,
		},
		{
			name: "duplicate struct markers for the same path are rejected",
			structMarkers: []*markers.StructMarker{
				{Name: strPtr("webstore"), Description: strPtr("First description")},
				{Name: strPtr("webstore"), Description: strPtr("Second description")},
			},
			wantErr: true,
		},
		{
			name: "duplicate root struct markers are rejected",
			structMarkers: []*markers.StructMarker{
				{Name: strPtr(markers.StructRootName), Description: strPtr("First description")},
				{Name: strPtr(markers.StructRootName), Description: strPtr("Second description")},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ws := &WorkloadSpec{
				APISpecFields: buildAPISpecFields(),
				StructMarkers: tt.structMarkers,
			}

			err := ws.applyStructMarkers()
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
		})
	}
}

// TestWorkloadSpec_applyStructMarkers_kubebuilderNotWrapped checks a +kubebuilder: line
// longer than the wrap width is kept intact on its own comment line.
func TestWorkloadSpec_applyStructMarkers_kubebuilderNotWrapped(t *testing.T) {
	t.Parallel()

	kubebuilderLine := `+kubebuilder:validation:XValidation:rule="!(self.a=='x'&&self.b=='y')",` +
		`message="a=x is not allowed with b=y; set b explicitly (e.g. z)"`

	root := &APIFields{
		Name:   "Spec",
		Type:   markers.FieldStruct,
		Sample: "spec:",
	}
	_ = root.AddField("webstore.image", markers.FieldString, nil, "nginx", true)

	ws := &WorkloadSpec{
		APISpecFields: root,
		StructMarkers: []*markers.StructMarker{
			{
				Name:        strPtr(markers.StructRootName),
				Description: strPtr("Root description\n" + kubebuilderLine),
			},
		},
	}

	assert.NoError(t, ws.applyStructMarkers())
	assert.Equal(t, []string{"", "Root description", kubebuilderLine}, ws.APISpecFields.Comments)
}
