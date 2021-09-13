// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"
	"strings"
)

const (
	coreRBACGroup = "core"
)

func defaultResourceVerbs() []string {
	return []string{
		"get", "list", "watch", "create", "update", "patch", "delete",
	}
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

func extractManifests(manifestContent []byte) []string {
	var manifests []string

	lines := strings.Split(string(manifestContent), "\n")

	var manifest string

	for _, line := range lines {
		if strings.TrimRight(line, " ") == "---" {
			if len(manifest) > 0 {
				manifests = append(manifests, manifest)
				manifest = ""
			}
		} else {
			manifest = manifest + "\n" + line
		}
	}

	if len(manifest) > 0 {
		manifests = append(manifests, manifest)
	}

	return manifests
}

func versionKindRecorded(ownershipRules *[]OwnershipRule, newOwnershipRule *OwnershipRule) bool {
	for _, r := range *ownershipRules {
		if r.Version == newOwnershipRule.Version && r.Kind == newOwnershipRule.Kind {
			return true
		}
	}

	return false
}

func versionGroupFromAPIVersion(apiVersion string) (version, group string) {
	apiVersionElements := strings.Split(apiVersion, "/")

	if len(apiVersionElements) == 1 {
		version = apiVersionElements[0]
		group = coreRBACGroup
	} else {
		version = apiVersionElements[1]
		group = rbacGroupFromGroup(apiVersionElements[0])
	}

	return version, group
}

func getFuncNames(sourceFiles []SourceFile) (createFuncNames, initFuncNames []string) {
	for _, sourceFile := range sourceFiles {
		for _, childResource := range sourceFile.Children {
			funcName := fmt.Sprintf("Create%s", childResource.UniqueName)

			if strings.EqualFold(childResource.Kind, "customresourcedefinition") {
				initFuncNames = append(initFuncNames, funcName)
			}

			createFuncNames = append(createFuncNames, funcName)
		}
	}

	return createFuncNames, initFuncNames
}
