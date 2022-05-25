// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestRules_hasResourceRule(t *testing.T) {
	t.Parallel()

	type args struct {
		rule *Rule
	}

	tests := []struct {
		name  string
		rules *Rules
		args  args
		want  bool
	}{
		{
			name:  "ensure rule set with existing rule returns true",
			rules: NewTestRules(),
			args: args{
				rule: NewTestRule(),
			},
			want: true,
		},
		{
			name:  "ensure rule set without rule returns false",
			rules: NewTestRules(),
			args: args{
				rule: &Rule{
					Group:    "fake",
					Resource: "alsoFake",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.rules.hasResourceRule(tt.args.rule); got != tt.want {
				t.Errorf("Rules.hasResourceRule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRules_addForManifest(t *testing.T) {
	t.Parallel()

	empty := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "ClusterRole",
			"metadata": map[string]interface{}{
				"name": "clusterrole",
			},
		},
	}

	invalidRules := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "ClusterRole",
			"metadata": map[string]interface{}{
				"name": "clusterrole",
			},
			"rules": map[string]interface{}{
				"apiGroups": "whoops",
			},
		},
	}

	invalidRule := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "ClusterRole",
			"metadata": map[string]interface{}{
				"name": "clusterrole",
			},
			"rules": []interface{}{
				map[string]interface{}{
					"apiGroups": "whoops",
					"resources": []interface{}{
						"services",
					},
					"verbs": []interface{}{
						"get",
					},
				},
			},
		},
	}

	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "Service",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"name":      "contour-svc",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"app": "contour",
				},
				"ports": []interface{}{
					map[string]interface{}{
						"protocol":   "TCP",
						"port":       80,
						"targetPort": 8080,
					},
				},
			},
		},
	}

	resourceRoleRule := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "ClusterRole",
			"metadata": map[string]interface{}{
				"name": "clusterrole",
			},
			"rules": []interface{}{
				map[string]interface{}{
					"apiGroups": []interface{}{
						"group",
					},
					"resources": []interface{}{
						"services",
					},
					"verbs": []interface{}{
						"get",
					},
				},
			},
		},
	}

	nonResourceRoleRule := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "ClusterRole",
			"metadata": map[string]interface{}{
				"name": "clusterrole",
			},
			"rules": []interface{}{
				map[string]interface{}{
					"nonResourceURLs": []interface{}{
						"/metrics",
					},
					"verbs": []interface{}{
						"get",
					},
				},
			},
		},
	}

	type args struct {
		manifest *unstructured.Unstructured
	}

	tests := []struct {
		name    string
		rules   *Rules
		args    args
		wantErr bool
		want    *Rules
	}{
		{
			name:  "resource without rules should return no error and add the regular resource",
			rules: &Rules{},
			args: args{
				manifest: empty,
			},
			wantErr: false,
			want: &Rules{
				{
					Group:    "rbac.authorization.k8s.io",
					Resource: "clusterroles",
					Verbs:    defaultResourceVerbs(),
				},
			},
		},
		{
			name:  "resource with invalid rule should return error and add the regular resource",
			rules: &Rules{},
			args: args{
				manifest: invalidRule,
			},
			wantErr: true,
			want: &Rules{
				{
					Group:    "rbac.authorization.k8s.io",
					Resource: "clusterroles",
					Verbs:    defaultResourceVerbs(),
				},
			},
		},
		{
			name:  "resource with invalid rules should return error and add the regular resource",
			rules: &Rules{},
			args: args{
				manifest: invalidRules,
			},
			wantErr: true,
			want: &Rules{
				{
					Group:    "rbac.authorization.k8s.io",
					Resource: "clusterroles",
					Verbs:    defaultResourceVerbs(),
				},
			},
		},
		{
			name:  "add regular resource rule",
			rules: &Rules{},
			args: args{
				manifest: resource,
			},
			wantErr: false,
			want: &Rules{
				{
					Group:    "core",
					Resource: "services",
					Verbs:    defaultResourceVerbs(),
				},
			},
		},
		{
			name:  "add non resource role rule",
			rules: &Rules{},
			args: args{
				manifest: nonResourceRoleRule,
			},
			wantErr: false,
			want: &Rules{
				{
					Group:    "rbac.authorization.k8s.io",
					Resource: "clusterroles",
					Verbs:    defaultResourceVerbs(),
				},
				{
					URLs:  []string{"/metrics"},
					Verbs: []string{"get"},
				},
			},
		},
		{
			name:  "add resource role rule",
			rules: &Rules{},
			args: args{
				manifest: resourceRoleRule,
			},
			wantErr: false,
			want: &Rules{
				{
					Group:    "rbac.authorization.k8s.io",
					Resource: "clusterroles",
					Verbs:    defaultResourceVerbs(),
				},
				{
					Group:    "group",
					Resource: "services",
					Verbs:    []string{"get"},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.rules.addForResource(tt.args.manifest); (err != nil) != tt.wantErr {
				t.Errorf("Rules.addForManifest() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, tt.rules)
		})
	}
}
