package v1

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/vmware-tanzu-labs/object-code-generator-for-k8s/pkg/generate"
	"gopkg.in/yaml.v2"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/utils"
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

func processResources(workloadPath string, resources []string) (*[]SourceFile, *[]RBACRule, *[]OwnershipRule, error) {
	// each sourceFile is a source code file that contains one or more child
	// resource definition
	sourceFiles := make([]SourceFile, len(resources))

	var rbacRules []RBACRule

	var ownershipRules []OwnershipRule

	for i, manifestFile := range resources {
		// determine sourceFile filename
		var sourceFile SourceFile
		sourceFile.Filename = filepath.Base(manifestFile)                // get filename from path
		sourceFile.Filename = strings.Split(sourceFile.Filename, ".")[0] // strip ".yaml"
		sourceFile.Filename += ".go"                                     // add correct file ext
		sourceFile.Filename = utils.ToFileName(sourceFile.Filename)      // kebab-case to snake_case

		var childResources []ChildResource

		// capture entire resource manifest file content
		manifestContent, err := ioutil.ReadFile(filepath.Join(filepath.Dir(workloadPath), manifestFile))
		if err != nil {
			return nil, nil, nil, err
		}

		manifests := extractManifests(manifestContent)

		for _, manifest := range manifests {
			// unmarshal yaml to get attributes
			var rawContent interface{}

			err = yaml.Unmarshal([]byte(manifest), &rawContent)
			if err != nil {
				return nil, nil, nil, err
			}

			// determine resource kind and name
			resourceKind := fmt.Sprintf("%s", rawContent.(map[interface{}]interface{})["kind"])
			resourceName := fmt.Sprintf("%s", rawContent.(map[interface{}]interface{})["metadata"].(map[interface{}]interface{})["name"])

			// generate a unique name for the resource using the kind and name
			resourceUniqueName := strings.Replace(strings.Title(resourceName), "-", "", -1)
			resourceUniqueName = strings.Replace(resourceUniqueName, ".", "", -1)
			resourceUniqueName = fmt.Sprintf("%s%s", resourceKind, resourceUniqueName)

			// determine resource group and version
			apiVersion := fmt.Sprintf("%s", rawContent.(map[interface{}]interface{})["apiVersion"])
			resourceVersion, resourceGroup := versionGroupFromAPIVersion(apiVersion)

			// determine group and resource for RBAC rule generation
			rbacRulesForManifest(resourceKind, resourceGroup, rawContent, &rbacRules)

			// determine group and kind for ownership rule generation
			newOwnershipRule := OwnershipRule{
				Version: apiVersion,
				Kind:    resourceKind,
				CoreAPI: isCoreAPI(resourceGroup),
			}

			ownershipExists := versionKindRecorded(&ownershipRules, &newOwnershipRule)
			if !ownershipExists {
				ownershipRules = append(ownershipRules, newOwnershipRule)
			}

			resource := ChildResource{
				Name:       resourceName,
				UniqueName: resourceUniqueName,
				Group:      resourceGroup,
				Version:    resourceVersion,
				Kind:       resourceKind,
			}

			// generate the object source code
			resourceDefinition, err := generate.Generate([]byte(manifest), "resourceObj")
			if err != nil {
				return nil, nil, nil, err
			}

			// add variables based on commented markers
			resourceDefinition = addVariables(resourceDefinition)

			// add the source code to the resource
			resource.SourceCode = resourceDefinition
			resource.StaticContent = manifest

			childResources = append(childResources, resource)
		}

		sourceFile.Children = childResources
		sourceFiles[i] = sourceFile
	}

	return &sourceFiles, &rbacRules, &ownershipRules, nil
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

func addVariables(resourceContent string) string {
	lines := strings.Split(resourceContent, "\n")
	for i, line := range lines {
		if containsMarker(line) {
			markedLine := processMarkedComments(line)
			lines[i] = markedLine
		}
	}

	return strings.Join(lines, "\n")
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
