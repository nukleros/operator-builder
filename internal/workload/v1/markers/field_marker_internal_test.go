// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package markers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nukleros/operator-builder/internal/markers/marker"
)

func TestFieldMarker_String(t *testing.T) {
	t.Parallel()

	testName := "fmtest"
	testString := "fm test"
	testBool := false

	type fields struct {
		Name        *string
		Type        FieldType
		Description *string
		Default     interface{}
		Arbitrary   *bool
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure field string output matches expected",
			fields: fields{
				Name:        &testName,
				Type:        FieldString,
				Description: &testString,
				Default:     testName,
				Arbitrary:   &testBool,
			},
			want: "FieldMarker{Name: fmtest Type: string Description: \"fm test\" Default: fmtest Arbitrary: false}",
		},
		{
			name: "ensure field with nil values output matches expected",
			fields: fields{
				Name:        &testName,
				Type:        FieldString,
				Description: nil,
				Default:     testName,
				Arbitrary:   nil,
			},
			want: "FieldMarker{Name: fmtest Type: string Description: \"\" Default: fmtest Arbitrary: false}",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := FieldMarker{
				Name:        tt.fields.Name,
				Type:        tt.fields.Type,
				Description: tt.fields.Description,
				Default:     tt.fields.Default,
			}
			if got := fm.String(); got != tt.want {
				t.Errorf("FieldMarker.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defineFieldMarker(t *testing.T) {
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
			name: "ensure valid registry can properly add a field marker",
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
			if err := defineFieldMarker(tt.args.registry); (err != nil) != tt.wantErr {
				t.Errorf("defineFieldMarker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFieldMarker_GetDefault(t *testing.T) {
	t.Parallel()

	fmDefault := "this is a field default value"

	type fields struct {
		Default interface{}
	}

	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		{
			name: "ensure field default returns as expected",
			fields: fields{
				Default: fmDefault,
			},
			want: fmDefault,
		},
		{
			name: "ensure field default with nil value returns as expected",
			fields: fields{
				Default: nil,
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := &FieldMarker{
				Default: tt.fields.Default,
			}
			if got := fm.GetDefault(); got != tt.want {
				t.Errorf("FieldMarker.GetDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldMarker_GetName(t *testing.T) {
	t.Parallel()

	name := "testGetFmName"
	emptyName := ""

	type fields struct {
		Name *string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure field name returns as expected",
			fields: fields{
				Name: &name,
			},
			want: name,
		},
		{
			name: "ensure field name with empty value returns as expected",
			fields: fields{
				Name: &emptyName,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := &FieldMarker{
				Name: tt.fields.Name,
			}
			if got := fm.GetName(); got != tt.want {
				t.Errorf("FieldMarker.GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldMarker_GetDescription(t *testing.T) {
	t.Parallel()

	fmDescription := "test description"

	type fields struct {
		Description *string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure field description returns as expected",
			fields: fields{
				Description: &fmDescription,
			},
			want: "test description",
		},
		{
			name: "ensure field description with nil value returns as expected",
			fields: fields{
				Description: nil,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := &FieldMarker{
				Description: tt.fields.Description,
			}
			if got := fm.GetDescription(); got != tt.want {
				t.Errorf("FieldMarker.GetDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldMarker_GetFieldType(t *testing.T) {
	t.Parallel()

	type fields struct {
		Type FieldType
	}

	tests := []struct {
		name   string
		fields fields
		want   FieldType
	}{
		{
			name: "ensure field type string returns as expected",
			fields: fields{
				Type: FieldString,
			},
			want: FieldString,
		},
		{
			name: "ensure field type struct returns as expected",
			fields: fields{
				Type: FieldStruct,
			},
			want: FieldStruct,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := &FieldMarker{
				Type: tt.fields.Type,
			}
			if got := fm.GetFieldType(); got != tt.want {
				t.Errorf("FieldMarker.GetFieldType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldMarker_GetReplaceText(t *testing.T) {
	t.Parallel()

	fmReplace := "test replace"

	type fields struct {
		Replace *string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure field replace text returns as expected",
			fields: fields{
				Replace: &fmReplace,
			},
			want: fmReplace,
		},
		{
			name: "ensure field replace text with empty value returns as expected",
			fields: fields{
				Replace: nil,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := &FieldMarker{
				Replace: tt.fields.Replace,
			}
			if got := fm.GetReplaceText(); got != tt.want {
				t.Errorf("FieldMarker.GetReplaceText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldMarker_GetSpecPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
	}{
		{
			name: "ensure field returns correct spec prefix",
			want: FieldSpecPrefix,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := &FieldMarker{}
			if got := fm.GetSpecPrefix(); got != tt.want {
				t.Errorf("FieldMarker.GetSpecPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldMarker_GetOriginalValue(t *testing.T) {
	t.Parallel()

	name := "fmOriginalName"

	type fields struct {
		originalValue interface{}
	}

	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		{
			name: "ensure field original value string returns as expected",
			fields: fields{
				originalValue: name,
			},
			want: name,
		},
		{
			name: "ensure field original value integer returns as expected",
			fields: fields{
				originalValue: 1,
			},
			want: 1,
		},
		{
			name: "ensure field original value negative integer returns as expected",
			fields: fields{
				originalValue: -1,
			},
			want: -1,
		},
		{
			name: "ensure field original value nil returns as expected",
			fields: fields{
				originalValue: nil,
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := &FieldMarker{
				originalValue: tt.fields.originalValue,
			}
			if got := fm.GetOriginalValue(); got != tt.want {
				t.Errorf("FieldMarker.GetOriginalValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldMarker_GetSourceCodeVariable(t *testing.T) {
	t.Parallel()

	type fields struct {
		sourceCodeVar string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "source code variable for field marker returns correctly",
			fields: fields{
				sourceCodeVar: "this.Is.A.Test",
			},
			want: "this.Is.A.Test",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := &FieldMarker{
				sourceCodeVar: tt.fields.sourceCodeVar,
			}
			if got := fm.GetSourceCodeVariable(); got != tt.want {
				t.Errorf("FieldMarker.GetSourceCodeVariable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldMarker_IsCollectionFieldMarker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{
			name: "ensure a field marker is never a collection field marker",
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := &FieldMarker{}
			if got := fm.IsCollectionFieldMarker(); got != tt.want {
				t.Errorf("FieldMarker.IsCollectionFieldMarker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldMarker_IsFieldMarker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{
			name: "ensure a field marker is always a field marker",
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := &FieldMarker{}
			if got := fm.IsFieldMarker(); got != tt.want {
				t.Errorf("FieldMarker.IsFieldMarker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldMarker_IsForCollection(t *testing.T) {
	t.Parallel()

	type fields struct {
		forCollection bool
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "ensure field for collection returns true",
			fields: fields{
				forCollection: true,
			},
			want: true,
		},
		{
			name: "ensure field for non-collection returns false",
			fields: fields{
				forCollection: false,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := &FieldMarker{
				forCollection: tt.fields.forCollection,
			}
			if got := fm.IsForCollection(); got != tt.want {
				t.Errorf("FieldMarker.IsForCollection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldMarker_SetOriginalValue(t *testing.T) {
	t.Parallel()

	fmOriginalValue := "testUpdateOriginal"
	fmFake := "testFmFake"

	type fields struct {
		originalValue interface{}
	}

	type args struct {
		value string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *FieldMarker
	}{
		{
			name: "ensure field original value with string already set is set as expected",
			fields: fields{
				originalValue: fmFake,
			},
			args: args{
				value: fmOriginalValue,
			},
			want: &FieldMarker{
				originalValue: &fmOriginalValue,
			},
		},
		{
			name: "ensure field original value wihtout string already set is set as expected",
			args: args{
				value: fmOriginalValue,
			},
			want: &FieldMarker{
				originalValue: &fmOriginalValue,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := &FieldMarker{
				originalValue: tt.fields.originalValue,
			}
			fm.SetOriginalValue(tt.args.value)
			assert.Equal(t, tt.want, fm)
		})
	}
}

func TestFieldMarker_SetDescription(t *testing.T) {
	t.Parallel()

	fmSetDescription := "testUpdate"
	fmSetDescriptionExist := "testExist"

	type fields struct {
		Description *string
	}

	type args struct {
		description string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *FieldMarker
	}{
		{
			name: "ensure field description with string already set is set as expected",
			fields: fields{
				Description: &fmSetDescriptionExist,
			},
			args: args{
				description: fmSetDescription,
			},
			want: &FieldMarker{
				Description: &fmSetDescription,
			},
		},
		{
			name: "ensure field description wihtout string already set is set as expected",
			args: args{
				description: fmSetDescription,
			},
			want: &FieldMarker{
				Description: &fmSetDescription,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := &FieldMarker{
				Description: tt.fields.Description,
			}
			fm.SetDescription(tt.args.description)
			assert.Equal(t, tt.want, fm)
		})
	}
}

func TestFieldMarker_SetForCollection(t *testing.T) {
	t.Parallel()

	type fields struct {
		forCollection bool
	}

	type args struct {
		forCollection bool
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *FieldMarker
	}{
		{
			name: "ensure field for collection already set is set as expected (true > false)",
			fields: fields{
				forCollection: true,
			},
			args: args{
				forCollection: false,
			},
			want: &FieldMarker{
				forCollection: false,
			},
		},
		{
			name: "ensure field for collection already set is set as expected (false > true)",
			fields: fields{
				forCollection: false,
			},
			args: args{
				forCollection: true,
			},
			want: &FieldMarker{
				forCollection: true,
			},
		},
		{
			name: "ensure field for collection wihtout already set is set as expected",
			args: args{
				forCollection: true,
			},
			want: &FieldMarker{
				forCollection: true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm := &FieldMarker{
				forCollection: tt.fields.forCollection,
			}
			fm.SetForCollection(tt.args.forCollection)
			assert.Equal(t, tt.want, fm)
		})
	}
}
