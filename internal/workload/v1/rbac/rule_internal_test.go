// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func NewTestRule() *Rule {
	return &Rule{
		Group:    "core",
		Resource: "exampleresources",
		Verbs:    []string{"get", "patch"},
	}
}

func NewTestNonResourceRule() *Rule {
	return &Rule{
		URLs:  []string{"/metrics"},
		Verbs: []string{"get", "patch"},
	}
}

func NewTestRules() *Rules {
	testRule := NewTestRule()
	testNonResourceRule := NewTestNonResourceRule()
	testRules := Rules{}
	testRules = append(testRules, *testRule, *testNonResourceRule)

	return &testRules
}

func TestRule_ToMarker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rule *Rule
		want string
	}{
		{
			name: "ensure resource rbac marker returns as expected",
			rule: NewTestRule(),
			want: "// +kubebuilder:rbac:groups=core,resources=exampleresources,verbs=get;patch",
		},
		{
			name: "ensure non-resource rbac marker returns as expected",
			rule: NewTestNonResourceRule(),
			want: "// +kubebuilder:rbac:verbs=get;patch,urls=/metrics",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rule := tt.rule
			if got := rule.ToMarker(); got != tt.want {
				t.Errorf("Rule.ToMarker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRule_addTo(t *testing.T) {
	t.Parallel()

	testRule := NewTestRule()
	testNonResourceRule := NewTestNonResourceRule()

	type fields struct {
		Group    string
		Resource string
		URLs     []string
		Verbs    []string
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
			name: "ensure new rule is added properly",
			fields: fields{
				Group:    "newGroup",
				Resource: "newResource",
				Verbs:    []string{"test"},
			},
			args: args{
				rules: &Rules{},
			},
			want: &Rules{
				{
					Group:    "newGroup",
					Resource: "newResource",
					Verbs:    []string{"test"},
				},
			},
		},
		{
			name: "ensure new non-resource rule is added properly",
			fields: fields{
				URLs:  []string{"yes"},
				Verbs: []string{"test"},
			},
			args: args{
				rules: &Rules{},
			},
			want: &Rules{
				{
					URLs:  []string{"yes"},
					Verbs: []string{"test"},
				},
			},
		},
		{
			name: "ensure existing rule is not added",
			fields: fields{
				Group:    testRule.Group,
				Resource: testRule.Resource,
				Verbs:    testRule.Verbs,
			},
			args: args{
				rules: NewTestRules(),
			},
			want: NewTestRules(),
		},
		{
			name: "ensure existing non-resource rule is not added",
			fields: fields{
				URLs:  testNonResourceRule.URLs,
				Verbs: testNonResourceRule.Verbs,
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
			rule := &Rule{
				Group:    tt.fields.Group,
				Resource: tt.fields.Resource,
				URLs:     tt.fields.URLs,
				Verbs:    tt.fields.Verbs,
			}
			rule.addTo(tt.args.rules)
			assert.Equal(t, tt.want, tt.args.rules)
		})
	}
}

func TestRule_addResourceRuleTo(t *testing.T) {
	t.Parallel()

	testRule := NewTestRule()

	type fields struct {
		Group    string
		Resource string
		URLs     []string
		Verbs    []string
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
			name: "ensure new rule is added properly",
			fields: fields{
				Group:    "newGroup",
				Resource: "newResource",
				Verbs:    []string{"test"},
			},
			args: args{
				rules: &Rules{},
			},
			want: &Rules{
				{
					Group:    "newGroup",
					Resource: "newResource",
					Verbs:    []string{"test"},
				},
			},
		},
		{
			name: "ensure new rule with rule is added properly",
			fields: fields{
				Group:    "newGroup",
				Resource: "newResource",
				Verbs:    []string{"test"},
			},
			args: args{
				rules: &Rules{
					*NewTestRule(),
				},
			},
			want: &Rules{
				*NewTestRule(),
				{
					Group:    "newGroup",
					Resource: "newResource",
					Verbs:    []string{"test"},
				},
			},
		},
		{
			name: "ensure existing rule is not added",
			fields: fields{
				Group:    testRule.Group,
				Resource: testRule.Resource,
				Verbs:    testRule.Verbs,
			},
			args: args{
				rules: NewTestRules(),
			},
			want: NewTestRules(),
		},
		{
			name: "ensure existing rule with new verbs are appended appropriately",
			fields: fields{
				Group:    testRule.Group,
				Resource: testRule.Resource,
				Verbs:    []string{"existing"},
			},
			args: args{
				&Rules{
					{
						Group:    testRule.Group,
						Resource: testRule.Resource,
						Verbs:    testRule.Verbs,
					},
				},
			},
			want: &Rules{
				{
					Group:    testRule.Group,
					Resource: testRule.Resource,
					Verbs:    []string{"get", "patch", "existing"},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rule := &Rule{
				Group:    tt.fields.Group,
				Resource: tt.fields.Resource,
				URLs:     tt.fields.URLs,
				Verbs:    tt.fields.Verbs,
			}
			rule.addResourceRuleTo(tt.args.rules)
			assert.Equal(t, tt.want, tt.args.rules)
		})
	}
}

func TestRule_addNonResourceRuleTo(t *testing.T) {
	t.Parallel()

	testNonResourceRule := NewTestNonResourceRule()

	type fields struct {
		Group    string
		Resource string
		URLs     []string
		Verbs    []string
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
			name: "ensure new non-resource rule is added properly",
			fields: fields{
				URLs:  []string{"yes"},
				Verbs: []string{"test"},
			},
			args: args{
				rules: &Rules{},
			},
			want: &Rules{
				{
					URLs:  []string{"yes"},
					Verbs: []string{"test"},
				},
			},
		},
		{
			name: "ensure new non-resource rule with rule is added properly",
			fields: fields{
				URLs:  []string{"/existing"},
				Verbs: []string{"test"},
			},
			args: args{
				rules: &Rules{
					*NewTestRule(),
				},
			},
			want: &Rules{
				*NewTestRule(),
				{
					URLs:  []string{"/existing"},
					Verbs: []string{"test"},
				},
			},
		},
		{
			name: "ensure existing non-resource rule is not added",
			fields: fields{
				URLs:  testNonResourceRule.URLs,
				Verbs: testNonResourceRule.Verbs,
			},
			args: args{
				rules: NewTestRules(),
			},
			want: NewTestRules(),
		},
		{
			name: "ensure existing non-resulte rule with new urls are appended appropriately",
			fields: fields{
				URLs:  []string{"/new"},
				Verbs: testNonResourceRule.Verbs,
			},
			args: args{
				&Rules{
					{
						URLs:  testNonResourceRule.URLs,
						Verbs: testNonResourceRule.Verbs,
					},
				},
			},
			want: &Rules{
				{
					URLs:  []string{"/metrics"},
					Verbs: testNonResourceRule.Verbs,
				},
				{
					URLs:  []string{"/new"},
					Verbs: testNonResourceRule.Verbs,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rule := &Rule{
				Group:    tt.fields.Group,
				Resource: tt.fields.Resource,
				URLs:     tt.fields.URLs,
				Verbs:    tt.fields.Verbs,
			}
			rule.addNonResourceRuleTo(tt.args.rules)
			assert.Equal(t, tt.want, tt.args.rules)
		})
	}
}

func TestRule_addVerb(t *testing.T) {
	t.Parallel()

	type args struct {
		verb string
	}

	tests := []struct {
		name string
		rbac *Rule
		args args
		want []string
	}{
		{
			name: "ensure new verb is appropriately added",
			rbac: NewTestRule(),
			args: args{
				verb: "delete",
			},
			want: []string{"get", "patch", "delete"},
		},
		{
			name: "ensure existing verb is not added",
			rbac: NewTestRule(),
			args: args{
				verb: "get",
			},
			want: []string{"get", "patch"},
		},
		{
			name: "ensure new verb is appropriately added to a non-resource rule",
			rbac: NewTestNonResourceRule(),
			args: args{
				verb: "put",
			},
			want: []string{"get", "patch", "put"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := tt.rbac
			r.addVerb(tt.args.verb)

			assert.Equal(t, r.Verbs, tt.want)
		})
	}
}

func TestRule_groupResourceEqual(t *testing.T) {
	t.Parallel()

	type fields struct {
		Group    string
		Resource string
		URLs     []string
		Verbs    []string
	}

	type args struct {
		compared *Rule
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "ensure rule with equal group and resource returns true",
			fields: fields{
				Group:    "core",
				Resource: "exampleresources",
			},
			args: args{
				compared: NewTestRule(),
			},
			want: true,
		},
		{
			name: "ensure rule with not equal group and resource returns false",
			fields: fields{
				Group:    "coreFake",
				Resource: "exampleResourceFake",
			},
			args: args{
				compared: NewTestRule(),
			},
			want: false,
		},
		{
			name: "ensure rule with not equal group returns false",
			fields: fields{
				Group:    "coreFake",
				Resource: "exampleresources",
			},
			args: args{
				compared: NewTestRule(),
			},
			want: false,
		},
		{
			name: "ensure rule with not equal resource returns false",
			fields: fields{
				Group:    "core",
				Resource: "exampleResourceFake",
			},
			args: args{
				compared: NewTestRule(),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rule := &Rule{
				Group:    tt.fields.Group,
				Resource: tt.fields.Resource,
				URLs:     tt.fields.URLs,
				Verbs:    tt.fields.Verbs,
			}
			if got := rule.groupResourceEqual(tt.args.compared); got != tt.want {
				t.Errorf("Rule.groupResourceEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRule_isResourceRule(t *testing.T) {
	t.Parallel()

	testRule := NewTestRule()
	testNonResourceRule := NewTestNonResourceRule()

	type fields struct {
		Group    string
		Resource string
		URLs     []string
		Verbs    []string
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "ensure non resource rule returns false",
			fields: fields{
				URLs:  testNonResourceRule.URLs,
				Verbs: testNonResourceRule.Verbs,
			},
			want: false,
		},
		{
			name: "ensure non resource rule returns false",
			fields: fields{
				Group:    testRule.Group,
				Resource: testRule.Resource,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rule := &Rule{
				Group:    tt.fields.Group,
				Resource: tt.fields.Resource,
				URLs:     tt.fields.URLs,
				Verbs:    tt.fields.Verbs,
			}
			if got := rule.isResourceRule(); got != tt.want {
				t.Errorf("Rule.isResourceRule() = %v, want %v", got, tt.want)
			}
		})
	}
}
