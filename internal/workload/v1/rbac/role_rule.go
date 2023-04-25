// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package rbac

import (
	"errors"
	"fmt"

	"github.com/nukleros/operator-builder/internal/utils"
)

var (
	ErrorProcessRoleRule      = errors.New("error processing role rule")
	ErrorProcessRoleRuleField = errors.New("error processing role rule field")
)

// RoleRule contains the info needed to create the kubebuilder:rbac markers
// in the controller when a resource that is of a role or clusterrole type is
// found.  This is because the underlying controller needs the same permissions
// for the role or clusterrole that it is attempting to manage.
type RoleRule struct {
	Groups    RoleRuleField
	Resources RoleRuleField
	Verbs     RoleRuleField
	URLs      RoleRuleField
}

type RoleRuleField []string

// addTo satisfies the rbacRuleProcessor interface by defining the logic that adds a
// role rule into an existing set of rules.
func (roleRule *RoleRule) addTo(rules *Rules) {
	// convert the role rule into a set of rules
	newRules := roleRule.toRules()

	for _, rule := range *newRules {
		rule.addTo(rules)
	}
}

// processRaw will take in a raw interface and convert it into a role rule.
func (roleRule *RoleRule) processRaw(rule interface{}) error {
	fields := map[*RoleRuleField]string{
		&roleRule.Groups:    "apiGroups",
		&roleRule.Resources: "resources",
		&roleRule.Verbs:     "verbs",
		&roleRule.URLs:      "nonResourceURLs",
	}

	for objectField, fieldKey := range fields {
		if err := objectField.setValues(rule, fieldKey); err != nil {
			return fmt.Errorf("%w; %s: %v", err, ErrorProcessRoleRule.Error(), rule)
		}
	}

	return nil
}

// setValues of a field for a particular role rule.
func (field *RoleRuleField) setValues(rule interface{}, fieldKey string) error {
	fieldValue := valueFromInterface(rule, fieldKey)
	if fieldValue == nil {
		return nil
	}

	fieldValues, err := utils.ToArrayString(fieldValue)
	if err != nil {
		return fmt.Errorf("%w; %s: [%s]", err, ErrorProcessRoleRuleField.Error(), fieldKey)
	}

	*field = fieldValues

	return nil
}

// toRules will convert a role rule into a set of regular rules.
func (roleRule *RoleRule) toRules() *Rules {
	// we must have verbs to create our rbac
	if len(roleRule.Verbs) == 0 {
		return &Rules{}
	}

	// we either need to have groups/resources or urls
	if len(roleRule.Groups) > 0 && len(roleRule.Resources) > 0 {
		return roleRule.groupResourceRules()
	} else if len(roleRule.URLs) > 0 {
		return roleRule.nonResourceRules()
	}

	return &Rules{}
}

// groupResourceRules will return a set of rules given a role rule which contains
// both a group and a resource.
func (roleRule *RoleRule) groupResourceRules() *Rules {
	rules := &Rules{}

	// assign a new rule for each group and kind match
	for _, rbacGroup := range roleRule.Groups {
		for _, rbacKind := range roleRule.Resources {
			rule := &Rule{
				Group:    getGroup(rbacGroup),
				Resource: getResource(rbacKind),
				Verbs:    roleRule.Verbs,
				URLs:     roleRule.URLs,
			}

			rule.addResourceRuleTo(rules)
		}
	}

	return rules
}

// nonResourceRules will return a set of rules given a role rule which does not
// contain a group and a resource and instead contains non resource urls.
func (roleRule *RoleRule) nonResourceRules() *Rules {
	return &Rules{
		{
			Verbs: roleRule.Verbs,
			URLs:  roleRule.URLs,
		},
	}
}
