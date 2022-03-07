// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package rbac

import (
	"reflect"
	"testing"
)

func Test_getGroup(t *testing.T) {
	t.Parallel()

	type args struct {
		group string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ensure empty group returns core group",
			args: args{
				group: "",
			},
			want: coreGroup,
		},
		{
			name: "ensure other group returns itself",
			args: args{
				group: "thisisatestgroup",
			},
			want: "thisisatestgroup",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getGroup(tt.args.group); got != tt.want {
				t.Errorf("getGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getFieldString(t *testing.T) {
	t.Parallel()

	type args struct {
		fields []string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ensure strings return semicolon separated string",
			args: args{
				fields: []string{"one", "two", "three"},
			},
			want: "one;two;three",
		},
		{
			name: "ensure empty string array returns empty string",
			args: args{
				fields: []string{},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getFieldString(tt.args.fields); got != tt.want {
				t.Errorf("getFieldString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getResource(t *testing.T) {
	t.Parallel()

	type args struct {
		kind string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ensure status kind returns the plural kind with the status appended",
			args: args{
				kind: "apple/status",
			},
			want: "apples/status",
		},
		{
			name: "ensure wildcard kind returns wildcard",
			args: args{
				kind: "*",
			},
			want: "*",
		},
		{
			name: "ensure wildcard kind with slash returns wildcard with slash",
			args: args{
				kind: "*/status",
			},
			want: "*/status",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getResource(tt.args.kind); got != tt.want {
				t.Errorf("getResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPlural(t *testing.T) {
	t.Parallel()

	type args struct {
		kind string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ensure kind returns correct plural value",
			args: args{
				kind: "apples",
			},
			want: "apples",
		},
		{
			name: "ensure kind with an irregular plural returns correct plural value",
			args: args{
				kind: "resourcequota",
			},
			want: "resourcequotas",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getPlural(tt.args.kind); got != tt.want {
				t.Errorf("getPlural() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_valueFromInterface(t *testing.T) {
	t.Parallel()

	mapIntInt := map[interface{}]interface{}{
		"key": "value",
	}

	mapStringInt := map[string]interface{}{
		"key": "value",
	}

	mapIntArrayInt := map[interface{}][]interface{}{
		"key": {"value"},
	}

	mapStringArrayInt := map[string][]interface{}{
		"key": {"value"},
	}

	type args struct {
		in  interface{}
		key string
	}

	tests := []struct {
		name    string
		args    args
		wantOut interface{}
	}{
		{
			name: "ensure map interface interface returns value",
			args: args{
				in:  mapIntInt,
				key: "key",
			},
			wantOut: "value",
		},
		{
			name: "ensure map string interface returns value",
			args: args{
				in:  mapStringInt,
				key: "key",
			},
			wantOut: "value",
		},
		{
			name: "ensure map interface array interface returns value",
			args: args{
				in:  mapIntArrayInt,
				key: "key",
			},
			wantOut: []interface{}{"value"},
		},
		{
			name: "ensure map string array interface returns value",
			args: args{
				in:  mapStringArrayInt,
				key: "key",
			},
			wantOut: []interface{}{"value"},
		},
		{
			name: "ensure unknown returns nil",
			args: args{
				in:  []string{"test"},
				key: "key",
			},
			wantOut: nil,
		},
		{
			name: "ensure nil returns nil",
			args: args{
				in:  nil,
				key: "key",
			},
			wantOut: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if gotOut := valueFromInterface(tt.args.in, tt.args.key); !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("valueFromInterface() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
