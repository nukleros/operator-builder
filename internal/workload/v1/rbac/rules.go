// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package rbac

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

type Rules []Rule

// add will add a set of new rules to an existing set of rules.
func (rules *Rules) Add(newRules ...rbacRuleProcessor) {
	for i := range newRules {
		newRules[i].addTo(rules)
	}
}

// addTo satisfies the rbacRuleProcessor interface by defining the logic that adds a rule into an
// existing set of rules.
func (rules *Rules) addTo(ruleSet *Rules) {
	rs := *rules

	for i := range *rules {
		rule := rs[i]

		ruleSet.Add(&rule)
	}
}

// addForWorkload will add a particular rule to a set of rules given a workload.
func (rules *Rules) addForWorkload(workload rbacWorkloadProcessor) {
	// workloadRule is a rule that creates rbac such that the controller can manage the
	// workload type in which it is responsible for reconciling
	workloadRule := &Rule{
		Group:    fmt.Sprintf("%s.%s", workload.GetAPIGroup(), workload.GetDomain()),
		Resource: getResource(workload.GetAPIKind()),
		Verbs:    defaultResourceVerbs(),
	}

	// statusRule is a rule that creates rbac such that the controller can manage its own
	// status updates for the workload that it is responsible for reconciling
	statusRule := &Rule{
		Group:    fmt.Sprintf("%s.%s", workload.GetAPIGroup(), workload.GetDomain()),
		Resource: fmt.Sprintf("%s/status", getResource(workload.GetAPIKind())),
		Verbs:    defaultStatusVerbs(),
	}

	rules.Add(workloadRule, statusRule)
}

// addForResource will add a particular rule given an unstructured manifest.
func (rules *Rules) addForResource(manifest *unstructured.Unstructured) error {
	kind := manifest.GetKind()

	rules.Add(
		&Rule{
			Group:    getGroup(manifest.GroupVersionKind().Group),
			Resource: getResource(kind),
			Verbs:    defaultResourceVerbs(),
		},
	)

	// if we are working with roles and cluster roles, we must also grant rbac to the resources
	// which are managed by them
	if strings.EqualFold(kind, "clusterrole") || strings.EqualFold(kind, "role") {
		roleRules := valueFromInterface(manifest.Object, "rules")
		if roleRules == nil {
			return nil
		}

		rbacRoleRules, err := utils.ToArrayInterface(roleRules)
		if err != nil {
			return fmt.Errorf("%w; error converting resource rules %v", err, roleRules)
		}

		for _, rbacRoleRule := range rbacRoleRules {
			rule := &RoleRule{}
			if err := rule.processRaw(rbacRoleRule); err != nil {
				return fmt.Errorf("%w; error processing rbac role rule %v", err, rbacRoleRule)
			}

			rules.Add(rule)
		}
	}

	return nil
}

// hasResourceRule determines if a set of rules has a rule which contains
// a specific group/resource combination.  A specific group/resource combination
// is used to guarantee uniqueness on a set of rules.
func (rules *Rules) hasResourceRule(rule *Rule) bool {
	for _, r := range *rules {
		if r.groupResourceEqual(rule) {
			return true
		}
	}

	return false
}

