// Copyright 2024 Nukleros
// SPDX-License-Identifier: MIT

package manifests

import (
	"reflect"
	"testing"
)

func Test_getFileNames(t *testing.T) {
	t.Parallel()

	type args struct {
		relativeFileName string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "ensure file names are shown in the appropriate order",
			args: args{
				relativeFileName: "workload/path/to/my/workload-resources.yaml",
			},
			want: []string{
				"workload_resources.go",
				"my_workload_resources.go",
				"to_my_workload_resources.go",
				"path_to_my_workload_resources.go",
				"workload_path_to_my_workload_resources.go",
			},
		},
		{
			name: "ensure file names that start with a hidden directory are shown in the appropriate order",
			args: args{
				relativeFileName: ".workload/path/to/my/workload-resources.yaml",
			},
			want: []string{
				"workload_resources.go",
				"my_workload_resources.go",
				"to_my_workload_resources.go",
				"path_to_my_workload_resources.go",
				"workload_path_to_my_workload_resources.go",
			},
		},
		{
			name: "ensure file names that end in test are shown in the appropriate order",
			args: args{
				relativeFileName: ".workload/path/to/my/workload-resources-test.yaml",
			},
			want: []string{
				"workload_resources.go",
				"my_workload_resources.go",
				"to_my_workload_resources.go",
				"path_to_my_workload_resources.go",
				"workload_path_to_my_workload_resources.go",
			},
		},
		{
			name: "ensure file names that end in internal-test are shown in the appropriate order",
			args: args{
				relativeFileName: ".workload/path/to/my/workload-resources-internal-test.yaml",
			},
			want: []string{
				"workload_resources.go",
				"my_workload_resources.go",
				"to_my_workload_resources.go",
				"path_to_my_workload_resources.go",
				"workload_path_to_my_workload_resources.go",
			},
		},
		{
			name: "ensure file names that are up a directory structure are shown in the appropriate order",
			args: args{
				relativeFileName: "../../workload/path/to/my/workload-resources-internal-test.yaml",
			},
			want: []string{
				"workload_resources.go",
				"my_workload_resources.go",
				"to_my_workload_resources.go",
				"path_to_my_workload_resources.go",
				"workload_path_to_my_workload_resources.go",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getFileNames(tt.args.relativeFileName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFileNames() = %v, want %v", got, tt.want)
			}
		})
	}
}
