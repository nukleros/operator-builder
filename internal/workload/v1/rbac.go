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
	if rs == nil {
		rs = &RBACRules{}
	}

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

func getResourceForRBAC(kind string) string {
	rbacResource := strings.Split(kind, "/")

	if rbacResource[0] == "*" {
		kind = "*"
	} else {
		kind = resource.RegularPlural(rbacResource[0])
	}

	if len(rbacResource) > 1 {
		kind = fmt.Sprintf("%s/%s", kind, rbacResource[1])
	}

	return kind
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

func (rs *RBACRules) addRulesForManifest(kind, group string, rawContent interface{}) {
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
		resourceRules := valueFromInterface(rawContent, "rules")
		if resourceRules == nil {
			return
		}

		for _, resourceRule := range resourceRules.([]interface{}) {
			rbacGroups := valueFromInterface(resourceRule, "apiGroups")
			rbacKinds := valueFromInterface(resourceRule, "resources")
			rbacVerbs := valueFromInterface(resourceRule, "verbs")

			// assign a new rule for each group and kind match
			if rbacGroups == nil {
				continue
			}

			for _, rbacGroup := range rbacGroups.([]interface{}) {
				if rbacKinds == nil {
					continue
				}

				for _, rbacKind := range rbacKinds.([]interface{}) {
					if rbacVerbs == nil {
						continue
					}
					// gather verbs and convert to strings
					var verbs []string
					for _, verb := range rbacVerbs.([]interface{}) {
						verbs = append(verbs, verb.(string))
					}

					rs.AddOrUpdateRules(
						&RBACRule{
							Group:    rbacGroupFromGroup(rbacGroup.(string)),
							Resource: getResourceForRBAC(rbacKind.(string)),
							Verbs:    verbs,
						},
					)
				}
			}
		}
	}
}

func defaultResourceVerbs() []string {
	return []string{
		"get", "list", "watch", "create", "update", "patch", "delete",
	}
}
