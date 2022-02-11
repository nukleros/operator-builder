// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getResourceDefinitionVar(t *testing.T) {
	t.Parallel()

	type args struct {
		path                string
		forCollectionMarker bool
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "child resource with collection field should refer to collection",
			args: args{
				path:                "test.path",
				forCollectionMarker: true,
			},
			want: "collection.Spec.Test.Path",
		},
		{
			name: "child resource with non-collection field should refer to parent",
			args: args{
				path:                "test.path",
				forCollectionMarker: false,
			},
			want: "parent.Spec.Test.Path",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getResourceDefinitionVar(tt.args.path, tt.args.forCollectionMarker); got != tt.want {
				t.Errorf("getResourceDefinitionVar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceMarker_setSourceCodeVar(t *testing.T) {
	t.Parallel()

	testPath := "test.set.source.code.var"

	type fields struct {
		Field           *string
		CollectionField *string
		sourceCodeVar   string
	}

	tests := []struct {
		name   string
		fields fields
		want   *ResourceMarker
	}{
		{
			name: "resource marker referencing non-collection field",
			fields: fields{
				Field: &testPath,
			},
			want: &ResourceMarker{
				Field:         &testPath,
				sourceCodeVar: "parent.Spec.Test.Set.Source.Code.Var",
			},
		},
		{
			name: "resource marker referencing collection field",
			fields: fields{
				CollectionField: &testPath,
			},
			want: &ResourceMarker{
				CollectionField: &testPath,
				sourceCodeVar:   "collection.Spec.Test.Set.Source.Code.Var",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := &ResourceMarker{
				Field:           tt.fields.Field,
				CollectionField: tt.fields.CollectionField,
				sourceCodeVar:   tt.fields.sourceCodeVar,
			}
			rm.setSourceCodeVar()
			assert.Equal(t, tt.want, rm)
		})
	}
}

func TestResourceMarker_hasField(t *testing.T) {
	t.Parallel()

	testPath := "test.has.field"
	testEmpty := ""

	type fields struct {
		Field           *string
		CollectionField *string
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "resource marker with field returns true",
			fields: fields{
				Field: &testPath,
			},
			want: true,
		},
		{
			name: "resource marker with collection field returns true",
			fields: fields{
				CollectionField: &testPath,
			},
			want: true,
		},
		{
			name: "resource marker with empty field and collection field returns false",
			fields: fields{
				Field:           &testEmpty,
				CollectionField: &testEmpty,
			},
			want: false,
		},
		{
			name:   "resource marker without field or collection field returns false",
			fields: fields{},
			want:   false,
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
			if got := rm.hasField(); got != tt.want {
				t.Errorf("ResourceMarker.hasField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceMarker_hasValue(t *testing.T) {
	t.Parallel()

	testValue := "test.has.value"

	type fields struct {
		Value interface{}
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "resource marker with nil value returns false",
			fields: fields{
				Value: nil,
			},
			want: false,
		},
		{
			name: "resource marker with value returns true",
			fields: fields{
				Value: &testValue,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := &ResourceMarker{
				Value: tt.fields.Value,
			}
			if got := rm.hasValue(); got != tt.want {
				t.Errorf("ResourceMarker.hasValue() = %v, want %v", got, tt.want)
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
		sourceCodeVar   string
		sourceCodeValue string
		fieldMarker     interface{}
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "nil include value produces error",
			fields: fields{
				Field:       &testField,
				Value:       &testValue,
				fieldMarker: "",
			},
			wantErr: true,
		},
		{
			name: "missing field produces error",
			fields: fields{
				Value:       &testValue,
				Include:     &testInclude,
				fieldMarker: "",
			},
			wantErr: true,
		},
		{
			name: "missing value produces error",
			fields: fields{
				Field:       &testField,
				Include:     &testInclude,
				fieldMarker: "",
			},
			wantErr: true,
		},
		{
			name: "missing field marker produces error",
			fields: fields{
				Field:   &testField,
				Value:   &testValue,
				Include: &testInclude,
			},
			wantErr: true,
		},
		{
			name: "valid resource marker produces no error",
			fields: fields{
				Field:       &testField,
				Value:       &testValue,
				Include:     &testInclude,
				fieldMarker: "",
			},
			wantErr: false,
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
				sourceCodeVar:   tt.fields.sourceCodeVar,
				sourceCodeValue: tt.fields.sourceCodeValue,
				fieldMarker:     tt.fields.fieldMarker,
			}
			if err := rm.validate(); (err != nil) != tt.wantErr {
				t.Errorf("ResourceMarker.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceMarker_associateFieldMarker(t *testing.T) {
	t.Parallel()

	testMissing := "missing"
	collectionFieldOne := "collectionFieldOne"
	collectionFieldTwo := "collectionFieldTwo"
	fieldOne := "fieldOne"
	fieldTwo := "fieldTwo"

	testMarkers := &markerCollection{
		collectionFieldMarkers: []*CollectionFieldMarker{
			{
				Name: collectionFieldOne,
				Type: FieldString,
			},
			{
				Name: collectionFieldTwo,
				Type: FieldString,
			},
		},
		fieldMarkers: []*FieldMarker{
			{
				Name: fieldOne,
				Type: FieldString,
			},
			{
				Name: fieldTwo,
				Type: FieldString,
			},
		},
	}

	type fields struct {
		Field           *string
		CollectionField *string
		Value           interface{}
		Include         *bool
		sourceCodeVar   string
		sourceCodeValue string
		fieldMarker     interface{}
	}

	type args struct {
		markers *markerCollection
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ResourceMarker
	}{
		{
			name: "resource marker with non-collection field first",
			args: args{
				markers: testMarkers,
			},
			fields: fields{
				Field: &fieldOne,
			},
			want: &ResourceMarker{
				Field: &fieldOne,
				fieldMarker: &FieldMarker{
					Name: fieldOne,
					Type: FieldString,
				},
			},
		},
		{
			name: "resource marker with non-collection field second",
			args: args{
				markers: testMarkers,
			},
			fields: fields{
				Field: &fieldTwo,
			},
			want: &ResourceMarker{
				Field: &fieldTwo,
				fieldMarker: &FieldMarker{
					Name: fieldTwo,
					Type: FieldString,
				},
			},
		},
		{
			name: "resource marker with collection field first",
			args: args{
				markers: testMarkers,
			},
			fields: fields{
				CollectionField: &collectionFieldOne,
			},
			want: &ResourceMarker{
				CollectionField: &collectionFieldOne,
				fieldMarker: &CollectionFieldMarker{
					Name: collectionFieldOne,
					Type: FieldString,
				},
			},
		},
		{
			name: "resource marker with collection field second",
			args: args{
				markers: testMarkers,
			},
			fields: fields{
				CollectionField: &collectionFieldTwo,
			},
			want: &ResourceMarker{
				CollectionField: &collectionFieldTwo,
				fieldMarker: &CollectionFieldMarker{
					Name: collectionFieldTwo,
					Type: FieldString,
				},
			},
		},
		{
			name: "resource marker with no related fields",
			args: args{
				markers: testMarkers,
			},
			fields: fields{
				CollectionField: &testMissing,
				Field:           &testMissing,
			},
			want: &ResourceMarker{
				CollectionField: &testMissing,
				Field:           &testMissing,
			},
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
				sourceCodeVar:   tt.fields.sourceCodeVar,
				sourceCodeValue: tt.fields.sourceCodeValue,
				fieldMarker:     tt.fields.fieldMarker,
			}
			rm.associateFieldMarker(tt.args.markers)
			assert.Equal(t, tt.want, rm)
		})
	}
}
