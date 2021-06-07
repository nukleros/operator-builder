package v1

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	pluralize "github.com/gertd/go-pluralize"
	"gopkg.in/yaml.v2"

	"github.com/vmware-tanzu-labs/object-code-generator-for-k8s/pkg/generate"
)

//func processResources(manifestFile, workloadPath string) (SourceFile, error) {
func processResources(workloadPath string, resources []string) (*[]SourceFile, *[]RBACRule, error) {

	// each sourceFile is a source code file that contains one or more child
	// resource definition
	var sourceFiles []SourceFile
	var rbacRules []RBACRule
	pluralize := pluralize.NewClient()

	for _, manifestFile := range resources {

		// determine sourceFile filename
		var sourceFile SourceFile
		sourceFile.Filename = strings.Replace(strings.Split(manifestFile, ".")[0]+".go", "-", "_", -1)

		var childResources []ChildResource

		// capture entire resource manifest file content
		manifestContent, err := ioutil.ReadFile(filepath.Join(filepath.Dir(workloadPath), manifestFile))
		if err != nil {
			return &[]SourceFile{}, &[]RBACRule{}, err
		}

		manifests := extractManifests(manifestContent)

		// generate a unique name for the resource using the kind and name
		resourceUniqueName := strings.Replace(strings.Title(resourceName), "-", "", -1)
		resourceUniqueName = strings.Replace(resourceUniqueName, ".", "", -1)
		resourceUniqueName = strings.Replace(resourceUniqueName, ":", "", -1)
		resourceUniqueName = fmt.Sprintf("%s%s", resourceKind, resourceUniqueName)

			// unmarshal yaml to get attributes
			var rawContent interface{}
			err = yaml.Unmarshal([]byte(manifest), &rawContent)
			if err != nil {
				return &[]SourceFile{}, &[]RBACRule{}, err
			}

			// determine resource kind and name
			resourceKind := fmt.Sprintf("%s", rawContent.(interface{}).(map[interface{}]interface{})["kind"])
			resourceName := fmt.Sprintf("%s", rawContent.(interface{}).(map[interface{}]interface{})["metadata"].(interface{}).(map[interface{}]interface{})["name"])

			// generate a unique name for the resource using the kind and name
			resourceUniqueName := strings.Replace(strings.Title(resourceName), "-", "", -1)
			resourceUniqueName = strings.Replace(resourceUniqueName, ".", "", -1)
			resourceUniqueName = fmt.Sprintf("%s%s", resourceKind, resourceUniqueName)

			// deteremine resource group and version
			apiVersion := fmt.Sprintf("%s", rawContent.(interface{}).(map[interface{}]interface{})["apiVersion"])
			apiVersionElements := strings.Split(apiVersion, "/")

			var resourceGroup string
			var resourceVersion string
			if len(apiVersionElements) == 1 {
				resourceGroup = "core"
				resourceVersion = apiVersionElements[0]
			} else {
				//resourceGroup = strings.Replace(apiVersionElements[0], ".k8s.io", "", -1)
				resourceGroup = apiVersionElements[0]
				resourceVersion = apiVersionElements[1]
			}

			// determine group and resource for RBAC rule generation
			resourcePlural := strings.ToLower(pluralize.Plural(resourceKind))
			newRBACRule := RBACRule{
				Group:    resourceGroup,
				Resource: resourcePlural,
			}
			exists := groupResourceRecorded(&rbacRules, &newRBACRule)
			if !exists {
				rbacRules = append(rbacRules, newRBACRule)
			}

			// generate object source code
			resourceDefinition, err := generate.Generate([]byte(manifest), "resourceObj")
			if err != nil {
				return &[]SourceFile{}, &[]RBACRule{}, err
			}

			// add variables based on commented markers
			resourceDefinition, err = addVariables(resourceDefinition)
			if err != nil {
				return &[]SourceFile{}, &[]RBACRule{}, err
			}

			resource := ChildResource{
				Name:          resourceName,
				UniqueName:    resourceUniqueName,
				Group:         resourceGroup,
				Version:       resourceVersion,
				Kind:          resourceKind,
				StaticContent: manifest,
				SourceCode:    resourceDefinition,
			}

			childResources = append(childResources, resource)
		}

		sourceFile.Children = childResources
		sourceFiles = append(sourceFiles, sourceFile)
	}

	return &sourceFiles, &rbacRules, nil
}

func extractManifests(manifestContent []byte) []string {

	var manifests []string

	lines := strings.Split(string(manifestContent), "\n")

	var manifest string
	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
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

func addVariables(resourceContent string) (string, error) {

	lines := strings.Split(string(resourceContent), "\n")
	for i, line := range lines {
		if containsMarker(line) {
			markedLine := processMarkedComments(line)
			lines[i] = markedLine
		}
	}

	return strings.Join(lines, "\n"), nil
}

func groupResourceRecorded(rbacRules *[]RBACRule, newRBACRule *RBACRule) bool {

	for _, r := range *rbacRules {
		if r.Group == newRBACRule.Group && r.Resource == newRBACRule.Resource {
			return true
		}
	}

	return false
}
