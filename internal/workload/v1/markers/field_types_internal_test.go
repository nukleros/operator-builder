// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package markers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldType_UnmarshalMarkerArg(t *testing.T) {
	t.Parallel()

	type args struct {
		in string
	}

	tests := []struct {
		name    string
		f       FieldType
		args    args
		wantErr bool
		expect  FieldType
	}{
		{
			name: "invalid type should return error",
			f:    FieldUnknownType,
			args: args{
				in: "fake",
			},
			wantErr: true,
			expect:  FieldUnknownType,
		},
		{
			name: "unknown field type should return error",
			f:    FieldUnknownType,
			args: args{
				in: "",
			},
			wantErr: true,
			expect:  FieldUnknownType,
		},
		{
			name: "string field type appropriately unmarshaled",
			f:    FieldString,
			args: args{
				in: "string",
			},
			wantErr: false,
			expect:  FieldString,
		},
		{
			name: "int field type appropriately unmarshaled",
			f:    FieldInt,
			args: args{
				in: "int",
			},
			wantErr: false,
			expect:  FieldInt,
		},
		{
			name: "bool field type appropriately unmarshaled",
			f:    FieldBool,
			args: args{
				in: "bool",
			},
			wantErr: false,
			expect:  FieldBool,
		},
		{
			name: "mismatched field type appropriately unmarshaled",
			f:    FieldUnknownType,
			args: args{
				in: "string",
			},
			wantErr: false,
			expect:  FieldString,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.f.UnmarshalMarkerArg(tt.args.in); (err != nil) != tt.wantErr {
				t.Errorf("FieldType.UnmarshalMarkerArg() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.expect, tt.f)
		})
	}
}

func TestFieldType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    FieldType
		want string
	}{
		{
			name: "unknown field type returns ''",
			f:    FieldUnknownType,
			want: "",
		},
		{
			name: "string field type returns 'string'",
			f:    FieldString,
			want: "string",
		},
		{
			name: "struct field type returns 'struct'",
			f:    FieldStruct,
			want: "struct",
		},
		{
			name: "int field type returns 'int'",
			f:    FieldInt,
			want: "int",
		},
		{
			name: "bool field type returns 'bool'",
			f:    FieldBool,
			want: "bool",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.f.String(); got != tt.want {
				t.Errorf("FieldType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
