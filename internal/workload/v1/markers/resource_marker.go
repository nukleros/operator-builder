// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package markers

import (
	"errors"
	"fmt"

	"github.com/nukleros/markers/marker"
)

var (
	ErrResourceMarkerInvalid           = errors.New("resource marker is invalid")
	ErrResourceMarkerCount             = errors.New("expected only 1 resource marker")
	ErrResourceMarkerAssociation       = errors.New("unable to associate resource marker with 'field' or 'collectionField' marker")
	ErrResourceMarkerTypeMismatch      = errors.New("resource marker and field marker have mismatched types")
	ErrResourceMarkerInvalidType       = errors.New("expected resource marker type")
	ErrResourceMarkerUnknownValueType  = errors.New("resource marker 'value' is of unknown type")
	ErrResourceMarkerMissingFieldValue = errors.New("resource marker missing 'collectionField', 'field' or 'value'")
	ErrResourceMarkerMissingInclude    = errors.New("resource marker missing 'include' value")
)

const (
	ResourceMarkerPrefix              = "+operator-builder:resource"
	ResourceMarkerCollectionFieldName = "collectionField"
	ResourceMarkerFieldName           = "field"
)

// If we have a valid resource marker,  we will either include or exclude the
// related object based on the inputs on the resource marker itself.  These are
// the resultant code snippets based on that logic.
const (
	includeCode = `if %s != %s {
		return []client.Object{}, nil
	}`

	excludeCode = `if %s == %s {
		return []client.Object{}, nil
	}`
)

// ResourceMarker is an object which represents a marker for an entire resource.  It
// allows actions against a resource.  A ResourceMarker is discovered when a manifest
// is parsed and matches the constants defined by the collectionFieldMarker
// constant above.
type ResourceMarker struct {
	// inputs from the marker itself
	Field           *string
	CollectionField *string
	Value           interface{}
	Include         *bool

	// other field which we use to pass information
	includeCode string
	fieldMarker FieldMarkerProcessor
}

// String simply returns the marker as it should be printed in string format.
func (rm ResourceMarker) String() string {
	var fieldString, collectionFieldString string

	var includeBool bool

	// set the values if they have been provided otherwise take the zero values
	if rm.Field != nil {
		fieldString = *rm.Field
	}

	if rm.CollectionField != nil {
		collectionFieldString = *rm.CollectionField
	}

	if rm.Include != nil {
		includeBool = *rm.Include
	}

	return fmt.Sprintf("ResourceMarker{Field: %s CollectionField: %s Value: %v Include: %v}",
		fieldString,
		collectionFieldString,
		rm.Value,
		includeBool,
	)
}

// defineResourceMarker will define a ResourceMarker and add it a registry of markers.
func defineResourceMarker(registry *marker.Registry) error {
	resourceMarker, err := marker.Define(ResourceMarkerPrefix, ResourceMarker{})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	registry.Add(resourceMarker)

	return nil
}

// GetIncludeCode is a convenience function to return the include code of the resource marker.
func (rm *ResourceMarker) GetIncludeCode() string {
	return rm.includeCode
}

// GetName is a convenience function to return the name of the associated field marker.
func (rm *ResourceMarker) GetName() string {
	if rm.GetField() != "" {
		return rm.GetField()
	}

	return rm.GetCollectionField()
}

// GetCollectionField is a convenience function to return the collection field as a string.
func (rm *ResourceMarker) GetCollectionField() string {
	if rm.CollectionField == nil {
		return ""
	}

	return *rm.CollectionField
}

// GetField is a convenience function to return the field as a string.
func (rm *ResourceMarker) GetField() string {
	if rm.Field == nil {
		return ""
	}

	return *rm.Field
}

// GetPrefix is a convenience function to return the prefix of a requested
// variable for a resource marker.
func (rm *ResourceMarker) GetPrefix() string {
	if rm.Field != nil {
		return FieldPrefix
	}

	return CollectionFieldPrefix
}

// GetSpecPrefix is a convenience function to return the spec prefix of a requested
// variable for a resource marker.
func (rm *ResourceMarker) GetSpecPrefix() string {
	if rm.Field != nil {
		return FieldSpecPrefix
	}

	return CollectionFieldSpecPrefix
}

// GetParent is a convenience function to satisfy the MarkerProcessor interface.  It will
// always return an empty string for a resource marker because we do not care about parent
// fields.
func (rm *ResourceMarker) GetParent() string {
	return ""
}

// Process will process a resource marker from a collection of collection field markers
// and field markers, associate them together and set the appropriate fields.
func (rm *ResourceMarker) Process(markers *MarkerCollection) error {
	// ensure we have a valid field marker before continuing to process
	if err := rm.validate(); err != nil {
		return fmt.Errorf("%w; %s", err, ErrResourceMarkerInvalid.Error())
	}

	// associate field markers from a collection of markers to this resource marker
	if fieldMarker := rm.getFieldMarker(markers); fieldMarker != nil {
		rm.fieldMarker = fieldMarker
	} else {
		return fmt.Errorf("%w; %s", ErrResourceMarkerAssociation, rm)
	}

	// set the source code value and return
	if err := rm.setSourceCode(); err != nil {
		return fmt.Errorf("%w; error setting source code value for resource marker: %v", err, rm)
	}

	return nil
}

// validate checks for a valid resource marker and returns an error if the
// resource marker is invalid.
func (rm *ResourceMarker) validate() error {
	// check include field for a provided value
	// NOTE: this field is mandatory now, but could be optional later, so we return
	// an error here rather than using a pointer to a bool to control the mandate.
	if rm.Include == nil {
		return fmt.Errorf("%w for marker %s", ErrResourceMarkerMissingInclude, rm)
	}

	// ensure that both a field and value exist
	if !rm.hasField() || !rm.hasValue() {
		return fmt.Errorf("%w for marker %s", ErrResourceMarkerMissingFieldValue, rm)
	}

	return nil
}

// hasField determines whether or not a parsed resource marker has either a field
// or a collection field.  One or the other is needed for processing a resource
// marker.
func (rm *ResourceMarker) hasField() bool {
	return rm.GetName() != ""
}

// hasValue determines whether or not a parsed resource marker has a value
// to check against.
func (rm *ResourceMarker) hasValue() bool {
	return rm.Value != nil
}

// isAssociated returns whether a given marker is associated with a given resource
// marker.
func (rm *ResourceMarker) isAssociated(fromMarker FieldMarkerProcessor) bool {
	var field string

	switch {
	case fromMarker.IsCollectionFieldMarker():
		field = rm.GetCollectionField()
	case fromMarker.IsFieldMarker() && fromMarker.IsForCollection():
		if rm.GetCollectionField() != "" {
			field = rm.GetCollectionField()
		} else {
			field = rm.GetField()
		}
	default:
		field = rm.GetField()
	}

	return field == fromMarker.GetName()
}

// getFieldMarker will return the associated collection marker or field marker
// with a particular resource marker given a collection of markers.
func (rm *ResourceMarker) getFieldMarker(markers *MarkerCollection) FieldMarkerProcessor {
	// return immediately if the marker collection we are trying to associate is empty
	if len(markers.CollectionFieldMarkers) == 0 && len(markers.FieldMarkers) == 0 {
		return nil
	}

	// attempt to associate the field marker first
	for _, fm := range markers.FieldMarkers {
		if associatedWith := rm.isAssociated(fm); associatedWith {
			return fm
		}
	}

	// attempt to associate a field marker from a collection field next
	for _, cfm := range markers.CollectionFieldMarkers {
		if associatedWith := rm.isAssociated(cfm); associatedWith {
			return cfm
		}
	}

	return nil
}

// setSourceCode sets the source code to use as generated by the resource marker.
func (rm *ResourceMarker) setSourceCode() error {
	var sourceCodeVar, sourceCodeValue string

	// get the source code variable
	sourceCodeVar, err := getSourceCodeVariable(rm)
	if err != nil {
		return fmt.Errorf("%w; error retrieving source code variable for resource marker: %s", err, rm)
	}

	// set the source code value and ensure the types match
	switch value := rm.Value.(type) {
	case string, int, bool:
		fieldMarkerType := rm.fieldMarker.GetFieldType().String()
		resourceMarkerType := fmt.Sprintf("%T", value)

		if fieldMarkerType != resourceMarkerType {
			return fmt.Errorf("%w; expected: %s, got: %s for marker %s",
				ErrResourceMarkerTypeMismatch,
				resourceMarkerType,
				fieldMarkerType,
				rm,
			)
		}

		if fieldMarkerType == "string" {
			sourceCodeValue = fmt.Sprintf("%q", value)
		} else {
			sourceCodeValue = fmt.Sprintf("%v", value)
		}
	default:
		return ErrResourceMarkerUnknownValueType
	}

	// set the include code for this marker
	if *rm.Include {
		rm.includeCode = fmt.Sprintf(includeCode, sourceCodeVar, sourceCodeValue)
	} else {
		rm.includeCode = fmt.Sprintf(excludeCode, sourceCodeVar, sourceCodeValue)
	}

	return nil
}
