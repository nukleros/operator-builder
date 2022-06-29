// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package manifests

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/markers"
)

var ErrProcessManifest = errors.New("error processing manifest file")

// Manifest represents a single input manifest for a given config.
type Manifest struct {
	Content        []byte          `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
	Filename       string          `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
	SourceFilename string          `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
	ChildResources []ChildResource `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
}

// Manifests represents a collection of manifests.
type Manifests []*Manifest

// ExpandManifests expands manifests from its globbed pattern and return the resultant manifest
// filenames from the glob.
func ExpandManifests(workloadPath string, manifestPaths []string) (*Manifests, error) {
	var manifests Manifests

	for i := range manifestPaths {
		files, err := utils.Glob(filepath.Join(workloadPath, manifestPaths[i]))
		if err != nil {
			return &Manifests{}, fmt.Errorf("failed to process glob pattern matching, %w", err)
		}

		for f := range files {
			rf, err := filepath.Rel(workloadPath, files[f])
			if err != nil {
				return &Manifests{}, fmt.Errorf("unable to determine relative file path, %w", err)
			}

			manifest := &Manifest{Filename: files[f], SourceFilename: getSourceFilename(rf)}
			manifests = append(manifests, manifest)
		}
	}

	return &manifests, nil
}

// ExtractManifests extracts the manifests as YAML strings from a manifest with
// existing manifest content.
func (manifest *Manifest) ExtractManifests() []string {
	var manifests []string

	lines := strings.Split(string(manifest.Content), "\n")

	var content string

	for _, line := range lines {
		if strings.TrimRight(line, " ") == "---" {
			if len(content) > 0 {
				manifests = append(manifests, content)
				content = ""
			}
		} else {
			content = content + "\n" + line
		}
	}

	if len(content) > 0 {
		manifests = append(manifests, content)
	}

	return manifests
}

// LoadContent sets the Content field of the manifest in raw format as []byte.
func (manifest *Manifest) LoadContent(isCollection bool) error {
	manifestContent, err := os.ReadFile(manifest.Filename)
	if err != nil {
		return fmt.Errorf("%w; %s for manifest file %s", err, ErrProcessManifest, manifest.Filename)
	}

	if isCollection {
		// replace all instances of collection markers and collection field markers with regular field markers
		// as a collection marker on a collection is simply a field marker to itself
		content := strings.ReplaceAll(string(manifestContent), markers.CollectionFieldMarkerPrefix, markers.FieldMarkerPrefix)
		content = strings.ReplaceAll(content, markers.ResourceMarkerCollectionFieldName, markers.ResourceMarkerFieldName)

		manifest.Content = []byte(content)
	} else {
		manifest.Content = manifestContent
	}

	return nil
}

// FromFiles returns new manifest objects given a set of file paths.
func FromFiles(manifestFiles []string) *Manifests {
	manifests := make(Manifests, len(manifestFiles))

	for i, manifestFile := range manifestFiles {
		manifest := &Manifest{
			Filename: manifestFile,
		}

		manifests[i] = manifest
	}

	return &manifests
}

// FuncNames returns the function names for a set of resources.  The function names are derived
// from the child resource unique names and refer to the functions that actually create the
// child resource objects in memory for the purposes of shipping to the Kubernetes API for
// deployment into the cluster.
func (manifests Manifests) FuncNames() (createFuncNames, initFuncNames []string) {
	foundCreateNames := make(map[string]int)
	foundInitNames := make(map[string]int)

	for m := range manifests {
		childResources := manifests[m].ChildResources

		for i := range childResources {
			// retrieve the create func names
			createFuncName := childResources[i].CreateFuncName()
			if foundCreateNames[createFuncName] > 0 {
				createFuncName = fmt.Sprintf("%s%v", createFuncName, foundCreateNames[createFuncName])
			}
			foundCreateNames[createFuncName]++
			createFuncNames = append(createFuncNames, createFuncName)

			// retrieve the init func names
			initFuncName := childResources[i].InitFuncName()
			if initFuncName == "" {
				continue
			}

			if foundInitNames[initFuncName] > 0 {
				initFuncName = fmt.Sprintf("%s%v", createFuncName, foundCreateNames[createFuncName])
			}
			foundInitNames[initFuncName]++
			initFuncNames = append(initFuncNames, initFuncName)
		}
	}

	return createFuncNames, initFuncNames
}

// getSourceFilename returns the unique file name for a source file.
func getSourceFilename(relativeFileName string) (name string) {
	name = filepath.Clean(relativeFileName)
	name = strings.ReplaceAll(name, "/", "_")               // get filename from path
	name = strings.ReplaceAll(name, filepath.Ext(name), "") // strip ".yaml"
	name = strings.ReplaceAll(name, ".", "")                // strip "." e.g. hidden files
	name += ".go"                                           // add correct file ext
	name = utils.ToFileName(name)                           // kebab-case to snake_case

	// strip any prefix that begins with _ or even multiple _s because go does not recognize these files
	for _, char := range name {
		if string(char) == "_" {
			name = strings.TrimPrefix(name, "_")
		} else {
			break
		}
	}

	return name
}
