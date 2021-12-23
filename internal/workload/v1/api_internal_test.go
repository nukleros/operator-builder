// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"testing"
)

func TestAPIFields_GenerateSampleSpec(t *testing.T) {
	t.Parallel()

	type fields struct {
		Name         string
		manifestName string
		Type         FieldType
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
