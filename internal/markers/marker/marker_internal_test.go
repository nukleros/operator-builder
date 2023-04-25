// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package marker

/*func Test_argumentInfo(t *testing.T) {
	t.Parallel()

	type args struct {
		fieldName string
		tag       reflect.StructTag
	}

	tests := []struct {
		name            string
		args            args
		wantArgName     string
		wantOptionalOpt bool
	}{
		{
			name: "correctly finds marker tag with field name override",
			args: args{
				fieldName: "tomato",
				tag:       reflect.StructTag(`marker:"vegetable"`),
			},
			wantArgName:     "vegetable",
			wantOptionalOpt: false,
		},
		{
			name: "correctly finds marker tag with field name override and optional flag",
			args: args{
				fieldName: "tomato",
				tag:       reflect.StructTag(`marker:"vegetable,optional"`),
			},
			wantArgName:     "vegetable",
			wantOptionalOpt: true,
		},
		{
			name: "correctly finds marker tag with optional flag",
			args: args{
				fieldName: "tomato",
				tag:       reflect.StructTag(`marker:",optional"`),
			},
			wantArgName:     "tomato",
			wantOptionalOpt: true,
		},
		{
			name: "does not pick up other tags",
			args: args{
				fieldName: "tomato",
				tag:       reflect.StructTag(`potato:",optional"`),
			},
			wantArgName:     "tomato",
			wantOptionalOpt: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotArgName, gotOptionalOpt := argumentInfo(tt.args.fieldName, tt.args.tag)
			if gotArgName != tt.wantArgName {
				t.Errorf("argumentInfo(%v, %v) gotArgName = %v, want %v", tt.args.fieldName, tt.args.tag, gotArgName, tt.wantArgName)
			}
			if gotOptionalOpt != tt.wantOptionalOpt {
				t.Errorf("argumentInfo(%v, %v) gotOptionalOpt = %v, want %v", tt.args.fieldName, tt.args.tag, gotOptionalOpt, tt.wantOptionalOpt)
			}
		})
	}
}

func TestMarker_loadFields(t *testing.T) {
	t.Parallel()

	type fields struct {
		Name   string
		Output reflect.Type
		Fields map[string]Argument
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &Marker{
				Name:   tt.fields.Name,
				Output: tt.fields.Output,
				Fields: tt.fields.Fields,
			}
			if err := m.loadFields(); (err != nil) != tt.wantErr {
				t.Errorf("Marker.loadFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
*/
