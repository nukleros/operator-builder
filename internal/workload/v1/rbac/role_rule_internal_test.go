// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package rbac

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func NewTestRoleRule() *RoleRule {
	return &RoleRule{
		Groups:    []string{"apps"},
		Resources: []string{"DaemonSets", "Deployments"},
		Verbs:     []string{"get", "patch"},
	}
}

func TestRoleRule_addTo(t *testing.T) {
	t.Parallel()

	testRule := NewTestRule()

	type fields struct {
		Groups    []string
		Resources []string
		URLs      []string
		Verbs     []string
	}

	type args struct {
		rules *Rules
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Rules
	}{
		{
			name: "ensure new role rule is added properly",
			fields: fields{
				Groups:    []string{"newGroup"},
				Resources: []string{"newResources"},
				Verbs:     []string{"test"},
			},
			args: args{
				rules: &Rules{},
			},
			want: &Rules{
				{
					Group:    "newGroup",
					Resource: "newresources",
					Verbs:    []string{"test"},
				},
			},
		},
		{
			name: "ensure existing rule is not added",
			fields: fields{
				Groups:    []string{testRule.Group},
				Resources: []string{testRule.Resource},
				Verbs:     testRule.Verbs,
			},
			args: args{
				rules: NewTestRules(),
			},
			want: NewTestRules(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rule := &RoleRule{
				Groups:    tt.fields.Groups,
				Resources: tt.fields.Resources,
				URLs:      tt.fields.URLs,
				Verbs:     tt.fields.Verbs,
			}
			rule.addTo(tt.args.rules)
			assert.Equal(t, tt.want, tt.args.rules)
		})
	}
}

func TestRoleRuleField_setValues(t *testing.T) {
	t.Parallel()

	rule := map[string]interface{}{
		"apiGroups": []string{
			"one",
			"two",
			"three",
		},
	}

	invalid := map[string]interface{}{
		"is": "invalid",
	}

	type args struct {
		rule     interface{}
		fieldKey string
	}

	tests := []struct {
		name    string
		field   *RoleRuleField
		args    args
		wantErr bool
		want    *RoleRuleField
	}{
		{
			name:  "ensure field is set appropriately without an error",
			field: &RoleRuleField{},
			args: args{
				rule:     rule,
				fieldKey: "apiGroups",
			},
			wantErr: false,
			want: &RoleRuleField{
				"one",
				"two",
				"three",
			},
		},
		{
			name:  "ensure missing key returns no error with no changes",
			field: &RoleRuleField{},
			args: args{
				rule:     rule,
				fieldKey: "missing",
			},
			wantErr: false,
			want:    &RoleRuleField{},
		},
		{
			name:  "ensure invalid rule returns an error with no changes",
			field: &RoleRuleField{},
			args: args{
				rule:     invalid,
				fieldKey: "is",
			},
			wantErr: true,
			want:    &RoleRuleField{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.field.setValues(tt.args.rule, tt.args.fieldKey); (err != nil) != tt.wantErr {
				t.Errorf("RoleRuleField.setValues() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, tt.field)
		})
	}
}

func TestRoleRule_processRaw(t *testing.T) {
	t.Parallel()

	rule := map[string][]interface{}{
		"apiGroups": {
			"dog",
			"cat",
		},
		"resources": {
			"toy",
			"treat",
		},
		"verbs": {
			"bark",
			"meow",
		},
		"nonResourceURLs": {
			"/dog",
			"/cat",
		},
	}

	invalid := map[string]interface{}{
		"apiGroups": "group",
	}

	type fields struct {
		Groups    RoleRuleField
		Resources RoleRuleField
		Verbs     RoleRuleField
		URLs      RoleRuleField
	}

	type args struct {
		rule interface{}
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    *RoleRule
	}{
		{
			name: "ensure valid rule is set appropriately without error",
			args: args{
				rule: rule,
			},
			wantErr: false,
			want: &RoleRule{
				Groups: RoleRuleField{
					"dog",
					"cat",
				},
				Resources: RoleRuleField{
					"toy",
					"treat",
				},
				Verbs: RoleRuleField{
					"bark",
					"meow",
				},
				URLs: RoleRuleField{
					"/dog",
					"/cat",
				},
			},
		},
		{
			name: "ensure invalid rule returns error and does not change role rule",
			args: args{
				rule: invalid,
			},
			wantErr: true,
			want:    &RoleRule{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			roleRule := &RoleRule{
				Groups:    tt.fields.Groups,
				Resources: tt.fields.Resources,
				Verbs:     tt.fields.Verbs,
				URLs:      tt.fields.URLs,
			}
			if err := roleRule.processRaw(tt.args.rule); (err != nil) != tt.wantErr {
				t.Errorf("RoleRule.processRaw() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, roleRule)
		})
	}
}

func TestRoleRule_toRules(t *testing.T) {
	t.Parallel()

	resourceRoleRule := &RoleRule{
		Groups: RoleRuleField{
			"core",
		},
		Resources: RoleRuleField{
			"dogs",
		},
		Verbs: RoleRuleField{
			"pet",
		},
	}

	nonResourceRoleRule := &RoleRule{
		URLs: RoleRuleField{
			"/bark",
		},
		Verbs: RoleRuleField{
			"pet",
		},
	}

	type fields struct {
		Groups    RoleRuleField
		Resources RoleRuleField
		Verbs     RoleRuleField
		URLs      RoleRuleField
	}

	tests := []struct {
		name      string
		fields    fields
		wantRules *Rules
	}{
		{
			name: "role rule without verbs returns empty rules",
			fields: fields{
				Verbs: RoleRuleField{},
			},
			wantRules: &Rules{},
		},
		{
			name: "invalid role rule returns empty rules",
			fields: fields{
				Groups: resourceRoleRule.Groups,
				Verbs:  resourceRoleRule.Verbs,
			},
			wantRules: &Rules{},
		},
		{
			name: "resource role rule returns resource rule",
			fields: fields{
				Groups:    resourceRoleRule.Groups,
				Resources: resourceRoleRule.Resources,
				Verbs:     resourceRoleRule.Verbs,
			},
			wantRules: &Rules{
				{
					Group:    "core",
					Resource: "dogs",
					Verbs:    []string{"pet"},
				},
			},
		},
		{
			name: "non-resource role rule returns non-resource rule",
			fields: fields{
				URLs:  nonResourceRoleRule.URLs,
				Verbs: nonResourceRoleRule.Verbs,
			},
			wantRules: &Rules{
				{
					URLs:  []string{"/bark"},
					Verbs: []string{"pet"},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			roleRule := &RoleRule{
				Groups:    tt.fields.Groups,
				Resources: tt.fields.Resources,
				Verbs:     tt.fields.Verbs,
				URLs:      tt.fields.URLs,
			}
			if gotRules := roleRule.toRules(); !reflect.DeepEqual(gotRules, tt.wantRules) {
				t.Errorf("RoleRule.toRules() = %v, want %v", gotRules, tt.wantRules)
			}
		})
	}
}
