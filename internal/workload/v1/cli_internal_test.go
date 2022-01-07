// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCliCommand_SetDefaults(t *testing.T) {
	t.Parallel()

	apiTestSpec := NewSampleAPISpec()

	type fields struct {
		Name          string
		Description   string
		VarName       string
		FileName      string
		IsSubcommand  bool
		IsRootcommand bool
	}

	type args struct {
		workload     WorkloadAPIBuilder
		isSubcommand bool
	}

	tests := []struct {
		name         string
		fields       fields
		args         args
		expected     *CliCommand
		isSubcommand bool
	}{
		{
			name:   "ensure collection fields are properly defaulted",
			fields: fields{},
			args: args{
				workload: NewWorkloadCollection(
					"needs-defaulted",
					*apiTestSpec,
					[]string{},
				),
				isSubcommand: true,
			},
			expected: &CliCommand{
				Name:          defaultCollectionSubcommandName,
				Description:   fmt.Sprintf("Manage %s workload", strings.ToLower(apiTestSpec.Kind)),
				IsSubcommand:  true,
				IsRootcommand: false,
			},
		},
		{
			name:   "ensure component fields are properly defaulted",
			fields: fields{},
			args: args{
				workload: NewComponentWorkload(
					"needs-defaulted",
					*apiTestSpec,
					[]string{},
					[]string{},
				),
				isSubcommand: true,
			},
			expected: &CliCommand{
				Name:          strings.ToLower(apiTestSpec.Kind),
				Description:   fmt.Sprintf("Manage %s workload", strings.ToLower(apiTestSpec.Kind)),
				IsSubcommand:  true,
				IsRootcommand: false,
			},
		},
		{
			name:   "ensure standalone fields are properly defaulted",
			fields: fields{},
			args: args{
				workload: NewStandaloneWorkload(
					"needs-defaulted",
					*apiTestSpec,
					[]string{},
				),
				isSubcommand: true,
			},
			expected: &CliCommand{
				Name:          strings.ToLower(apiTestSpec.Kind),
				Description:   fmt.Sprintf("Manage %s workload", strings.ToLower(apiTestSpec.Kind)),
				IsSubcommand:  true,
				IsRootcommand: false,
			},
		},
		{
			name: "ensure default fields remain persistent",
			fields: fields{
				Name:        "remain-persistent",
				Description: "remain-persistent",
			},
			args: args{
				workload: NewWorkloadCollection(
					"remain-persistent",
					*apiTestSpec,
					[]string{},
				),
				isSubcommand: true,
			},
			expected: &CliCommand{
				Name:          "remain-persistent",
				Description:   "remain-persistent",
				IsSubcommand:  true,
				IsRootcommand: false,
			},
		},
		{
			name: "ensure subcommand and rootcommand fields are properly set if subcommand",
			fields: fields{
				Name:        "sub-root-for-sub",
				Description: "sub-root-for-sub",
			},
			args: args{
				workload: NewWorkloadCollection(
					"remain-persistent",
					*apiTestSpec,
					[]string{},
				),
				isSubcommand: true,
			},
			expected: &CliCommand{
				Name:          "sub-root-for-sub",
				Description:   "sub-root-for-sub",
				IsSubcommand:  true,
				IsRootcommand: false,
			},
		},
		{
			name: "ensure subcommand and rootcommand fields are properly set if rootcommand",
			fields: fields{
				Name:        "sub-root-for-root",
				Description: "sub-root-for-root",
			},
			args: args{
				workload: NewWorkloadCollection(
					"remain-persistent",
					*apiTestSpec,
					[]string{},
				),
				isSubcommand: false,
			},
			expected: &CliCommand{
				Name:          "sub-root-for-root",
				Description:   "sub-root-for-root",
				IsSubcommand:  false,
				IsRootcommand: true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cli := &CliCommand{
				Name:          tt.fields.Name,
				Description:   tt.fields.Description,
				VarName:       tt.fields.VarName,
				FileName:      tt.fields.FileName,
				IsSubcommand:  tt.fields.IsSubcommand,
				IsRootcommand: tt.fields.IsRootcommand,
			}
			cli.SetDefaults(tt.args.workload, tt.args.isSubcommand)
			assert.Equal(t, tt.expected, cli)
		})
	}
}

func TestCliCommand_getDefaultName(t *testing.T) {
	t.Parallel()

	apiTestSpec := NewSampleAPISpec()

	type fields struct {
		Name          string
		Description   string
		VarName       string
		FileName      string
		IsSubcommand  bool
		IsRootcommand bool
	}

	type args struct {
		workload WorkloadAPIBuilder
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "ensure collection root command name is properly returned",
			fields: fields{
				Name:          "collection-root",
				Description:   "collection-root",
				IsRootcommand: true,
			},
			args: args{
				workload: NewWorkloadCollection(
					"collection-root",
					*apiTestSpec,
					[]string{},
				),
			},
			want: strings.ToLower(apiTestSpec.Kind),
		},
		{
			name: "ensure collection sub command name is properly returned",
			fields: fields{
				Name:         "collection-sub",
				Description:  "collection-sub",
				IsSubcommand: true,
			},
			args: args{
				workload: NewWorkloadCollection(
					"collection-sub",
					*apiTestSpec,
					[]string{},
				),
			},
			want: defaultCollectionSubcommandName,
		},
		{
			name: "ensure component sub command name is properly returned",
			fields: fields{
				Name:         "component-sub",
				Description:  "component-sub",
				IsSubcommand: true,
			},
			args: args{
				workload: NewComponentWorkload(
					"component-sub",
					*apiTestSpec,
					[]string{},
					[]string{},
				),
			},
			want: strings.ToLower(apiTestSpec.Kind),
		},
		{
			name: "ensure standalone root command name is properly returned",
			fields: fields{
				Name:          "standalone-root",
				Description:   "standalone-root",
				IsRootcommand: true,
			},
			args: args{
				workload: NewStandaloneWorkload(
					"standalone-root",
					*apiTestSpec,
					[]string{},
				),
			},
			want: strings.ToLower(apiTestSpec.Kind),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cli := &CliCommand{
				Name:          tt.fields.Name,
				Description:   tt.fields.Description,
				VarName:       tt.fields.VarName,
				FileName:      tt.fields.FileName,
				IsSubcommand:  tt.fields.IsSubcommand,
				IsRootcommand: tt.fields.IsRootcommand,
			}
			if got := cli.getDefaultName(tt.args.workload); got != tt.want {
				t.Errorf("CliCommand.getDefaultName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCliCommand_getDefaultDescription(t *testing.T) {
	t.Parallel()

	apiTestSpec := NewSampleAPISpec()
	lowerKind := strings.ToLower(apiTestSpec.Kind)

	type fields struct {
		Name          string
		Description   string
		VarName       string
		FileName      string
		IsSubcommand  bool
		IsRootcommand bool
	}

	type args struct {
		workload WorkloadAPIBuilder
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "ensure collection root command description is properly returned",
			fields: fields{
				Name:          "collection-root",
				Description:   "collection-root",
				IsRootcommand: true,
			},
			args: args{
				workload: NewWorkloadCollection(
					"collection-root",
					*apiTestSpec,
					[]string{},
				),
			},
			want: fmt.Sprintf(defaultCollectionRootcommandDescription, lowerKind),
		},
		{
			name: "ensure collection sub command description is properly returned",
			fields: fields{
				Name:         "collection-sub",
				Description:  "collection-sub",
				IsSubcommand: true,
			},
			args: args{
				workload: NewWorkloadCollection(
					"collection-sub",
					*apiTestSpec,
					[]string{},
				),
			},
			want: fmt.Sprintf(defaultCollectionSubcommandDescription, lowerKind),
		},
		{
			name: "ensure component sub command description is properly returned",
			fields: fields{
				Name:         "component-sub",
				Description:  "component-sub",
				IsSubcommand: true,
			},
			args: args{
				workload: NewComponentWorkload(
					"component-sub",
					*apiTestSpec,
					[]string{},
					[]string{},
				),
			},
			want: fmt.Sprintf(defaultDescription, lowerKind),
		},
		{
			name: "ensure standalone root command description is properly returned",
			fields: fields{
				Name:          "standalone-root",
				Description:   "standalone-root",
				IsRootcommand: true,
			},
			args: args{
				workload: NewStandaloneWorkload(
					"standalone-root",
					*apiTestSpec,
					[]string{},
				),
			},
			want: fmt.Sprintf(defaultDescription, lowerKind),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cli := &CliCommand{
				Name:          tt.fields.Name,
				Description:   tt.fields.Description,
				VarName:       tt.fields.VarName,
				FileName:      tt.fields.FileName,
				IsSubcommand:  tt.fields.IsSubcommand,
				IsRootcommand: tt.fields.IsRootcommand,
			}
			if got := cli.getDefaultDescription(tt.args.workload); got != tt.want {
				t.Errorf("CliCommand.getDefaultDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCliCommand_setCommonValues(t *testing.T) {
	t.Parallel()

	type fields struct {
		Name          string
		Description   string
		VarName       string
		FileName      string
		IsSubcommand  bool
		IsRootcommand bool
	}

	type args struct {
		workload     WorkloadAPIBuilder
		isSubcommand bool
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		expected *CliCommand
	}{
		{
			name: "ensure varname and filename fields are set",
			fields: fields{
				Name:          "mycommand",
				Description:   "mycommand test",
				IsSubcommand:  true,
				IsRootcommand: false,
			},
			args: args{
				workload: NewWorkloadCollection(
					"missing-description",
					*NewSampleAPISpec(),
					[]string{},
				),
				isSubcommand: true,
			},
			expected: &CliCommand{
				Name:          "mycommand",
				VarName:       "Mycommand",
				FileName:      "mycommand",
				Description:   "mycommand test",
				IsSubcommand:  true,
				IsRootcommand: false,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cli := &CliCommand{
				Name:          tt.fields.Name,
				Description:   tt.fields.Description,
				VarName:       tt.fields.VarName,
				FileName:      tt.fields.FileName,
				IsSubcommand:  tt.fields.IsSubcommand,
				IsRootcommand: tt.fields.IsRootcommand,
			}
			cli.setCommonValues(tt.args.workload, tt.args.isSubcommand)
			assert.Equal(t, tt.expected, cli)
		})
	}
}

func Test_hasName(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		input    CliCommand
		expected bool
	}{
		{
			name: "command has a name field",
			input: CliCommand{
				Name: "HasNameField",
			},
			expected: true,
		},
		{
			name:     "command does not have a name field",
			input:    CliCommand{},
			expected: false,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hasName := tt.input.hasName()
			assert.Equal(t, tt.expected, hasName)
		})
	}
}

func Test_hasDescription(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		input    CliCommand
		expected bool
	}{
		{
			name: "command has a description field",
			input: CliCommand{
				Description: "HasDescriptionField",
			},
			expected: true,
		},
		{
			name:     "command does not have a description field",
			input:    CliCommand{},
			expected: false,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hasDescription := tt.input.hasDescription()
			assert.Equal(t, tt.expected, hasDescription)
		})
	}
}
