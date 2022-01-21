// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/inspect"
	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/marker"
)

const (
	FieldMarkerType MarkerType = iota
	CollectionMarkerType
	ResourceMarkerType
)

const (
	collectionFieldMarker = "+operator-builder:collection:field"
	fieldMarker           = "+operator-builder:field"
	resourceMarker        = "+operator-builder:resource"

	collectionFieldSpecPrefix = "collection.Spec"
	fieldSpecPrefix           = "parent.Spec"

	resourceMarkerCollectionFieldName = "collectionField"
	resourceMarkerFieldName           = "field"
)

type MarkerType int

type FieldMarker struct {
	Name          string
	Type          FieldType
	Description   *string
	Default       interface{} `marker:",optional"`
	Replace       *string
	originalValue interface{}
}

type ResourceMarker struct {
	Field           *string
	CollectionField *string
	Value           interface{}
	Include         *bool

	sourceCodeVar   string
	sourceCodeValue string
	fieldMarker     interface{}
}

var (
	ErrMismatchedMarkerTypes            = errors.New("resource marker and field marker have mismatched types")
	ErrResourceMarkerUnknownValueType   = errors.New("resource marker 'value' is of unknown type")
	ErrResourceMarkerMissingFieldValue  = errors.New("resource marker missing 'collectionField', 'field' or 'value'")
	ErrResourceMarkerMissingInclude     = errors.New("resource marker missing 'include' value")
	ErrResourceMarkerMissingFieldMarker = errors.New("resource marker has no associated 'field' or 'collectionField' marker")
	ErrFieldMarkerInvalidType           = errors.New("field marker type is invalid")
)

func (fm FieldMarker) String() string {
	return fmt.Sprintf("FieldMarker{Name: %s Type: %v Description: %q Default: %v}",
		fm.Name,
		fm.Type,
		*fm.Description,
		fm.Default,
	)
}

func defineFieldMarker(registry *marker.Registry) error {
	fieldMarker, err := marker.Define(fieldMarker, FieldMarker{})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	registry.Add(fieldMarker)

	return nil
}

type CollectionFieldMarker FieldMarker

func (cfm CollectionFieldMarker) String() string {
	return fmt.Sprintf("CollectionFieldMarker{Name: %s Type: %v Description: %q Default: %v}",
		cfm.Name,
		cfm.Type,
		*cfm.Description,
		cfm.Default,
	)
}

func defineCollectionFieldMarker(registry *marker.Registry) error {
	collectionMarker, err := marker.Define(collectionFieldMarker, CollectionFieldMarker{})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	registry.Add(collectionMarker)

	return nil
}

//nolint:gocritic //needed to implement string interface
func (rm ResourceMarker) String() string {
	return fmt.Sprintf("ResourceMarker{Field: %s CollectionField: %s Value: %v Include: %v}",
		*rm.Field,
		*rm.CollectionField,
		rm.Value,
		*rm.Include,
	)
}

func defineResourceMarker(registry *marker.Registry) error {
	resourceMarker, err := marker.Define(resourceMarker, ResourceMarker{})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	registry.Add(resourceMarker)

	return nil
}

func InitializeMarkerInspector(markerTypes ...MarkerType) (*inspect.Inspector, error) {
	registry := marker.NewRegistry()

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

func TransformYAML(results ...*inspect.YAMLResult) error {
	const varTag = "!!var"

	const strTag = "!!str"

	var key *yaml.Node

	var value *yaml.Node

	for _, r := range results {
		if len(r.Nodes) > 1 {
			key = r.Nodes[0]
			value = r.Nodes[1]
		} else {
			key = r.Nodes[0]
			value = r.Nodes[0]
		}

		replaceText := strings.TrimSuffix(r.MarkerText, "\n")
		replaceText = strings.ReplaceAll(replaceText, "\n", "\n#")

		key.FootComment = ""

		switch t := r.Object.(type) {
		case FieldMarker:
			if t.Description != nil {
				*t.Description = strings.TrimPrefix(*t.Description, "\n")
				key.HeadComment = key.HeadComment + "\n# " + *t.Description
			}

			key.HeadComment = strings.ReplaceAll(key.HeadComment, replaceText, "controlled by field: "+t.Name)
			value.LineComment = strings.ReplaceAll(value.LineComment, replaceText, "controlled by field: "+t.Name)

			t.originalValue = value.Value

			if t.Replace != nil {
				value.Tag = strTag

				re, err := regexp.Compile(*t.Replace)
				if err != nil {
					return fmt.Errorf("unable to convert %s to regex, %w", *t.Replace, err)
				}

				value.Value = re.ReplaceAllString(value.Value, fmt.Sprintf("!!start parent.Spec.%s !!end", strings.Title((t.Name))))
			} else {
				value.Tag = varTag
				value.Value = getResourceDefinitionVar(strings.Title(t.Name), false)
			}

			r.Object = t

		case CollectionFieldMarker:
			if t.Description != nil {
				*t.Description = strings.TrimPrefix(*t.Description, "\n")
				key.HeadComment = "# " + *t.Description
			}

			key.HeadComment = strings.ReplaceAll(key.HeadComment, replaceText, "controlled by collection field: "+t.Name)
			value.LineComment = strings.ReplaceAll(value.LineComment, replaceText, "controlled by collection field: "+t.Name)

			t.originalValue = value.Value

			if t.Replace != nil {
				value.Tag = strTag

				re, err := regexp.Compile(*t.Replace)
				if err != nil {
					return fmt.Errorf("unable to convert %s to regex, %w", *t.Replace, err)
				}

				value.Value = re.ReplaceAllString(value.Value, fmt.Sprintf("!!start collection.Spec.%s !!end", strings.Title((t.Name))))
			} else {
				value.Tag = varTag
				value.Value = getResourceDefinitionVar(strings.Title(t.Name), true)
			}

			r.Object = t
		}
	}

	return nil
}

func containsMarkerType(s []MarkerType, e MarkerType) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

func inspectMarkersForYAML(yamlContent []byte, markerTypes ...MarkerType) ([]*yaml.Node, []*inspect.YAMLResult, error) {
	insp, err := InitializeMarkerInspector(markerTypes...)
	if err != nil {
		return nil, nil, fmt.Errorf("%w; error initializing markers %v", err, markerTypes)
	}

	nodes, results, err := insp.InspectYAML(yamlContent, TransformYAML)
	if err != nil {
		return nil, nil, fmt.Errorf("%w; error inspecting YAML for markers %v", err, markerTypes)
	}

	return nodes, results, nil
}

func getResourceDefinitionVar(path string, forCollectionMarker bool) string {
	// return the collectionFieldSpecPrefix only on a non-collection child resource
	// with a collection marker
	if forCollectionMarker {
		return fmt.Sprintf("%s.%s", collectionFieldSpecPrefix, strings.Title(path))
	}

	return fmt.Sprintf("%s.%s", fieldSpecPrefix, strings.Title(path))
}

func (rm *ResourceMarker) setSourceCodeVar() {
	if rm.Field != nil {
		rm.sourceCodeVar = getResourceDefinitionVar(*rm.Field, false)
	} else {
		rm.sourceCodeVar = getResourceDefinitionVar(*rm.CollectionField, true)
	}
}

func (rm *ResourceMarker) hasField() bool {
	var hasField, hasCollectionField bool

	if rm.Field != nil {
		if *rm.Field != "" {
			hasField = true
		}
	}

	if rm.CollectionField != nil {
		if *rm.CollectionField != "" {
			hasCollectionField = true
		}
	}

	return hasField || hasCollectionField
}

func (rm *ResourceMarker) hasValue() bool {
	return rm.Value != nil
}

func (rm *ResourceMarker) associateFieldMarker(spec *WorkloadSpec) {
	// return immediately if our entire workload spec has no field markers
	if len(spec.CollectionFieldMarkers) == 0 && len(spec.FieldMarkers) == 0 {
		return
	}

	// associate first relevant field marker with this marker
	for _, fm := range spec.FieldMarkers {
		if rm.Field != nil {
			if fm.Name == *rm.Field {
				rm.fieldMarker = fm

				return
			}
		}
	}

	// associate first relevant collection field marker with this marker
	for _, cm := range spec.CollectionFieldMarkers {
		if rm.CollectionField != nil {
			if cm.Name == *rm.CollectionField {
				rm.fieldMarker = cm

				return
			}
		}
	}
}

func (rm *ResourceMarker) validate() error {
	// check include field for a provided value
	// NOTE: this field is mandatory now, but could be optional later, so we return
	// an error here rather than using a pointer to a bool to control the mandate.
	if rm.Include == nil {
		return fmt.Errorf("%w for marker %s", ErrResourceMarkerMissingInclude, rm)
	}

	if rm.fieldMarker == nil {
		return fmt.Errorf("%w for marker %s", ErrResourceMarkerMissingFieldMarker, rm)
	}

	// ensure that both a field and value exist
	if !rm.hasField() || !rm.hasValue() {
		return fmt.Errorf("%w for marker %s", ErrResourceMarkerMissingFieldValue, rm)
	}

	return nil
}

func (rm *ResourceMarker) process() error {
	if err := rm.validate(); err != nil {
		return err
	}

	var fieldType string

	// determine if our associated field marker is a collection or regular field marker and
	// set appropriate variables
	switch marker := rm.fieldMarker.(type) {
	case *CollectionFieldMarker:
		fieldType = marker.Type.String()

		rm.setSourceCodeVar()
	case *FieldMarker:
		fieldType = marker.Type.String()

		rm.setSourceCodeVar()
	default:
		return fmt.Errorf("%w; type %T for marker %s", ErrFieldMarkerInvalidType, fieldMarker, rm)
	}

	// set the sourceCodeValue to check against
	switch value := rm.Value.(type) {
	case string:
		if fieldType != "string" {
			return fmt.Errorf("%w; expected: string, got: %s for marker %s", ErrMismatchedMarkerTypes, fieldType, rm)
		}

		rm.sourceCodeValue = fmt.Sprintf("%q", value)
	case int:
		if fieldType != "int" {
			return fmt.Errorf("%w; expected: int, got: %s for marker %s", ErrMismatchedMarkerTypes, fieldType, rm)
		}

		rm.sourceCodeValue = fmt.Sprintf("%v", value)
	case bool:
		if fieldType != "bool" {
			return fmt.Errorf("%w; expected: bool, got: %s for marker %s", ErrMismatchedMarkerTypes, fieldType, rm)
		}

		rm.sourceCodeValue = fmt.Sprintf("%v", value)
	default:
		return ErrResourceMarkerUnknownValueType
	}

	return nil
}
