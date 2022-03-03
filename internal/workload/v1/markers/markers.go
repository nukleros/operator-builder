// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package markers

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/inspect"
	markerparser "github.com/vmware-tanzu-labs/operator-builder/internal/markers/marker"
)

// MarkerType defines the types of markers that are accepted by the parser.
type MarkerType int

const (
	FieldMarkerType MarkerType = iota
	CollectionMarkerType
	ResourceMarkerType
	UnknownMarkerType
)

// FieldMarkerProcessor is an interface that requires specific methods that are
// necessary for parsing a field marker or a collection field marker.
type FieldMarkerProcessor interface {
	GetName() string
	GetDefault() interface{}
	GetDescription() string
	GetFieldType() FieldType
	GetOriginalValue() interface{}
	GetReplaceText() string
	GetSpecPrefix() string
	GetSourceCodeVariable() string

	IsCollectionFieldMarker() bool
	IsFieldMarker() bool
	IsForCollection() bool

	SetDescription(string)
	SetOriginalValue(string)
	SetForCollection(bool)
}

// MarkerProcessor is a more generic interface that requires specific methods that are
// necessary for parsing any type of marker.
type MarkerProcessor interface {
	GetName() string
	GetSpecPrefix() string
}

// MarkerCollection is an object that stores a set of markers.
type MarkerCollection struct {
	FieldMarkers           []*FieldMarker
	CollectionFieldMarkers []*CollectionFieldMarker
}

// ContainsMarkerType will determine if a given marker type exists within
// a set of marker types.
func ContainsMarkerType(s []MarkerType, e MarkerType) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

// InspectForYAML will inspect yamlContent for a set of markers.  It will find
// all of the markers within the yamlContent and return the resultant lines and
// any associated errors.
func InspectForYAML(yamlContent []byte, markerTypes ...MarkerType) ([]*yaml.Node, []*inspect.YAMLResult, error) {
	insp, err := initializeMarkerInspector(markerTypes...)
	if err != nil {
		return nil, nil, fmt.Errorf("%w; error initializing markers %v", err, markerTypes)
	}

	nodes, results, err := insp.InspectYAML(yamlContent, transformYAML)
	if err != nil {
		return nil, nil, fmt.Errorf("%w; error inspecting YAML for markers %v", err, markerTypes)
	}

	return nodes, results, nil
}

// initializeMarkerInspector will create a new registry and initialize an inspector
// for specific marker types.
func initializeMarkerInspector(markerTypes ...MarkerType) (*inspect.Inspector, error) {
	registry := markerparser.NewRegistry()

	var err error

	for _, markerType := range markerTypes {
		switch markerType {
		case FieldMarkerType:
			err = defineFieldMarker(registry)
		case CollectionMarkerType:
			err = defineCollectionFieldMarker(registry)
		case ResourceMarkerType:
			err = defineResourceMarker(registry)
		}
	}

	if err != nil {
		return nil, err
	}

	return inspect.NewInspector(registry), nil
}

// transformYAML will transform a YAML result into the proper format for scaffolding
// resultant code and API definitions.
func transformYAML(results ...*inspect.YAMLResult) error {
	for _, result := range results {
		// convert to interface
		var marker FieldMarkerProcessor

		switch t := result.Object.(type) {
		case FieldMarker:
			t.sourceCodeVar = getSourceCodeVariable(&t)
			marker = &t
		case CollectionFieldMarker:
			t.sourceCodeVar = getSourceCodeVariable(&t)
			marker = &t
		default:
			continue
		}

		// get common variables and confirm that we are not working with a reserved marker
		if isReserved(marker.GetName()) {
			return fmt.Errorf("%s %w", marker.GetName(), ErrFieldMarkerReserved)
		}

		key, value := getKeyValue(result)

		setComments(marker, result, key, value)

		if err := setValue(marker, value); err != nil {
			return fmt.Errorf("%w; error setting value for marker %s", err, result.MarkerText)
		}

		result.Object = marker
	}

	return nil
}

// reservedMarkers represents a list of markers which cannot be used
// within a manifest.  They are reserved for internal purposes.  If any of the
// reservedMarkers are found, we will throw an error and notify the user.
func reservedMarkers() []string {
	return []string{
		"collection",
		"collection.name",
		"collection.namespace",
	}
}

// isReserved is a convenience method which returns whether or not a marker, given
// the fieldName as a string, is reserved for internal purposes.
func isReserved(fieldName string) bool {
	for _, reserved := range reservedMarkers() {
		if strings.Title(fieldName) == strings.Title(reserved) {
			return true
		}
	}

	return false
}

// getSourceCodeFieldVariable gets a full variable name for a marker as it is intended to be
// passed into the generate package to generate the source code.  This includes particular
// tags that are needed by the generator to properly identify when a variable starts and ends.
func getSourceCodeFieldVariable(marker FieldMarkerProcessor) string {
	return fmt.Sprintf("!!start %s !!end", marker.GetSourceCodeVariable())
}

// getSourceCodeVariable gets a full variable name for a marker as it is intended to be
// scaffolded in the source code.
func getSourceCodeVariable(marker MarkerProcessor) string {
	return fmt.Sprintf("%s.%s", marker.GetSpecPrefix(), strings.Title((marker.GetName())))
}

// getKeyValue gets the key and value from a YAML result.
func getKeyValue(result *inspect.YAMLResult) (key, value *yaml.Node) {
	if len(result.Nodes) > 1 {
		return result.Nodes[0], result.Nodes[1]
	}

	return result.Nodes[0], result.Nodes[0]
}

// setComments sets the comments for use by the resultant code.
func setComments(marker FieldMarkerProcessor, result *inspect.YAMLResult, key, value *yaml.Node) {
	// update the description to ensure new lines are commented
	if marker.GetDescription() != "" {
		marker.SetDescription(strings.TrimPrefix(marker.GetDescription(), "\n"))
		key.HeadComment = key.HeadComment + "\n# " + marker.GetDescription()
	}

	// set replace text to ensure that our markers are commented
	replaceText := strings.TrimSuffix(result.MarkerText, "\n")
	replaceText = strings.ReplaceAll(replaceText, "\n", "\n#")

	// set the append text to notify the user where a marker was originated from in their source code
	var appendText string
	switch t := marker.(type) {
	case *FieldMarker:
		appendText = "controlled by field: " + t.Name
	case *CollectionFieldMarker:
		appendText = "controlled by collection field: " + t.Name
	}

	// set the comments on the yaml nodes
	key.FootComment = ""
	key.HeadComment = strings.ReplaceAll(key.HeadComment, replaceText, appendText)
	value.LineComment = strings.ReplaceAll(value.LineComment, replaceText, appendText)
}

// setValue will set the value appropriately.  This is based on whether the marker has
// requested replacement text.
func setValue(marker FieldMarkerProcessor, value *yaml.Node) error {
	const varTag = "!!var"

	const strTag = "!!str"

	markerReplaceText := marker.GetReplaceText()

	marker.SetOriginalValue(value.Value)

	if markerReplaceText != "" {
		value.Tag = strTag

		re, err := regexp.Compile(markerReplaceText)
		if err != nil {
			return fmt.Errorf("unable to convert %s to regex, %w", markerReplaceText, err)
		}

		value.Value = re.ReplaceAllString(value.Value, getSourceCodeFieldVariable(marker))
	} else {
		value.Tag = varTag
		value.Value = marker.GetSourceCodeVariable()
	}

	return nil
}
