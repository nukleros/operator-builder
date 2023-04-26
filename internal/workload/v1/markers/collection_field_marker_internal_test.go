// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package markers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nukleros/operator-builder/internal/markers/marker"
)

func TestCollectionFieldMarker_String(t *testing.T) {
	t.Parallel()

	testString := "cfm test"
	testName := "cfmtest"

	type fields struct {
		Name        *string
		Type        FieldType
		Description *string
		Default     interface{}
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure collection field string output matches expected",
			fields: fields{
				Name:        &testName,
				Type:        FieldString,
				Description: &testString,
				Default:     testName,
			},
			want: "CollectionFieldMarker{Name: cfmtest Type: string Description: \"cfm test\" Default: cfmtest}",
		},
		{
			name: "ensure collection field with nil values output matches expected",
			fields: fields{
				Name:        &testName,
				Type:        FieldString,
				Description: nil,
				Default:     testName,
			},
			want: "CollectionFieldMarker{Name: cfmtest Type: string Description: \"\" Default: cfmtest}",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfm := CollectionFieldMarker{
				Name:        tt.fields.Name,
				Type:        tt.fields.Type,
				Description: tt.fields.Description,
				Default:     tt.fields.Default,
			}
			if got := cfm.String(); got != tt.want {
				t.Errorf("CollectionFieldMarker.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defineCollectionFieldMarker(t *testing.T) {
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
			name: "ensure valid registry can properly add a collection field marker",
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
			if err := defineCollectionFieldMarker(tt.args.registry); (err != nil) != tt.wantErr {
				t.Errorf("defineCollectionFieldMarker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCollectionFieldMarker_GetDefault(t *testing.T) {
	t.Parallel()

	cfmDefault := "this is a collection default value"

	type fields struct {
		Default interface{}
	}

	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		{
			name: "ensure collection field default returns as expected",
			fields: fields{
				Default: cfmDefault,
			},
			want: cfmDefault,
		},
		{
			name: "ensure collection field default with nil value returns as expected",
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
			cfm := &CollectionFieldMarker{
				Default: tt.fields.Default,
			}
			if got := cfm.GetDefault(); got != tt.want {
				t.Errorf("CollectionFieldMarker.GetDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectionFieldMarker_GetName(t *testing.T) {
	t.Parallel()

	name := "getNameTest"
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
			name: "ensure collection field name returns as expected",
			fields: fields{
				Name: &name,
			},
			want: name,
		},
		{
			name: "ensure collection field name with empty value returns as expected",
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
			cfm := &CollectionFieldMarker{
				Name: tt.fields.Name,
			}
			if got := cfm.GetName(); got != tt.want {
				t.Errorf("CollectionFieldMarker.GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectionFieldMarker_GetDescription(t *testing.T) {
	t.Parallel()

	cfmDescription := "test collection description"

	type fields struct {
		Description *string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure collection field description returns as expected",
			fields: fields{
				Description: &cfmDescription,
			},
			want: cfmDescription,
		},
		{
			name: "ensure collection field description with nil value returns as expected",
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
			cfm := &CollectionFieldMarker{
				Description: tt.fields.Description,
			}
			if got := cfm.GetDescription(); got != tt.want {
				t.Errorf("CollectionFieldMarker.GetDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectionFieldMarker_GetFieldType(t *testing.T) {
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
			name: "ensure collection field type string returns as expected",
			fields: fields{
				Type: FieldString,
			},
			want: FieldString,
		},
		{
			name: "ensure collection field type struct returns as expected",
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
			cfm := &CollectionFieldMarker{
				Type: tt.fields.Type,
			}
			if got := cfm.GetFieldType(); got != tt.want {
				t.Errorf("CollectionFieldMarker.GetFieldType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectionFieldMarker_GetReplaceText(t *testing.T) {
	t.Parallel()

	cfmReplace := "test collection replace"

	type fields struct {
		Replace *string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure collection field replace text returns as expected",
			fields: fields{
				Replace: &cfmReplace,
			},
			want: cfmReplace,
		},
		{
			name: "ensure collection field replace text with empty value returns as expected",
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
			cfm := &CollectionFieldMarker{
				Replace: tt.fields.Replace,
			}
			if got := cfm.GetReplaceText(); got != tt.want {
				t.Errorf("CollectionFieldMarker.GetReplaceText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectionFieldMarker_GetSpecPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
	}{
		{
			name: "ensure collection field returns correct spec prefix",
			want: CollectionFieldSpecPrefix,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfm := &CollectionFieldMarker{}
			if got := cfm.GetSpecPrefix(); got != tt.want {
				t.Errorf("CollectionFieldMarker.GetSpecPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectionFieldMarker_GetOriginalValue(t *testing.T) {
	t.Parallel()

	originalValue := "originalValueTest"

	type fields struct {
		originalValue interface{}
	}

	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		{
			name: "ensure collection field original value string returns as expected",
			fields: fields{
				originalValue: originalValue,
			},
			want: originalValue,
		},
		{
			name: "ensure collection field original value integer returns as expected",
			fields: fields{
				originalValue: 1,
			},
			want: 1,
		},
		{
			name: "ensure collection field original value negative integer returns as expected",
			fields: fields{
				originalValue: -1,
			},
			want: -1,
		},
		{
			name: "ensure collection field original value nil returns as expected",
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
			cfm := &CollectionFieldMarker{
				originalValue: tt.fields.originalValue,
			}
			if got := cfm.GetOriginalValue(); got != tt.want {
				t.Errorf("CollectionFieldMarker.GetOriginalValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectionFieldMarker_GetSourceCodeVariable(t *testing.T) {
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
			cfm := &CollectionFieldMarker{
				sourceCodeVar: tt.fields.sourceCodeVar,
			}
			if got := cfm.GetSourceCodeVariable(); got != tt.want {
				t.Errorf("CollectionFieldMarker.GetSourceCodeVariable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectionFieldMarker_IsCollectionFieldMarker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{
			name: "ensure a collection field marker is always a collection field marker",
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfm := &CollectionFieldMarker{}
			if got := cfm.IsCollectionFieldMarker(); got != tt.want {
				t.Errorf("CollectionFieldMarker.IsCollectionFieldMarker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectionFieldMarker_IsFieldMarker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{
			name: "ensure a collection field marker is never a field marker",
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfm := &CollectionFieldMarker{}
			if got := cfm.IsFieldMarker(); got != tt.want {
				t.Errorf("CollectionFieldMarker.IsFieldMarker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectionFieldMarker_IsForCollection(t *testing.T) {
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
			// this is technically an invalid test case as collection field markers on collections
			// are automatically converted to field markers.  test this anyway.
			name: "ensure collection field for collection returns true",
			fields: fields{
				forCollection: true,
			},
			want: true,
		},
		{
			name: "ensure collection field for non-collection returns false",
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
			cfm := &CollectionFieldMarker{
				forCollection: tt.fields.forCollection,
			}
			if got := cfm.IsForCollection(); got != tt.want {
				t.Errorf("CollectionFieldMarker.IsForCollection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectionFieldMarker_SetOriginalValue(t *testing.T) {
	t.Parallel()

	cfmOriginalValue := "testCollectionUpdateOriginal"

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
		want   *CollectionFieldMarker
	}{
		{
			name: "ensure collection field original value with string already set is set as expected",
			fields: fields{
				originalValue: "test",
			},
			args: args{
				value: cfmOriginalValue,
			},
			want: &CollectionFieldMarker{
				originalValue: &cfmOriginalValue,
			},
		},
		{
			name: "ensure collection field original value wihtout string already set is set as expected",
			args: args{
				value: cfmOriginalValue,
			},
			want: &CollectionFieldMarker{
				originalValue: &cfmOriginalValue,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfm := &CollectionFieldMarker{
				originalValue: tt.fields.originalValue,
			}
			cfm.SetOriginalValue(tt.args.value)
			assert.Equal(t, tt.want, cfm)
		})
	}
}

func TestCollectionFieldMarker_SetDescription(t *testing.T) {
	t.Parallel()

	cfmSetDescription := "testCollectionUpdate"
	cfmSetDescriptionExist := "testCollectionExist"

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
		want   *CollectionFieldMarker
	}{
		{
			name: "ensure collection field description with string already set is set as expected",
			fields: fields{
				Description: &cfmSetDescriptionExist,
			},
			args: args{
				description: cfmSetDescription,
			},
			want: &CollectionFieldMarker{
				Description: &cfmSetDescription,
			},
		},
		{
			name: "ensure collection field description wihtout string already set is set as expected",
			args: args{
				description: cfmSetDescription,
			},
			want: &CollectionFieldMarker{
				Description: &cfmSetDescription,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfm := &CollectionFieldMarker{
				Description: tt.fields.Description,
			}
			cfm.SetDescription(tt.args.description)
			assert.Equal(t, tt.want, cfm)
		})
	}
}

func TestCollectionFieldMarker_SetForCollection(t *testing.T) {
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
		want   *CollectionFieldMarker
	}{
		{
			name: "ensure collection field for collection already set is set as expected (true > false)",
			fields: fields{
				forCollection: true,
			},
			args: args{
				forCollection: false,
			},
			want: &CollectionFieldMarker{
				forCollection: false,
			},
		},
		{
			name: "ensure collection field for collection already set is set as expected (false > true)",
			fields: fields{
				forCollection: false,
			},
			args: args{
				forCollection: true,
			},
			want: &CollectionFieldMarker{
				forCollection: true,
			},
		},
		{
			name: "ensure collection field for collection wihtout already set is set as expected",
			args: args{
				forCollection: true,
			},
			want: &CollectionFieldMarker{
				forCollection: true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfm := &CollectionFieldMarker{
				forCollection: tt.fields.forCollection,
			}
			cfm.SetForCollection(tt.args.forCollection)
			assert.Equal(t, tt.want, cfm)
		})
	}
}
