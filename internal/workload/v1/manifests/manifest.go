// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package manifests

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nukleros/operator-builder/internal/utils"
	"github.com/nukleros/operator-builder/internal/workload/v1/markers"
)

var ErrProcessManifest = errors.New("error processing manifest file")

// Manifest represents a single input manifest for a given config.
type Manifest struct {
	Content                  []byte          `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
	Filename                 string          `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
	SourceFilename           string          `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
	ChildResources           []ChildResource `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
	PreferredSourceFileNames []string        `json:",omitempty" yaml:",omitempty" validate:"omitempty"`
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

			manifest := &Manifest{Filename: files[f], PreferredSourceFileNames: getFileNames(rf)}
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
			if content != "" {
				manifests = append(manifests, content)
				content = ""
			}
		} else {
			content = content + "\n" + line
		}
	}

	if content != "" {
		manifests = append(manifests, content)
	}

	return manifests
}

// LoadContent sets the Content field of the manifest in raw format as []byte.
func (manifest *Manifest) LoadContent(isCollection bool) error {
	manifestContent, err := os.ReadFile(manifest.Filename)
	if err != nil {
		return fmt.Errorf("%w; %s for manifest file %s", err, ErrProcessManifest.Error(), manifest.Filename)
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

// getFileNames returns all available file names.
func getFileNames(relativeFileName string) []string {
	// remove ./ and ../ from relative file name
	relativeFileName = strings.ReplaceAll(relativeFileName, "../", "")
	relativeFileName = strings.ReplaceAll(relativeFileName, "./", "")

	splitFilePath := strings.Split(filepath.Clean(relativeFileName), "/")

	fileNames := make([]string, len(splitFilePath))

	// prefer the flat file name, by itself, first working back to the full name
	var priority int

	var fileName string

	for i := len(splitFilePath); i > 0; i-- {
		// set the file path if unset, otherwise append
		if fileName == "" {
			fileName = splitFilePath[i-1]
		} else {
			fileName = fmt.Sprintf("%s/%s", splitFilePath[i-1], fileName)
		}

		fileName = strings.ReplaceAll(fileName, "/", "_")                   // get filename from path
		fileName = strings.ReplaceAll(fileName, filepath.Ext(fileName), "") // strip ".yaml"
		fileName = strings.ReplaceAll(fileName, ".", "")                    // strip "." e.g. hidden files
		fileName += ".go"                                                   // add correct file ext
		fileName = utils.ToFileName(fileName)                               // kebab-case to snake_case
		fileName = strings.ReplaceAll(fileName, "_internal_test.go", ".go") // ensure we do not end up with an internal test file
		fileName = strings.ReplaceAll(fileName, "_test.go", ".go")          // ensure we do not end up with a test file

		// strip any prefix that begins with _ or even multiple _s because go does not recognize these files
		for _, char := range fileName {
			if string(char) == "_" {
				fileName = strings.TrimPrefix(fileName, "_")
			} else {
				break
			}
		}

		// insert the file name at as specific priority and increase the cound
		fileNames[priority] = fileName
		priority++
	}

	return fileNames
}
