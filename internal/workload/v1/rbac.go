// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
)

const (
	coreRBACGroup = "core"
)

// RBACRule contains the info needed to create the kubebuilder:rbac markers in
// the controller.
type RBACRule struct {
	Group      string
	Resource   string
	Verbs      []string
	VerbString string
}

// RBACRoleRule contains the info needed to create the kubebuilder:rbac markers
// in the controller when a resource that is of a role or clusterrole type is
// found.  This is because the underlying controller needs the same permissions
// for the role or clusterrole that it is attempting to manage.
type RBACRoleRule struct {
	Groups    []string
	Resources []string
	Verbs     []string
}

type RBACRules []RBACRule

func (r *RBACRule) AddVerb(verb string) {
	var found bool

	for _, existingVerb := range r.Verbs {
		if existingVerb == verb {
			found = true

			break
		}
	}

	if !found {
		r.Verbs = append(r.Verbs, verb)
		r.VerbString = rbacVerbsToString(r.Verbs)
	}
}

func rbacGroupFromGroup(group string) string {
	if group == "" {
		return coreRBACGroup
	}

	return group
}

func rbacVerbsToString(verbs []string) string {
	return strings.Join(verbs, ";")
}

func (r *RBACRule) groupResourceEqual(newRBACRule *RBACRule) bool {
	if r.Group == newRBACRule.Group && r.Resource == newRBACRule.Resource {
		return true
	}

	return false
}

func (rs *RBACRules) groupResourceRecorded(newRBACRule *RBACRule) bool {
	if rs == nil {
		return false
	}

	for _, r := range *rs {
		r := r
		if r.groupResourceEqual(newRBACRule) {
			return true
		}
	}

	return false
}

func (rs *RBACRules) AddOrUpdateRules(newRule *RBACRule) {
	if !rs.groupResourceRecorded(newRule) {
		newRule.VerbString = rbacVerbsToString(newRule.Verbs)
		*rs = append(*rs, *newRule)
	} else {
		rules := *rs
		for i := range rules {
			if rules[i].groupResourceEqual(newRule) {
				for _, verb := range newRule.Verbs {
					rules[i].AddVerb(verb)
				}
			}
		}
	}
}

func (rs *RBACRules) AddOrUpdateRoleRules(newRule *RBACRoleRule) {
	// assign a new rule for each group and kind match
	if len(newRule.Groups) == 0 {
		return
	}

	for _, rbacGroup := range newRule.Groups {
		if len(newRule.Resources) == 0 {
			return
		}

		for _, rbacKind := range newRule.Resources {
			if len(newRule.Verbs) == 0 {
				return
			}

			rs.AddOrUpdateRules(
				&RBACRule{
					Group:    rbacGroupFromGroup(rbacGroup),
					Resource: getResourceForRBAC(rbacKind),
					Verbs:    newRule.Verbs,
				},
			)
		}
	}
}

func getResourceForRBAC(kind string) string {
	rbacResource := strings.Split(kind, "/")

	if rbacResource[0] == "*" {
		kind = "*"
	} else {
		kind = getPluralRBAC(rbacResource[0])
	}

	if len(rbacResource) > 1 {
		kind = fmt.Sprintf("%s/%s", kind, rbacResource[1])
	}

	return kind
}

// getPluralRBAC will transform known irregulars into a proper type for rbac
// rules.
func getPluralRBAC(kind string) string {
	pluralMap := map[string]string{
		"resourcequota": "resourcequotas",
	}
	plural := resource.RegularPlural(kind)

	if pluralMap[plural] != "" {
		return pluralMap[plural]
	}

	return plural
}

func valueFromInterface(in interface{}, key string) (out interface{}) {
	switch asType := in.(type) {
	case map[interface{}]interface{}:
		out = asType[key]
	case map[string]interface{}:
		out = asType[key]
	}

	return out
}

func (rs *RBACRules) addRulesForManifest(kind, group string, rawContent interface{}) error {
	rs.AddOrUpdateRules(
		&RBACRule{
			Group:    group,
			Resource: getResourceForRBAC(kind),
			Verbs:    defaultResourceVerbs(),
		},
	)

	// if we are working with roles and cluster roles, we must also grant rbac to the resources
	// which are managed by them
	if strings.EqualFold(kind, "clusterrole") || strings.EqualFold(kind, "role") {
		rules := valueFromInterface(rawContent, "rules")
		if rules == nil {
			return nil
		}

		rbacRoleRules, err := toArrayInterface(rules)
		if err != nil {
			return fmt.Errorf("%w; error converting resource rules %v", err, rules)
		}

		for _, rbacRoleRule := range rbacRoleRules {
			rule := &RBACRoleRule{}
			if err := rule.processRawRule(rbacRoleRule); err != nil {
				return fmt.Errorf("%w; error processing rbac role rule %v", err, rules)
			}

			rs.AddOrUpdateRoleRules(rule)
		}
	}

	return nil
}

func (roleRule *RBACRoleRule) processRawRule(rule interface{}) error {
	rbacGroups, err := toArrayString(valueFromInterface(rule, "apiGroups"))
	if err != nil {
		return fmt.Errorf("%w; error converting rbac groups for rule %v", err, rule)
	}

	rbacKinds, err := toArrayString(valueFromInterface(rule, "resources"))
	if err != nil {
		return fmt.Errorf("%w; error converting rbac kinds for rule %v", err, rule)
	}

	rbacVerbs, err := toArrayString(valueFromInterface(rule, "verbs"))
	if err != nil {
		return fmt.Errorf("%w; error converting rbac verbs for rule %v", err, rule)
	}

	roleRule.Groups = rbacGroups
	roleRule.Resources = rbacKinds
	roleRule.Verbs = rbacVerbs

	return nil
}

func defaultResourceVerbs() []string {
	return []string{
		"get", "list", "watch", "create", "update", "patch", "delete",
	}
}
