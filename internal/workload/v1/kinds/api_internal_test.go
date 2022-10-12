// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package kinds

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nukleros/operator-builder/internal/workload/v1/markers"
)

func TestAPIFields_GenerateSampleSpec(t *testing.T) {
	t.Parallel()

	type fields struct {
		Name         string
		manifestName string
		Type         markers.FieldType
		Tags         string
		Comments     []string
		Markers      []string
		Children     []*APIFields
		Default      string
		Sample       string
	}

	tests := []struct {
		name         string
		fields       fields
		requiredOnly bool
		want         string
	}{
		{
			name: "test generation",
			fields: fields{
				Sample: "spec:",
				Children: []*APIFields{
					{
						Sample: "test: content",
					},
				},
			},
			want: "spec:\n  test: content\n",
		},
		{
			name: "test nested generation",
			fields: fields{
				Sample: "spec:",
				Children: []*APIFields{
					{
						Sample: "test:",
						Children: []*APIFields{
							{
								Sample: "levelTwo:",
								Children: []*APIFields{
									{
										Sample: "hello: world",
									},
								},
							},
						},
					},
					{
						Sample: "levelOne: hello",
					},
				},
			},
			want: "spec:\n  test:\n    levelTwo:\n      hello: world\n  levelOne: hello\n",
		},
		{
			name: "test required only generation",
			fields: fields{
				Sample: "spec:",
				Children: []*APIFields{
					{
						Sample: "test: content",
					},
					{
						Sample:  "test2: content2",
						Default: "defaultValue",
					},
				},
			},
			requiredOnly: true,
			want:         "spec:\n  test: content\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := &APIFields{
				Name:         tt.fields.Name,
				manifestName: tt.fields.manifestName,
				Type:         tt.fields.Type,
				Tags:         tt.fields.Tags,
				Comments:     tt.fields.Comments,
				Markers:      tt.fields.Markers,
				Children:     tt.fields.Children,
				Default:      tt.fields.Default,
				Sample:       tt.fields.Sample,
			}
			if got := api.GenerateSampleSpec(tt.requiredOnly); got != tt.want {
				t.Errorf("CRDFields.GenerateSampleSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIFields_generateStructName(t *testing.T) {
	t.Parallel()

	type args struct {
		manifestName string
		path         string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "single nest name generation",
			args: args{
				manifestName: "webStore",
				path:         "webStore.image",
			},
			want: "SpecWebStore",
		},
		{
			name: "multi nest name generation",
			args: args{
				manifestName: "tag",
				path:         "webStore.image.tag.extension",
			},
			want: "SpecWebStoreImageTag",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			crd := &APIFields{
				manifestName: tt.args.manifestName,
			}
			crd.generateStructName(tt.args.path)
			if got := crd.StructName; got != tt.want {
				t.Errorf("CRDFields.generateStructName(%v) = %v, want %v", tt.args.path, got, tt.want)
			}
		})
	}
}

func TestAPIFields_needsGenerate(t *testing.T) {
	t.Parallel()

	type args struct {
		requiredOnly bool
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "api field needs generation if without required fields only",
			args: args{
				requiredOnly: false,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := &APIFields{}
			if got := api.needsGenerate(tt.args.requiredOnly); got != tt.want {
				t.Errorf("APIFields.needsGenerate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIFields_hasRequiredField(t *testing.T) {
	t.Parallel()

	withRequired := &APIFields{
		Children: []*APIFields{},
		Default:  "",
	}

	withNotRequired := &APIFields{
		Children: []*APIFields{},
		Default:  "default",
	}

	type fields struct {
		Children []*APIFields
		Default  string
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "flat api field is a required field",
			fields: fields{
				Children: withRequired.Children,
				Default:  withRequired.Default,
			},
			want: true,
		},
		{
			name: "flat api field is not a required field",
			fields: fields{
				Children: withNotRequired.Children,
				Default:  withNotRequired.Default,
			},
			want: false,
		},
		{
			name: "nested api field has a required field",
			fields: fields{
				Children: []*APIFields{
					{
						Children: []*APIFields{
							withRequired,
						},
					},
					withRequired,
				},
				Default: "",
			},
			want: true,
		},
		{
			name: "nested api field does not have a required field",
			fields: fields{
				Children: []*APIFields{
					{
						Children: []*APIFields{
							withNotRequired,
						},
					},
					withRequired,
				},
				Default: "",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := &APIFields{
				Children: tt.fields.Children,
				Default:  tt.fields.Default,
			}
			if got := api.hasRequiredField(); got != tt.want {
				t.Errorf("APIFields.hasRequiredField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mustWrite(t *testing.T) {
	t.Parallel()

	type args struct {
		n   int
		err error
	}

	tests := []struct {
		name        string
		args        args
		shouldPanic bool
	}{
		{
			name: "must write panic",
			args: args{
				n:   -1,
				err: fmt.Errorf("test panic"), //nolint
			},
			shouldPanic: true,
		},
		{
			name: "must write success",
			args: args{
				n:   -1,
				err: nil,
			},
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic")
					}
				}()
			}
			mustWrite(tt.args.n, tt.args.err)
		})
	}
}

func TestAPIFields_getSampleValue(t *testing.T) {
	t.Parallel()

	testString := "testString"
	testInt := 1
	testBool := true

	type fields struct {
		Name         string
		StructName   string
		manifestName string
		Type         markers.FieldType
		Tags         string
		Comments     []string
		Markers      []string
		Children     []*APIFields
		Default      string
		Sample       string
		Last         bool
	}

	type args struct {
		sampleVal interface{}
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "test string value",
			args: args{
				sampleVal: testString,
			},
			fields: fields{
				Type: markers.FieldString,
			},
			want: fmt.Sprintf("%q", testString),
		},
		{
			name: "test pointer to string value",
			args: args{
				sampleVal: &testString,
			},
			fields: fields{
				Type: markers.FieldString,
			},
			want: fmt.Sprintf("%q", testString),
		},
		{
			name: "test int value",
			args: args{
				sampleVal: testInt,
			},
			fields: fields{
				Type: markers.FieldInt,
			},
			want: fmt.Sprintf("%v", testInt),
		},
		{
			name: "test pointer to int value",
			args: args{
				sampleVal: &testInt,
			},
			fields: fields{
				Type: markers.FieldInt,
			},
			want: fmt.Sprintf("%v", testInt),
		},
		{
			name: "test bool value",
			args: args{
				sampleVal: testBool,
			},
			fields: fields{
				Type: markers.FieldBool,
			},
			want: fmt.Sprintf("%v", testBool),
		},
		{
			name: "test pointer to bool value",
			args: args{
				sampleVal: &testBool,
			},
			fields: fields{
				Type: markers.FieldBool,
			},
			want: fmt.Sprintf("%v", testBool),
		},
		{
			name: "test other value",
			args: args{
				sampleVal: []string{"test", "get", "sample"},
			},
			fields: fields{
				Type: markers.FieldUnknownType,
			},
			want: "[test get sample]",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := &APIFields{
				Name:         tt.fields.Name,
				StructName:   tt.fields.StructName,
				manifestName: tt.fields.manifestName,
				Type:         tt.fields.Type,
				Tags:         tt.fields.Tags,
				Comments:     tt.fields.Comments,
				Markers:      tt.fields.Markers,
				Children:     tt.fields.Children,
				Default:      tt.fields.Default,
				Sample:       tt.fields.Sample,
			}
			if got := api.getSampleValue(tt.args.sampleVal); got != tt.want {
				t.Errorf("APIFields.getSampleValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIFields_setSample(t *testing.T) {
	t.Parallel()

	type fields struct {
		manifestName string
		Type         markers.FieldType
		Sample       string
	}

	type args struct {
		sampleVal interface{}
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		expect *APIFields
	}{
		{
			name: "set string sample",
			args: args{
				sampleVal: "string",
			},
			fields: fields{
				manifestName: "string",
				Type:         markers.FieldString,
			},
			expect: &APIFields{
				manifestName: "string",
				Type:         markers.FieldString,
				Sample:       "string: \"string\"",
			},
		},
		{
			name: "set struct sample",
			args: args{
				sampleVal: "struct",
			},
			fields: fields{
				manifestName: "struct",
				Type:         markers.FieldStruct,
			},
			expect: &APIFields{
				manifestName: "struct",
				Type:         markers.FieldStruct,
				Sample:       "struct:",
			},
		},
		{
			name: "set other sample",
			args: args{
				sampleVal: []string{"test", "sample"},
			},
			fields: fields{
				manifestName: "other",
			},
			expect: &APIFields{
				manifestName: "other",
				Type:         markers.FieldUnknownType,
				Sample:       "other: [test sample]",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := &APIFields{
				manifestName: tt.fields.manifestName,
				Type:         tt.fields.Type,
				Sample:       tt.fields.Sample,
			}
			api.setSample(tt.args.sampleVal)
			assert.Equal(t, tt.expect, api)
		})
	}
}

func TestAPIFields_setDefault(t *testing.T) {
	t.Parallel()

	type fields struct {
		manifestName string
		Type         markers.FieldType
		Markers      []string
		Default      string
		Sample       string
	}

	type args struct {
		sampleVal interface{}
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		expect *APIFields
	}{
		{
			name: "set default for string",
			args: args{
				sampleVal: "string",
			},
			fields: fields{
				manifestName: "string",
				Type:         markers.FieldString,
				Markers: []string{
					"marker1",
					"marker2",
				},
			},
			expect: &APIFields{
				manifestName: "string",
				Type:         markers.FieldString,
				Sample:       "string: \"string\"",
				Default:      "\"string\"",
				Markers: []string{
					"marker1",
					"marker2",
				},
			},
		},
		{
			name: "set default for other",
			args: args{
				sampleVal: []string{"other"},
			},
			fields: fields{
				manifestName: "other",
			},
			expect: &APIFields{
				manifestName: "other",
				Type:         markers.FieldUnknownType,
				Sample:       "other: [other]",
				Default:      "[other]",
				Markers: []string{
					"+kubebuilder:default=[other]",
					"+kubebuilder:validation:Optional",
					"(Default: [other])",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := &APIFields{
				manifestName: tt.fields.manifestName,
				Type:         tt.fields.Type,
				Markers:      tt.fields.Markers,
				Default:      tt.fields.Default,
				Sample:       tt.fields.Sample,
			}
			api.setDefault(tt.args.sampleVal)
			assert.Equal(t, tt.expect, api)
		})
	}
}

func TestAPIFields_setCommentsAndDefault(t *testing.T) {
	t.Parallel()

	type fields struct {
		manifestName string
		Type         markers.FieldType
		Comments     []string
		Markers      []string
		Default      string
		Sample       string
	}

	type args struct {
		comments   []string
		sampleVal  interface{}
		hasDefault bool
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		expect *APIFields
	}{
		{
			name: "set comments for string",
			args: args{
				sampleVal: "string",
				comments: []string{
					"comment3",
					"comment4",
				},
				hasDefault: true,
			},
			fields: fields{
				manifestName: "string",
				Type:         markers.FieldString,
				Comments: []string{
					"comment1",
					"comment2",
				},
			},
			expect: &APIFields{
				manifestName: "string",
				Type:         markers.FieldString,
				Sample:       "string: \"string\"",
				Default:      "\"string\"",
				Markers: []string{
					"+kubebuilder:default=\"string\"",
					"+kubebuilder:validation:Optional",
					"(Default: \"string\")",
				},
				Comments: []string{
					"comment3",
					"comment4",
				},
			},
		},
		{
			name: "set comments for other",
			args: args{
				sampleVal: "other",
			},
			fields: fields{
				manifestName: "other",
			},
			expect: &APIFields{
				manifestName: "other",
				Markers: []string{
					"+kubebuilder:validation:Required",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := &APIFields{
				Comments:     tt.fields.Comments,
				manifestName: tt.fields.manifestName,
				Type:         tt.fields.Type,
				Markers:      tt.fields.Markers,
				Default:      tt.fields.Default,
				Sample:       tt.fields.Sample,
			}
			api.setCommentsAndDefault(tt.args.comments, tt.args.sampleVal, tt.args.hasDefault)
			assert.Equal(t, tt.expect, api)
		})
	}
}

func TestAPIFields_newChild(t *testing.T) {
	t.Parallel()

	testString := "string"
	testInt := 1
	testBool := true

	type fields struct {
		Name         string
		StructName   string
		manifestName string
		Type         markers.FieldType
		Tags         string
		Comments     []string
		Markers      []string
		Children     []*APIFields
		Default      string
		Sample       string
	}

	type args struct {
		name      string
		fieldType markers.FieldType
		sample    interface{}
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *APIFields
	}{
		{
			name: "new child for string",
			args: args{
				name:      "string",
				fieldType: markers.FieldString,
				sample:    "string",
			},
			fields: fields{},
			want: &APIFields{
				Name:         "String",
				manifestName: "string",
				Type:         markers.FieldString,
				Sample:       "string: \"string\"",
				Tags:         "`json:\"string,omitempty\"`",
				Comments:     []string{},
				Markers:      []string{},
			},
		},
		{
			name: "new child for string pointer",
			args: args{
				name:      "string",
				fieldType: markers.FieldString,
				sample:    &testString,
			},
			fields: fields{},
			want: &APIFields{
				Name:         "String",
				manifestName: "string",
				Type:         markers.FieldString,
				Sample:       "string: \"string\"",
				Tags:         "`json:\"string,omitempty\"`",
				Comments:     []string{},
				Markers:      []string{},
			},
		},
		{
			name: "new child for unknown",
			args: args{
				name:      "unknown",
				fieldType: markers.FieldUnknownType,
				sample:    []string{"test", "unknown"},
			},
			fields: fields{},
			want: &APIFields{
				Name:         "Unknown",
				manifestName: "unknown",
				Type:         markers.FieldUnknownType,
				Sample:       "unknown: [test unknown]",
				Tags:         "`json:\"unknown,omitempty\"`",
				Comments:     []string{},
				Markers:      []string{},
			},
		},
		{
			name: "new child for int",
			args: args{
				name:      "int",
				fieldType: markers.FieldInt,
				sample:    1,
			},
			fields: fields{},
			want: &APIFields{
				Name:         "Int",
				manifestName: "int",
				Type:         markers.FieldInt,
				Sample:       "int: 1",
				Tags:         "`json:\"int,omitempty\"`",
				Comments:     []string{},
				Markers:      []string{},
			},
		},
		{
			name: "new child for int",
			args: args{
				name:      "int",
				fieldType: markers.FieldInt,
				sample:    &testInt,
			},
			fields: fields{},
			want: &APIFields{
				Name:         "Int",
				manifestName: "int",
				Type:         markers.FieldInt,
				Sample:       "int: 1",
				Tags:         "`json:\"int,omitempty\"`",
				Comments:     []string{},
				Markers:      []string{},
			},
		},
		{
			name: "new child for bool",
			args: args{
				name:      "bool",
				fieldType: markers.FieldBool,
				sample:    true,
			},
			fields: fields{},
			want: &APIFields{
				Name:         "Bool",
				manifestName: "bool",
				Type:         markers.FieldBool,
				Sample:       "bool: true",
				Tags:         "`json:\"bool,omitempty\"`",
				Comments:     []string{},
				Markers:      []string{},
			},
		},
		{
			name: "new child for bool pointer",
			args: args{
				name:      "bool",
				fieldType: markers.FieldBool,
				sample:    &testBool,
			},
			fields: fields{},
			want: &APIFields{
				Name:         "Bool",
				manifestName: "bool",
				Type:         markers.FieldBool,
				Sample:       "bool: true",
				Tags:         "`json:\"bool,omitempty\"`",
				Comments:     []string{},
				Markers:      []string{},
			},
		},
		{
			name: "new child for struct",
			args: args{
				name:      "struct",
				fieldType: markers.FieldStruct,
				sample:    "struct",
			},
			fields: fields{},
			want: &APIFields{
				Name:         "Struct",
				manifestName: "struct",
				Type:         markers.FieldStruct,
				Sample:       "struct:",
				Tags:         "`json:\"struct,omitempty\"`",
				Comments:     []string{},
				Markers:      []string{},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := &APIFields{
				Name:         tt.fields.Name,
				StructName:   tt.fields.StructName,
				manifestName: tt.fields.manifestName,
				Type:         tt.fields.Type,
				Tags:         tt.fields.Tags,
				Comments:     tt.fields.Comments,
				Markers:      tt.fields.Markers,
				Children:     tt.fields.Children,
				Default:      tt.fields.Default,
				Sample:       tt.fields.Sample,
			}
			got := api.newChild(tt.args.name, tt.args.fieldType, tt.args.sample)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAPIFields_isEqual(t *testing.T) {
	t.Parallel()

	type fields struct {
		Name         string
		StructName   string
		manifestName string
		Type         markers.FieldType
		Tags         string
		Comments     []string
		Markers      []string
		Children     []*APIFields
		Default      string
		Sample       string
	}

	type args struct {
		input *APIFields
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "different field types are not equal",
			args: args{
				input: &APIFields{
					Type: markers.FieldString,
				},
			},
			fields: fields{
				Type: markers.FieldStruct,
			},
			want: false,
		},
		{
			name: "different default values are not equal",
			args: args{
				input: &APIFields{
					Default: "test1",
				},
			},
			fields: fields{
				Default: "test2",
			},
			want: false,
		},
		{
			name: "input comment length of one is equal",
			args: args{
				input: &APIFields{
					Comments: []string{"test"},
				},
			},
			fields: fields{},
			want:   true,
		},
		{
			name: "api comment length of one is equal",
			args: args{
				input: &APIFields{},
			},
			fields: fields{
				Comments: []string{"test"},
			},
			want: true,
		},
		{
			name: "misordered comments are not equal",
			args: args{
				input: &APIFields{
					Comments: []string{"test1", "test2"},
				},
			},
			fields: fields{
				Comments: []string{"test2", "test1"},
			},
			want: false,
		},
		{
			name: "different lengths are not equal",
			args: args{
				input: &APIFields{
					Comments: []string{"test1", "test2"},
				},
			},
			fields: fields{
				Comments: []string{"test1", "test2", "test3"},
			},
			want: false,
		},
		{
			name: "ordered comments are equal",
			args: args{
				input: &APIFields{
					Comments: []string{"test1", "test2"},
				},
			},
			fields: fields{
				Comments: []string{"test1", "test2"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := &APIFields{
				Name:         tt.fields.Name,
				StructName:   tt.fields.StructName,
				manifestName: tt.fields.manifestName,
				Type:         tt.fields.Type,
				Tags:         tt.fields.Tags,
				Comments:     tt.fields.Comments,
				Markers:      tt.fields.Markers,
				Children:     tt.fields.Children,
				Default:      tt.fields.Default,
				Sample:       tt.fields.Sample,
			}
			got := api.isEqual(tt.args.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAPIFields_AddField(t *testing.T) {
	t.Parallel()

	type fields struct {
		Name         string
		StructName   string
		manifestName string
		Type         markers.FieldType
		Tags         string
		Comments     []string
		Markers      []string
		Children     []*APIFields
		Default      string
		Sample       string
	}

	type args struct {
		path       string
		fieldType  markers.FieldType
		comments   []string
		sample     interface{}
		hasDefault bool
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid nested",
			args: args{
				path:       "nested.path",
				fieldType:  markers.FieldString,
				comments:   []string{"test"},
				sample:     "test",
				hasDefault: true,
			},
			fields: fields{
				Comments: []string{"test1", "test2"},
				Children: []*APIFields{
					{
						Type:         markers.FieldStruct,
						manifestName: "nested",
						Children: []*APIFields{
							{
								Type:         markers.FieldString,
								manifestName: "path",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid flat",
			args: args{
				path:       "path",
				fieldType:  markers.FieldString,
				comments:   []string{"test"},
				sample:     "test",
				hasDefault: true,
			},
			fields: fields{
				Comments: []string{"test1", "test2"},
				Children: []*APIFields{
					{
						Type:         markers.FieldString,
						manifestName: "path",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid missing",
			args: args{
				path:       "path",
				fieldType:  markers.FieldString,
				comments:   []string{"test"},
				sample:     "test",
				hasDefault: true,
			},
			fields: fields{
				Comments: []string{"test1", "test2"},
			},
			wantErr: false,
		},
		{
			name: "valid missing nested",
			args: args{
				path:       "nested.path",
				fieldType:  markers.FieldString,
				comments:   []string{"test"},
				sample:     "test",
				hasDefault: true,
			},
			fields: fields{
				Comments: []string{"test1", "test2"},
			},
			wantErr: false,
		},
		{
			name: "ovveride flat value results in an error",
			args: args{
				path:       "nested.path",
				fieldType:  markers.FieldString,
				comments:   []string{"test"},
				sample:     "test",
				hasDefault: true,
			},
			fields: fields{
				Comments: []string{"test1", "test2"},
				Children: []*APIFields{
					{
						manifestName: "nested",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid nested inequal child",
			args: args{
				path:       "nested.path",
				fieldType:  markers.FieldString,
				comments:   []string{"test"},
				sample:     "test",
				hasDefault: true,
			},
			fields: fields{
				Comments: []string{"test1", "test2"},
				Children: []*APIFields{
					{
						Type:         markers.FieldStruct,
						manifestName: "nested",
						Children: []*APIFields{
							{
								Type:         markers.FieldString,
								manifestName: "path",
								Default:      "value",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := &APIFields{
				Name:         tt.fields.Name,
				StructName:   tt.fields.StructName,
				manifestName: tt.fields.manifestName,
				Type:         tt.fields.Type,
				Tags:         tt.fields.Tags,
				Comments:     tt.fields.Comments,
				Markers:      tt.fields.Markers,
				Children:     tt.fields.Children,
				Default:      tt.fields.Default,
				Sample:       tt.fields.Sample,
			}

			if err := api.AddField(
				tt.args.path, tt.args.fieldType, tt.args.comments, tt.args.sample, tt.args.hasDefault,
			); (err != nil) != tt.wantErr {
				t.Errorf("APIFields.AddField() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
