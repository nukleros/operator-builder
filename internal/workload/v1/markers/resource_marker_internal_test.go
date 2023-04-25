// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package markers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nukleros/operator-builder/internal/markers/marker"
)

func TestResourceMarker_String(t *testing.T) {
	t.Parallel()

	testField := "rmtest"
	testIncludeTrue := true

	type fields struct {
		Field           *string
		CollectionField *string
		Value           interface{}
		Include         *bool
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure resource marker with set field values output matches expected",
			fields: fields{
				Field:   &testField,
				Value:   testField,
				Include: &testIncludeTrue,
			},
			want: "ResourceMarker{Field: rmtest CollectionField:  Value: rmtest Include: true}",
		},
		{
			name: "ensure resource marker with set collection field values output matches expected",
			fields: fields{
				CollectionField: &testField,
				Value:           testField,
				Include:         &testIncludeTrue,
			},
			want: "ResourceMarker{Field:  CollectionField: rmtest Value: rmtest Include: true}",
		},
		{
			name: "ensure resource marker with nil values output matches expected",
			fields: fields{
				Field:           nil,
				CollectionField: nil,
				Value:           nil,
				Include:         nil,
			},
			want: "ResourceMarker{Field:  CollectionField:  Value: <nil> Include: false}",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := ResourceMarker{
				Field:           tt.fields.Field,
				CollectionField: tt.fields.CollectionField,
				Value:           tt.fields.Value,
				Include:         tt.fields.Include,
			}
			if got := rm.String(); got != tt.want {
				t.Errorf("ResourceMarker.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defineResourceMarker(t *testing.T) {
	t.Parallel()

	type args struct {
		registry *marker.Registry
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ensure valid registry can properly add a resource marker",
			args: args{
				registry: marker.NewRegistry(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := defineResourceMarker(tt.args.registry); (err != nil) != tt.wantErr {
				t.Errorf("defineResourceMarker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceMarker_GetIncludeCode(t *testing.T) {
	t.Parallel()

	type fields struct {
		includeCode string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure resource include code returns as expected",
			fields: fields{
				includeCode: "resource marker include code",
			},
			want: "resource marker include code",
		},
		{
			name: "ensure resource include code with empty value returns as expected",
			fields: fields{
				includeCode: "",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := &ResourceMarker{
				includeCode: tt.fields.includeCode,
			}
			if got := rm.GetIncludeCode(); got != tt.want {
				t.Errorf("FieldMarker.GetIncludeCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceMarker_GetSpecPrefix(t *testing.T) {
	t.Parallel()

	field := "this.is.a.spec.test.prefix"

	type fields struct {
		Field           *string
		CollectionField *string
		Value           interface{}
		Include         *bool
		includeCode     string
		fieldMarker     FieldMarkerProcessor
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure a resource marker with field marker returns field marker prefix",
			fields: fields{
				Field: &field,
			},
			want: FieldSpecPrefix,
		},
		{
			name: "ensure a resource marker with collection field marker returns collection field marker prefix",
			fields: fields{
				CollectionField: &field,
			},
			want: CollectionFieldSpecPrefix,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := &ResourceMarker{
				Field:           tt.fields.Field,
				CollectionField: tt.fields.CollectionField,
				Value:           tt.fields.Value,
				Include:         tt.fields.Include,
				includeCode:     tt.fields.includeCode,
				fieldMarker:     tt.fields.fieldMarker,
			}
			if got := rm.GetSpecPrefix(); got != tt.want {
				t.Errorf("ResourceMarker.GetSpecPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceMarker_GetField(t *testing.T) {
	t.Parallel()

	rm := "test.field"

	type fields struct {
		Field *string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure field with value returns as expected",
			fields: fields{
				Field: &rm,
			},
			want: rm,
		},
		{
			name: "ensure field with nil value returns as expected",
			fields: fields{
				Field: nil,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := &ResourceMarker{
				Field: tt.fields.Field,
			}
			if got := rm.GetField(); got != tt.want {
				t.Errorf("ResourceMarker.GetField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceMarker_GetCollectionField(t *testing.T) {
	t.Parallel()

	rm := "test.collection.field"

	type fields struct {
		CollectionField *string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure collection field with value returns as expected",
			fields: fields{
				CollectionField: &rm,
			},
			want: rm,
		},
		{
			name: "ensure collection field with nil value returns as expected",
			fields: fields{
				CollectionField: nil,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := &ResourceMarker{
				CollectionField: tt.fields.CollectionField,
			}
			if got := rm.GetCollectionField(); got != tt.want {
				t.Errorf("ResourceMarker.GetCollectionField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceMarker_GetName(t *testing.T) {
	t.Parallel()

	field := "test.get.name.field"
	collectionField := "test.get.name.collection.field"

	type fields struct {
		Field           *string
		CollectionField *string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure field with value returns as field value",
			fields: fields{
				Field: &field,
			},
			want: field,
		},
		{
			name: "ensure collection field with value returns collection field value",
			fields: fields{
				CollectionField: &collectionField,
			},
			want: collectionField,
		},
		{
			name: "ensure marker with both values returns field value",
			fields: fields{
				Field:           &field,
				CollectionField: &collectionField,
			},
			want: field,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := &ResourceMarker{
				Field:           tt.fields.Field,
				CollectionField: tt.fields.CollectionField,
			}
			if got := rm.GetName(); got != tt.want {
				t.Errorf("ResourceMarker.GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceMarker_validate(t *testing.T) {
	t.Parallel()

	testField := "test.validate"
	testValue := "testValue"
	testInclude := true

	type fields struct {
		Field           *string
		CollectionField *string
		Value           interface{}
		Include         *bool
		fieldMarker     FieldMarkerProcessor
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid resource marker does not produce error",
			fields: fields{
				Field:       &testField,
				Value:       &testValue,
				Include:     &testInclude,
				fieldMarker: nil,
			},
			wantErr: false,
		},
		{
			name: "nil include value produces error",
			fields: fields{
				Field:       &testField,
				Value:       &testValue,
				fieldMarker: nil,
			},
			wantErr: true,
		},
		{
			name: "missing field produces error",
			fields: fields{
				Value:       &testValue,
				Include:     &testInclude,
				fieldMarker: nil,
			},
			wantErr: true,
		},
		{
			name: "missing value produces error",
			fields: fields{
				Field:       &testField,
				Include:     &testInclude,
				fieldMarker: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := &ResourceMarker{
				Field:           tt.fields.Field,
				CollectionField: tt.fields.CollectionField,
				Value:           tt.fields.Value,
				Include:         tt.fields.Include,
				fieldMarker:     tt.fields.fieldMarker,
			}
			if err := rm.validate(); (err != nil) != tt.wantErr {
				t.Errorf("ResourceMarker.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceMarker_isAssociated(t *testing.T) {
	t.Parallel()

	randomString := "thisIsRandom"
	testMarkerString := "test"
	testCollectionString := "test.collection"

	// this is the test of a standard marker which would be discovered in any manifest type
	// standalone = no collection involved
	// component  = the field was set on itself with operator-builder:field marker
	// collection = the field was set on itself with operator-builder:field or
	//              operator-builder:collection:field marker.  It is important to note
	//              that all field markers on a collection are immediately discovered
	//              as field markers, regardless of how they are labeled.
	testFieldMarker := &FieldMarker{
		Name:          &testMarkerString,
		Type:          FieldString,
		forCollection: false,
	}

	// this is the test of a standard marker which was discovered on a collection
	testFieldMarkerOnCollection := &FieldMarker{
		Name:          &testCollectionString,
		Type:          FieldString,
		forCollection: true,
	}

	// this is the test of a standard collection marker.
	testCollectionMarker := &CollectionFieldMarker{
		Name:          &testMarkerString,
		Type:          FieldString,
		forCollection: false,
	}

	type fields struct {
		Field           *string
		CollectionField *string
		Value           interface{}
		Include         *bool
		includeCode     string
		fieldMarker     FieldMarkerProcessor
	}

	type args struct {
		fromMarker FieldMarkerProcessor
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "operator-builder:resource:field=test is associated with a field marker",
			fields: fields{
				Field: testFieldMarker.Name,
			},
			args: args{
				fromMarker: testFieldMarker,
			},
			want: true,
		},
		{
			name: "operator-builder:resource:field=test is not associated with a collection field marker",
			fields: fields{
				Field: testCollectionMarker.Name,
			},
			args: args{
				fromMarker: testCollectionMarker,
			},
			want: false,
		},
		{
			name: "operator-builder:resource:field=test with random string is not associated with a field marker",
			fields: fields{
				Field: &randomString,
			},
			args: args{
				fromMarker: testFieldMarker,
			},
			want: false,
		},
		{
			name: "operator-builder:resource:field=test with random string is not associated with a collection field marker",
			fields: fields{
				CollectionField: &randomString,
			},
			args: args{
				fromMarker: testCollectionMarker,
			},
			want: false,
		},
		{
			name: "operator-builder:resource:field=test with nil is not associated with a field marker",
			fields: fields{
				Field: nil,
			},
			args: args{
				fromMarker: testFieldMarker,
			},
			want: false,
		},
		{
			name: "operator-builder:resource:field=test with nil is not associated with a collection field marker",
			fields: fields{
				CollectionField: nil,
			},
			args: args{
				fromMarker: testCollectionMarker,
			},
			want: false,
		},
		{
			name: "operator-builder:resource:collectionField=testCollection is associated with a field marker",
			fields: fields{
				CollectionField: testCollectionMarker.Name,
			},
			args: args{
				fromMarker: testCollectionMarker,
			},
			want: true,
		},
		{
			name: "operator-builder:resource:collectionField=test is associated with a field marker from a collection",
			fields: fields{
				CollectionField: testFieldMarkerOnCollection.Name,
			},
			args: args{
				fromMarker: testFieldMarkerOnCollection,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := &ResourceMarker{
				Field:           tt.fields.Field,
				CollectionField: tt.fields.CollectionField,
				Value:           tt.fields.Value,
				Include:         tt.fields.Include,
				includeCode:     tt.fields.includeCode,
				fieldMarker:     tt.fields.fieldMarker,
			}
			if got := rm.isAssociated(tt.args.fromMarker); got != tt.want {
				t.Errorf("ResourceMarker.isAssociated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceMarker_getFieldMarker(t *testing.T) {
	t.Parallel()

	fieldOne := "fieldOne"
	fieldTwo := "fieldTwo"
	fieldMissing := "missing"

	testMarkers := &MarkerCollection{
		CollectionFieldMarkers: []*CollectionFieldMarker{
			{
				Name: &fieldOne,
				Type: FieldString,
			},
			{
				Name: &fieldTwo,
				Type: FieldString,
			},
		},
		FieldMarkers: []*FieldMarker{
			{
				Name:          &fieldOne,
				Type:          FieldString,
				forCollection: false,
			},
			{
				Name:          &fieldOne,
				Type:          FieldString,
				forCollection: true,
			},
			{
				Name:          &fieldTwo,
				Type:          FieldString,
				forCollection: false,
			},
		},
	}

	type fields struct {
		Field           *string
		CollectionField *string
		Value           interface{}
		Include         *bool
		fieldMarker     FieldMarkerProcessor
	}

	type args struct {
		markers *MarkerCollection
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   FieldMarkerProcessor
	}{
		{
			name: "resource marker with field within its component returns field marker",
			args: args{
				markers: testMarkers,
			},
			fields: fields{
				Field: &fieldOne,
			},
			want: &FieldMarker{
				Name: &fieldOne,
				Type: FieldString,
			},
		},
		{
			name: "resource marker with collection field returns collection field marker",
			args: args{
				markers: testMarkers,
			},
			fields: fields{
				CollectionField: &fieldTwo,
			},
			want: &CollectionFieldMarker{
				Name: &fieldTwo,
				Type: FieldString,
			},
		},
		{
			name: "resource marker with field within its component returns second field marker",
			args: args{
				markers: testMarkers,
			},
			fields: fields{
				Field: &fieldTwo,
			},
			want: &FieldMarker{
				Name: &fieldTwo,
				Type: FieldString,
			},
		},
		{
			name: "resource marker with collection field request returns field marker from collection",
			args: args{
				markers: testMarkers,
			},
			fields: fields{
				CollectionField: &fieldOne,
			},
			want: &FieldMarker{
				Name:          &fieldOne,
				Type:          FieldString,
				forCollection: true,
			},
		},
		{
			name: "resource marker with missing field returns nil",
			args: args{
				markers: testMarkers,
			},
			fields: fields{
				Field: &fieldMissing,
			},
			want: nil,
		},
		{
			name: "resource marker with missing collection field returns nil",
			args: args{
				markers: testMarkers,
			},
			fields: fields{
				CollectionField: &fieldMissing,
			},
			want: nil,
		},
		{
			name: "marker collection without any markers returns nil",
			args: args{
				markers: &MarkerCollection{},
			},
			fields: fields{},
			want:   nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := &ResourceMarker{
				Field:           tt.fields.Field,
				CollectionField: tt.fields.CollectionField,
				Value:           tt.fields.Value,
				Include:         tt.fields.Include,
				fieldMarker:     tt.fields.fieldMarker,
			}
			got := rm.getFieldMarker(tt.args.markers)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResourceMarker_Process(t *testing.T) {
	t.Parallel()

	fieldOne := "field.process.test"
	fieldTwo := "field.two.process.test"
	fieldMissing := "field.missing"
	include := true

	testMarkers := &MarkerCollection{
		CollectionFieldMarkers: []*CollectionFieldMarker{
			{
				Name: &fieldOne,
				Type: FieldString,
			},
		},
		FieldMarkers: []*FieldMarker{
			{
				Name:          &fieldOne,
				Type:          FieldString,
				forCollection: true,
			},
			{
				Name:          &fieldOne,
				Type:          FieldString,
				forCollection: false,
			},
			{
				Name:          &fieldTwo,
				Type:          FieldString,
				forCollection: true,
			},
		},
	}

	type fields struct {
		Field           *string
		CollectionField *string
		Value           interface{}
		Include         *bool
		includeCode     string
		fieldMarker     FieldMarkerProcessor
	}

	type args struct {
		markers *MarkerCollection
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    *ResourceMarker
	}{
		{
			name: "ensure valid marker returns no errors during processing",
			fields: fields{
				Field:   &fieldOne,
				Value:   "this.is.super.valid",
				Include: &include,
			},
			args: args{
				markers: testMarkers,
			},
			wantErr: false,
		},
		{
			name: "ensure missing marker returns error setting field marker",
			fields: fields{
				Field:   &fieldOne,
				Value:   []string{"thisisinvalid"},
				Include: &include,
			},
			args: args{
				markers: testMarkers,
			},
			wantErr: true,
		},
		{
			name:   "ensure invalid marker returns error on validation",
			fields: fields{},
			args: args{
				markers: testMarkers,
			},
			wantErr: true,
		},
		{
			name: "ensure missing marker returns error setting field marker",
			fields: fields{
				Field:   &fieldMissing,
				Value:   "testValue",
				Include: &include,
			},
			args: args{
				markers: testMarkers,
			},
			wantErr: true,
		},
		{
			name: "ensure missing marker returns error setting source code",
			fields: fields{
				Field:   &fieldOne,
				Value:   []string{"thisisinvalid"},
				Include: &include,
			},
			args: args{
				markers: testMarkers,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := &ResourceMarker{
				Field:           tt.fields.Field,
				CollectionField: tt.fields.CollectionField,
				Value:           tt.fields.Value,
				Include:         tt.fields.Include,
				includeCode:     tt.fields.includeCode,
				fieldMarker:     tt.fields.fieldMarker,
			}
			if err := rm.Process(tt.args.markers); (err != nil) != tt.wantErr {
				t.Errorf("ResourceMarker.Process() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil {
				assert.Equal(t, tt.want, rm)
			}
		})
	}
}

func TestResourceMarker_setSourceCode(t *testing.T) {
	t.Parallel()

	includeTrue := true
	includeFalse := false
	testSourceCodeField := "test.nested.field"

	testCollectionMarker := &CollectionFieldMarker{
		Name: &testSourceCodeField,
		Type: FieldString,
	}

	testFieldMarker := &FieldMarker{
		Name: &testSourceCodeField,
		Type: FieldInt,
	}

	testInvalidMarker := &FieldMarker{
		Name: &testSourceCodeField,
		Type: FieldUnknownType,
	}

	type fields struct {
		Field           *string
		CollectionField *string
		Value           interface{}
		Include         *bool
		includeCode     string
		fieldMarker     FieldMarkerProcessor
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ensure valid field marker produces no error on include",
			fields: fields{
				fieldMarker: testFieldMarker,
				Include:     &includeTrue,
				Value:       1,
			},
			wantErr: false,
		},
		{
			name: "ensure valid field marker produces no error on exclude",
			fields: fields{
				fieldMarker: testFieldMarker,
				Include:     &includeFalse,
				Value:       0,
			},
			wantErr: false,
		},
		{
			name: "ensure valid collection marker produces no error on include",
			fields: fields{
				fieldMarker: testCollectionMarker,
				Include:     &includeTrue,
				Value:       "testInclude",
			},
			wantErr: false,
		},
		{
			name: "ensure valid collection marker produces no error on exclude",
			fields: fields{
				fieldMarker: testCollectionMarker,
				Include:     &includeFalse,
				Value:       "testExclude",
			},
			wantErr: false,
		},
		{
			name: "ensure invalid marker with mismatched types produces error",
			fields: fields{
				fieldMarker: testFieldMarker,
				Include:     &includeTrue,
				Value:       "testMismatch",
			},
			wantErr: true,
		},
		{
			name: "ensure invalid marker with unknown field marker type produces error",
			fields: fields{
				fieldMarker: testInvalidMarker,
				Include:     &includeTrue,
				Value:       1,
			},
			wantErr: true,
		},
		{
			name: "ensure invalid marker with unknown resource marker value type produces error",
			fields: fields{
				fieldMarker: testFieldMarker,
				Include:     &includeTrue,
				Value:       []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := &ResourceMarker{
				Field:           tt.fields.Field,
				CollectionField: tt.fields.CollectionField,
				Value:           tt.fields.Value,
				Include:         tt.fields.Include,
				includeCode:     tt.fields.includeCode,
				fieldMarker:     tt.fields.fieldMarker,
			}
			if err := rm.setSourceCode(); (err != nil) != tt.wantErr {
				t.Errorf("ResourceMarker.setSourceCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
