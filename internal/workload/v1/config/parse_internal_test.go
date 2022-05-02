// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	if err != nil {
		panic("unable to get working directory for `parse_internal_test.go` test")
	}

	testPath := wd + "/../../../../test/"

	type args struct {
		configPath string
	}

	tests := []struct {
		name    string
		args    args
		want    *Processor // for now we are only testing for an error
		wantErr bool
	}{
		{
			name: "ensure valid simple standalone does not return an error",
			args: args{
				configPath: testPath + "cases/standalone/.workloadConfig/workload.yaml",
			},
			wantErr: false,
		},
		{
			name: "ensure valid simple collection does not return an error",
			args: args{
				configPath: testPath + "cases/collection/.workloadConfig/workload.yaml",
			},
			wantErr: false,
		},
		{
			name: "ensure valid complex standalone does not return an error",
			args: args{
				configPath: testPath + "cases/edge-standalone/.workloadConfig/workload.yaml",
			},
			wantErr: false,
		},
		{
			name: "ensure valid complex collection does not return an error",
			args: args{
				configPath: testPath + "cases/edge-collection/.workloadConfig/workload.yaml",
			},
			wantErr: false,
		},
		{
			name: "ensure passing a component workload as the parent returns an error",
			args: args{
				configPath: testPath + "configs/component/valid.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure error when config path is blank",
			args: args{
				configPath: "",
			},
			wantErr: true,
		},
		{
			name: "ensure file with invalid yaml cannot parse",
			args: args{
				configPath: testPath + "configs/component/invalid-yaml.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure missing file returns error",
			args: args{
				configPath: testPath + "configs/collection/this-does-not-exist.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure collection which contains a component with missing dependencies returns an error",
			args: args{
				configPath: testPath + "configs/collection/invalid-missing-dependencies.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure collection which contains a component with overlapping names returns an error",
			args: args{
				configPath: testPath + "configs/collection/invalid-overlapping-names.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure collection which contains a component with overlapping kinds returns an error",
			args: args{
				configPath: testPath + "configs/collection/invalid-overlapping-kinds.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure standalone with missing domain returns an error",
			args: args{
				configPath: testPath + "configs/standalone/invalid-missing-domain.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure standalone with missing group returns an error",
			args: args{
				configPath: testPath + "configs/standalone/invalid-missing-group.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure standalone with missing version returns an error",
			args: args{
				configPath: testPath + "configs/standalone/invalid-missing-version.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure stadalone with missing kind returns an error",
			args: args{
				configPath: testPath + "configs/standalone/invalid-missing-kind.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure standalone with missing name returns an error",
			args: args{
				configPath: testPath + "configs/standalone/invalid-missing-name.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure collection with missing domain returns an error",
			args: args{
				configPath: testPath + "configs/collection/invalid-missing-domain.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure collection with missing group returns an error",
			args: args{
				configPath: testPath + "configs/collection/invalid-missing-group.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure collection with missing version returns an error",
			args: args{
				configPath: testPath + "configs/collection/invalid-missing-version.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure collection with missing kind returns an error",
			args: args{
				configPath: testPath + "configs/collection/invalid-missing-kind.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure collection with missing name returns an error",
			args: args{
				configPath: testPath + "configs/collection/invalid-missing-name.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure component with missing group returns an error",
			args: args{
				configPath: testPath + "configs/component/invalid-missing-group.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure component with missing version returns an error",
			args: args{
				configPath: testPath + "configs/component/invalid-missing-version.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure component with missing kind returns an error",
			args: args{
				configPath: testPath + "configs/component/invalid-missing-kind.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure component with missing name returns an error",
			args: args{
				configPath: testPath + "configs/component/invalid-missing-name.yaml",
			},
			wantErr: true,
		},
		{
			name: "ensure workload with invalid type returns an error",
			args: args{
				configPath: testPath + "configs/invalid-type.yaml",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(tt.args.configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
		})
	}
}
