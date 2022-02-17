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
	Group    string
	Resource string
	Verbs    []string
	URLs     []string
}

// RBACRoleRule contains the info needed to create the kubebuilder:rbac markers
// in the controller when a resource that is of a role or clusterrole type is
// found.  This is because the underlying controller needs the same permissions
// for the role or clusterrole that it is attempting to manage.
type RBACRoleRule struct {
	Groups    RBACRoleRuleField
	Resources RBACRoleRuleField
	Verbs     RBACRoleRuleField
	URLs      RBACRoleRuleField
}

type RBACRoleRuleField []string

type RBACRules []RBACRule

func (r *RBACRule) addVerb(verb string) {
	var found bool

	for _, existingVerb := range r.Verbs {
		if existingVerb == verb {
			found = true

			break
		}
	}

	if !found {
		r.Verbs = append(r.Verbs, verb)
	}
}

func rbacGroupFromGroup(group string) string {
	if group == "" {
		return coreRBACGroup
	}

	return group
}

func rbacFieldsToString(verbs []string) string {
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

func (rs *RBACRules) hasURL(url string) bool {
	for _, rule := range *rs {
		if len(rule.URLs) == 0 {
			continue
		}

		for i := range rule.URLs {
			if rule.URLs[i] == url {
				return true
			}
		}
	}

	return false
}

func (r *RBACRule) ToMarker() string {
	const kubebuilderPrefix = "// +kubebuilder:rbac"

	if len(r.URLs) > 0 {
		return fmt.Sprintf("%s:verbs=%s,urls=%s",
			kubebuilderPrefix,
			rbacFieldsToString(r.Verbs),
			rbacFieldsToString(r.URLs),
		)
	}

	return fmt.Sprintf("%s:groups=%s,resources=%s,verbs=%s",
		kubebuilderPrefix,
		r.Group,
		r.Resource,
		rbacFieldsToString(r.Verbs),
	)
}

func (rs *RBACRules) AddOrUpdateRules(newRules ...*RBACRule) {
	for i := range newRules {
		switch {
		case newRules[i].hasGroupResource():
			rs.addForGroupResource(newRules[i])
		case newRules[i].hasURLs():
			rs.addForURLs(newRules[i])
		default:
			continue
		}
	}
}

func (rs *RBACRules) addForGroupResource(newRule *RBACRule) {
	rules := *rs

	if !rules.groupResourceRecorded(newRule) {
		*rs = append(*rs, *newRule)
	} else {
		for i := range rules {
			if rules[i].groupResourceEqual(newRule) {
				for _, verb := range newRule.Verbs {
					rules[i].addVerb(verb)
				}
			}
		}
	}
}

func (rs *RBACRules) addForURLs(newRule *RBACRule) {
	rules := *rs

	for _, url := range newRule.URLs {
		for i := range rules {
			if rs.hasURL(url) {
				for _, verb := range newRule.Verbs {
					rules[i].addVerb(verb)
				}
			} else {
				*rs = append(*rs, *newRule)
			}
		}
	}
}

func (r *RBACRule) hasGroupResource() bool {
	return r.Group != "" && r.Resource != ""
}

func (r *RBACRule) hasURLs() bool {
	return len(r.URLs) > 0
}

func (rs *RBACRules) AddOrUpdateRoleRules(newRule *RBACRoleRule) {
	// we must have verbs to create our rbac
	if len(newRule.Verbs) == 0 {
		return
	}

	// we either need to have groups/resources or urls
	if len(newRule.Groups) == 0 || len(newRule.Resources) == 0 {
		if len(newRule.URLs) == 0 {
			return
		}
	}

	// assign a new rule for each group and kind match
	for _, rbacGroup := range newRule.Groups {
		for _, rbacKind := range newRule.Resources {
			rs.AddOrUpdateRules(
				&RBACRule{
					Group:    rbacGroupFromGroup(rbacGroup),
					Resource: getResourceForRBAC(rbacKind),
					Verbs:    newRule.Verbs,
					URLs:     newRule.URLs,
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
	case map[interface{}][]interface{}:
		out = asType[key]
	case map[string][]interface{}:
		out = asType[key]
	}

	return out
}

func (rs *RBACRules) addRuleForWorkload(workload WorkloadAPIBuilder, forCollection bool) {
	var verbs []string

	if forCollection {
		verbs = []string{"get", "list", "watch"}
	} else {
		verbs = defaultResourceVerbs()
	}

	// add permissions for the controller to be able to watch itself and update its own status
	rs.AddOrUpdateRules(
		&RBACRule{
			Group:    fmt.Sprintf("%s.%s", workload.GetAPIGroup(), workload.GetDomain()),
			Resource: getResourceForRBAC(workload.GetAPIKind()),
			Verbs:    verbs,
		},
		&RBACRule{
			Group:    fmt.Sprintf("%s.%s", workload.GetAPIGroup(), workload.GetDomain()),
			Resource: fmt.Sprintf("%s/status", getResourceForRBAC(workload.GetAPIKind())),
			Verbs:    defaultStatusVerbs(),
		},
	)
}

func (rs *RBACRules) addRulesForWorkload(workload WorkloadAPIBuilder) {
	rs.addRuleForWorkload(workload, false)

	if workload.IsComponent() {
		rs.addRuleForWorkload(workload.GetCollection(), true)
	}
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
			if err := rule.processRawRoleRule(rbacRoleRule); err != nil {
				return fmt.Errorf("%w; error processing rbac role rule %v", err, rbacRoleRule)
			}

			rs.AddOrUpdateRoleRules(rule)
		}
	}

	return nil
}

func (roleRule *RBACRoleRule) processRawRoleRule(rule interface{}) error {
	fields := map[*RBACRoleRuleField]string{
		&roleRule.Groups:    "apiGroups",
		&roleRule.Resources: "resources",
		&roleRule.Verbs:     "verbs",
		&roleRule.URLs:      "nonResourceURLs",
	}

	for objectField, fieldKey := range fields {
		if err := objectField.setRbacRoleRuleField(rule, fieldKey); err != nil {
			return fmt.Errorf("%w; error processing raw fule %v", err, rule)
		}
	}

	return nil
}

func (field *RBACRoleRuleField) setRbacRoleRuleField(rule interface{}, fieldKey string) error {
	fieldValue := valueFromInterface(rule, fieldKey)
	if fieldValue == nil {
		return nil
	}

	fieldValues, err := toArrayString(fieldValue)
	if err != nil {
		return fmt.Errorf("%w; error converting rbac field key %s for rule %v", err, fieldKey, rule)
	}

	*field = fieldValues

	return nil
}

func defaultStatusVerbs() []string {
	return []string{
		"get", "update", "patch",
	}
}

func defaultResourceVerbs() []string {
	return []string{
		"get", "list", "watch", "create", "update", "patch", "delete",
	}
}
