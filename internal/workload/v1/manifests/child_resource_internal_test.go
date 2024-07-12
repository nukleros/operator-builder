// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package manifests

import "testing"

func TestChildResource_MutateFileName(t *testing.T) {
	t.Parallel()

	type fields struct {
		UniqueName string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure file name returns as expected",
			fields: fields{
				UniqueName: "ThisIsAFunctionName",
			},
			want: "this_is_a_function_name.go",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resource := &ChildResource{
				UniqueName: tt.fields.UniqueName,
			}
			if got := resource.MutateFileName(); got != tt.want {
				t.Errorf("ChildResource.MutateFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stripTags(t *testing.T) {
	t.Parallel()

	type args struct {
		value string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ensure tags are stripped",
			args: args{
				value: "!!start parent.Spec.Some.Thing !!end",
			},
			want: " parent.Spec.Some.Thing ",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := stripTags(tt.args.value); got != tt.want {
				t.Errorf("stripTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChildResource_NameComment(t *testing.T) {
	t.Parallel()

	type fields struct {
		Name string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ensure resource without tags return appropriately",
			fields: fields{
				Name: "my-service",
			},
			want: "my-service",
		},
		{
			name: "ensure resource with tags as prefix and suffix return appropriately",
			fields: fields{
				Name: "!!start parent.Spec.Some.Thing !!end",
			},
			want: "parent.spec.some.thing",
		},
		{
			name: "ensure resource with replace after suffix return appropriately",
			fields: fields{
				Name: "!!start parent.Spec.Some.Thing !!end-svc",
			},
			want: "parent.spec.some.thing + -svc",
		},
		{
			name: "ensure resource with replace before prefix return appropriately",
			fields: fields{
				Name: "prefix-!!start parent.Spec.Some.Thing !!end",
			},
			want: "prefix- + parent.spec.some.thing",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resource := &ChildResource{
				Name: tt.fields.Name,
			}
			if got := resource.NameComment(); got != tt.want {
				t.Errorf("ChildResource.NameComment() = %v, want %v", got, tt.want)
			}
		})
	}
}
