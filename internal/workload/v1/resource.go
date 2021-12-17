// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
)

// SourceFile represents a golang source code file that contains one or more
// child resource objects.
type SourceFile struct {
	Filename  string
	Children  []ChildResource
	HasStatic bool
}

// ChildResource contains attributes for resources created by the custom resource.
// These definitions are inferred from the resource manifests.
type ChildResource struct {
	Name          string
	UniqueName    string
	Group         string
	Version       string
	Kind          string
	StaticContent string
	SourceCode    string
}

// Resource represents a single input manifest for a given config.
type Resource struct {
	FileName         string
	relativeFileName string
	Content          []byte
}

func (r *Resource) UnmarshalYAML(node *yaml.Node) error {
	r.FileName = node.Value

	return nil
}

func (r *Resource) loadManifest() error {
	manifestContent, err := os.ReadFile(r.FileName)
	if err != nil {
		return formatProcessError(r.FileName, err)
	}

	r.Content = manifestContent

	return nil
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

func determineSourceFileName(manifestFile string) SourceFile {
	var sourceFile SourceFile
	sourceFile.Filename = filepath.Clean(manifestFile)
	sourceFile.Filename = strings.ReplaceAll(sourceFile.Filename, "/", "_")                              // get filename from path
	sourceFile.Filename = strings.ReplaceAll(sourceFile.Filename, filepath.Ext(sourceFile.Filename), "") // strip ".yaml"
	sourceFile.Filename = strings.ReplaceAll(sourceFile.Filename, ".", "")                               // strip "." e.g. hidden files
	sourceFile.Filename += ".go"                                                                         // add correct file ext
	sourceFile.Filename = utils.ToFileName(sourceFile.Filename)                                          // kebab-case to snake_case

	return sourceFile
}

func expandResources(path string, resources []*Resource) ([]*Resource, error) {
	var expandedResources []*Resource

	for _, r := range resources {
		files, err := Glob(filepath.Join(path, r.FileName))
		if err != nil {
			return []*Resource{}, fmt.Errorf("failed to process glob pattern matching, %w", err)
		}

		for _, f := range files {
			rf, err := filepath.Rel(path, f)
			if err != nil {
				return []*Resource{}, fmt.Errorf("unable to determine relative file path, %w", err)
			}

			res := &Resource{FileName: f, relativeFileName: rf}
			expandedResources = append(expandedResources, res)
		}
	}

	return expandedResources, nil
}
