// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/markers"
)

var (
	ErrChildResourceResourceMarkerInspect = errors.New("error inspecting resource markers for child resource")
	ErrChildResourceResourceMarkerProcess = errors.New("error processing resource markers for child resource")
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
	IncludeCode   string
}

// Resource represents a single input manifest for a given config.
type Resource struct {
	relativeFileName string
	FileName         string `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
	Content          []byte `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
}

func (r *Resource) UnmarshalYAML(node *yaml.Node) error {
	r.FileName = node.Value

	return nil
}

func (r *Resource) loadContent(isCollection bool) error {
	manifestContent, err := os.ReadFile(r.FileName)
	if err != nil {
		return formatProcessError(r.FileName, err)
	}

	if isCollection {
		// replace all instances of collection markers and collection field markers with regular field markers
		// as a collection marker on a collection is simply a field marker to itself
		content := strings.ReplaceAll(string(manifestContent), markers.CollectionFieldMarkerPrefix, markers.FieldMarkerPrefix)
		content = strings.ReplaceAll(content, markers.ResourceMarkerCollectionFieldName, markers.ResourceMarkerFieldName)

		r.Content = []byte(content)
	} else {
		r.Content = manifestContent
	}

	return nil
}

func (r *Resource) extractManifests() []string {
	var manifests []string

	lines := strings.Split(string(r.Content), "\n")

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

func ResourcesFromFiles(resourceFiles []string) []*Resource {
	return getResourcesFromFiles(resourceFiles)
}

func getResourcesFromFiles(resourceFiles []string) []*Resource {
	resources := make([]*Resource, len(resourceFiles))

	for i, resourceFile := range resourceFiles {
		resource := &Resource{
			FileName: resourceFile,
		}

		resources[i] = resource
	}

	return resources
}

func getFuncNames(sourceFiles []SourceFile) (createFuncNames, initFuncNames []string) {
	for _, sourceFile := range sourceFiles {
		for i := range sourceFile.Children {
			funcName := fmt.Sprintf("Create%s", sourceFile.Children[i].UniqueName)

			if strings.EqualFold(sourceFile.Children[i].Kind, "customresourcedefinition") {
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

	// strip any prefix that begins with _ or even multiple _s because go does not recognize these files
	for _, char := range sourceFile.Filename {
		if string(char) == "_" {
			sourceFile.Filename = strings.TrimPrefix(sourceFile.Filename, "_")
		} else {
			break
		}
	}

	return sourceFile
}

func expandResources(path string, resources []*Resource) ([]*Resource, error) {
	var expandedResources []*Resource

	for _, r := range resources {
		files, err := utils.Glob(filepath.Join(path, r.FileName))
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

func (cr *ChildResource) processResourceMarkers(markerCollection *markers.MarkerCollection) error {
	// obtain the marker results from the child resource input yaml
	_, markerResults, err := markers.InspectForYAML([]byte(cr.StaticContent), markers.ResourceMarkerType)
	if err != nil {
		return fmt.Errorf("%w; %s", err, ErrChildResourceResourceMarkerInspect)
	}

	// ensure we have the expected number of resource markers
	//   - 0: return immediately as resource markers are not required
	//   - 1: continue processing normally
	//   - 2: return an error notifying the user that we only expect 1
	//        resource marker
	if len(markerResults) == 0 {
		return nil
	}

	//nolint: godox // depends on https://github.com/vmware-tanzu-labs/operator-builder/issues/271
	// TODO: we need to ensure only one marker is found and return an error if we find more than one.
	// this becomes difficult as the results are returned as yaml nodes.  for now, we just focus on the
	// first result and all others are ignored but we should notify the user.
	result := markerResults[0]

	// process the marker
	marker, ok := result.Object.(markers.ResourceMarker)
	if !ok {
		return ErrChildResourceResourceMarkerProcess
	}

	if err := marker.Process(markerCollection); err != nil {
		return fmt.Errorf("%w; %s", err, ErrChildResourceResourceMarkerProcess)
	}

	if marker.GetIncludeCode() != "" {
		cr.IncludeCode = marker.GetIncludeCode()
	}

	return nil
}
