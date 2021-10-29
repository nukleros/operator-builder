// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import "strings"

// OwnershipRule contains the info needed to create the controller ownership
// functionality when setting up the controller with the manager.  This allows
// the controller to reconcile the state of a deleted resource that it manages.
type OwnershipRule struct {
	Version string
	Kind    string
	CoreAPI bool
}

type OwnershipRules []OwnershipRule

func (or *OwnershipRules) addOrUpdateOwnership(version, kind, group string) {
	if or == nil {
		or = &OwnershipRules{}
	}
	// determine group and kind for ownership rule generation
	newOwnershipRule := OwnershipRule{
		Version: version,
		Kind:    kind,
		CoreAPI: isCoreAPI(group),
	}

	if !or.versionKindRecorded(&newOwnershipRule) {
		*or = append(*or, newOwnershipRule)
	}
}

func (or *OwnershipRules) versionKindRecorded(newOwnershipRule *OwnershipRule) bool {
	for _, r := range *or {
		if r.Version == newOwnershipRule.Version && r.Kind == newOwnershipRule.Kind {
			return true
		}
	}

	return false
}

func coreAPIs() []string {
	return []string{
		"apps", "batch", "autoscaling", "extensions", "policy",
	}
}

func isCoreAPI(group string) bool {
	// return if the group is missing or labeled as core
	if group == "" || group == "core" {
		return true
	}

	// return if the group contains the kubernetes api group strings
	if strings.Contains(group, "k8s.io") || strings.Contains(group, "kubernetes.io") {
		return true
	}

	// loop through known groups and return true if found
	for _, coreGroup := range coreAPIs() {
		if group == coreGroup {
			return true
		}
	}

	return false
}
