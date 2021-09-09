package v1

import (
	"fmt"
	"strings"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

func (rule *RBACRule) AddVerb(verb string) {
	var found bool

	for _, existingVerb := range rule.Verbs {
		if existingVerb == verb {
			found = true

			break
		}
	}

	if !found {
		rule.Verbs = append(rule.Verbs, verb)
		rule.VerbString = rbacVerbsToString(rule.Verbs)
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

func groupResourceEqual(rbacRule, newRBACRule *RBACRule) bool {
	if rbacRule.Group == newRBACRule.Group && rbacRule.Resource == newRBACRule.Resource {
		return true
	}

	return false
}

func groupResourceRecorded(rbacRules *[]RBACRule, newRBACRule *RBACRule) bool {
	for _, r := range *rbacRules {
		r := r
		if groupResourceEqual(&r, newRBACRule) {
			return true
		}
	}

	return false
}

func rbacRulesAddOrUpdate(rbacRules *[]RBACRule, newRule *RBACRule) {
	if !groupResourceRecorded(rbacRules, newRule) {
		newRule.VerbString = rbacVerbsToString(newRule.Verbs)
		*rbacRules = append(*rbacRules, *newRule)
	} else {
		rules := *rbacRules
		for i := range rules {
			if groupResourceEqual(&rules[i], newRule) {
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
		kind = utils.PluralizeKind(rbacResource[0])
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

func rbacRulesForManifest(kind, group string, rawContent interface{}, rbacRules *[]RBACRule) {
	rbacRulesAddOrUpdate(
		rbacRules,
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

					rbacRulesAddOrUpdate(
						rbacRules,
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
