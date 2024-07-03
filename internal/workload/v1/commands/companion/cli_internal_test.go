// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package companion

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type companionCLITester struct {
	collection bool
	kind       string
}

func (tester *companionCLITester) IsCollection() bool {
	return tester.collection
}

func (tester *companionCLITester) GetAPIKind() string {
	return tester.kind
}

func TestCLI_SetDefaults(t *testing.T) {
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
		workload     companionCLIProcessor
		isSubcommand bool
	}

	tests := []struct {
		name         string
		fields       fields
		args         args
		expected     *CLI
		isSubcommand bool
	}{
		{
			name:   "ensure collection fields are properly defaulted",
			fields: fields{},
			args: args{
				workload: &companionCLITester{
					collection: true,
					kind:       "NeedsDefaulted",
				},
				isSubcommand: true,
			},
			expected: &CLI{
				Name:          defaultCollectionSubcommandName,
				Description:   fmt.Sprintf("Manage %s workload", "needsdefaulted"),
				IsSubcommand:  true,
				IsRootcommand: false,
			},
		},
		{
			name:   "ensure component fields are properly defaulted",
			fields: fields{},
			args: args{
				workload: &companionCLITester{
					collection: false,
					kind:       "NeedsDefaulted",
				},
				isSubcommand: true,
			},
			expected: &CLI{
				Name:          "needsdefaulted",
				Description:   fmt.Sprintf("Manage %s workload", "needsdefaulted"),
				IsSubcommand:  true,
				IsRootcommand: false,
			},
		},
		{
			name:   "ensure standalone fields are properly defaulted",
			fields: fields{},
			args: args{
				workload: &companionCLITester{
					collection: false,
					kind:       "NeedsDefaulted",
				},
				isSubcommand: true,
			},
			expected: &CLI{
				Name:          "needsdefaulted",
				Description:   fmt.Sprintf("Manage %s workload", "needsdefaulted"),
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
				workload: &companionCLITester{
					collection: true,
					kind:       "RemainPersistent",
				},
				isSubcommand: true,
			},
			expected: &CLI{
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
				workload: &companionCLITester{
					collection: true,
					kind:       "SubCommand",
				},
				isSubcommand: true,
			},
			expected: &CLI{
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
				workload: &companionCLITester{
					collection: true,
					kind:       "RootCommand",
				},
				isSubcommand: false,
			},
			expected: &CLI{
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
			cli := &CLI{
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

func TestCLI_getDefaultName(t *testing.T) {
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
		workload companionCLIProcessor
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
				workload: &companionCLITester{
					collection: true,
					kind:       "CollectionRoot",
				},
			},
			want: "collectionroot",
		},
		{
			name: "ensure collection sub command name is properly returned",
			fields: fields{
				Name:         "collection-sub",
				Description:  "collection-sub",
				IsSubcommand: true,
			},
			args: args{
				workload: &companionCLITester{
					collection: true,
					kind:       "CollectionSub",
				},
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
				workload: &companionCLITester{
					collection: false,
					kind:       "ComponentSub",
				},
			},
			want: "componentsub",
		},
		{
			name: "ensure standalone root command name is properly returned",
			fields: fields{
				Name:          "standalone-root",
				Description:   "standalone-root",
				IsRootcommand: true,
			},
			args: args{
				workload: &companionCLITester{
					collection: false,
					kind:       "StandaloneRoot",
				},
			},
			want: "standaloneroot",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cli := &CLI{
				Name:          tt.fields.Name,
				Description:   tt.fields.Description,
				VarName:       tt.fields.VarName,
				FileName:      tt.fields.FileName,
				IsSubcommand:  tt.fields.IsSubcommand,
				IsRootcommand: tt.fields.IsRootcommand,
			}
			if got := cli.getDefaultName(tt.args.workload); got != tt.want {
				t.Errorf("CLI.getDefaultName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCLI_getDefaultDescription(t *testing.T) {
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
		workload companionCLIProcessor
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
				workload: &companionCLITester{
					collection: true,
					kind:       "CollectionRoot",
				},
			},
			want: fmt.Sprintf(defaultCollectionRootcommandDescription, "collectionroot"),
		},
		{
			name: "ensure collection sub command description is properly returned",
			fields: fields{
				Name:         "collection-sub",
				Description:  "collection-sub",
				IsSubcommand: true,
			},
			args: args{
				workload: &companionCLITester{
					collection: true,
					kind:       "CollectionSub",
				},
			},
			want: fmt.Sprintf(defaultCollectionSubcommandDescription, "collectionsub"),
		},
		{
			name: "ensure component sub command description is properly returned",
			fields: fields{
				Name:         "component-sub",
				Description:  "component-sub",
				IsSubcommand: true,
			},
			args: args{
				workload: &companionCLITester{
					collection: true,
					kind:       "ComponentSub",
				},
			},
			want: fmt.Sprintf(defaultDescription, "componentsub"),
		},
		{
			name: "ensure standalone root command description is properly returned",
			fields: fields{
				Name:          "standalone-root",
				Description:   "standalone-root",
				IsRootcommand: true,
			},
			args: args{
				workload: &companionCLITester{
					collection: false,
					kind:       "StandaloneRoot",
				},
			},
			want: fmt.Sprintf(defaultDescription, "standaloneroot"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cli := &CLI{
				Name:          tt.fields.Name,
				Description:   tt.fields.Description,
				VarName:       tt.fields.VarName,
				FileName:      tt.fields.FileName,
				IsSubcommand:  tt.fields.IsSubcommand,
				IsRootcommand: tt.fields.IsRootcommand,
			}
			if got := cli.getDefaultDescription(tt.args.workload); got != tt.want {
				t.Errorf("CLI.getDefaultDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCLI_SetCommonValues(t *testing.T) {
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
		workload     companionCLIProcessor
		isSubcommand bool
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		expected *CLI
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
				workload: &companionCLITester{
					collection: true,
					kind:       "VarNameFileName",
				},
				isSubcommand: true,
			},
			expected: &CLI{
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
			cli := &CLI{
				Name:          tt.fields.Name,
				Description:   tt.fields.Description,
				VarName:       tt.fields.VarName,
				FileName:      tt.fields.FileName,
				IsSubcommand:  tt.fields.IsSubcommand,
				IsRootcommand: tt.fields.IsRootcommand,
			}
			cli.SetCommonValues(tt.args.workload, tt.args.isSubcommand)
			assert.Equal(t, tt.expected, cli)
		})
	}
}

func Test_HasName(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		input    CLI
		expected bool
	}{
		{
			name: "command has a name field",
			input: CLI{
				Name: "HasNameField",
			},
			expected: true,
		},
		{
			name:     "command does not have a name field",
			input:    CLI{},
			expected: false,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hasName := tt.input.HasName()
			assert.Equal(t, tt.expected, hasName)
		})
	}
}

func Test_HasDescription(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		input    CLI
		expected bool
	}{
		{
			name: "command has a description field",
			input: CLI{
				Description: "HasDescriptionField",
			},
			expected: true,
		},
		{
			name:     "command does not have a description field",
			input:    CLI{},
			expected: false,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hasDescription := tt.input.HasDescription()
			assert.Equal(t, tt.expected, hasDescription)
		})
	}
}

func TestCLI_GetSubCmdRelativeFileName(t *testing.T) {
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
		rootCmdName      string
		subCommandFolder string
		group            string
		fileName         string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "ensure file path generation is correct",
			fields: fields{
				Name:          "mycommand",
				VarName:       "Mycommand",
				FileName:      "mycommand",
				Description:   "mycommand test",
				IsSubcommand:  true,
				IsRootcommand: false,
			},
			args: args{
				rootCmdName:      "testctl",
				subCommandFolder: "test",
				group:            "test",
				fileName:         "command",
			},
			want: "cmd/testctl/commands/test/test/command.go",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cli := &CLI{
				Name:          tt.fields.Name,
				Description:   tt.fields.Description,
				VarName:       tt.fields.VarName,
				FileName:      tt.fields.FileName,
				IsSubcommand:  tt.fields.IsSubcommand,
				IsRootcommand: tt.fields.IsRootcommand,
			}
			if got := cli.GetSubCmdRelativeFileName(
				tt.args.rootCmdName,
				tt.args.subCommandFolder,
				tt.args.group,
				tt.args.fileName,
			); got != tt.want {
				t.Errorf("CLI.GetSubCmdRelativeFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}
