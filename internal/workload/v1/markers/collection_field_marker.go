// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package markers

import (
	"fmt"

	"github.com/nukleros/operator-builder/internal/markers/marker"
)

const (
	CollectionFieldMarkerPrefix = "+operator-builder:collection:field"
	CollectionFieldPrefix       = "collection"
	CollectionFieldSpecPrefix   = "collection.Spec"
)

// CollectionFieldMarker is an object which represents a marker that is associated with a
// collection field that exsists within a manifest.  A CollectionFieldMarker is discovered
// when a manifest is parsed and matches the constants defined by the collectionFieldMarker
// constant above.  It is represented identically to a FieldMarker with the caveat that it
// is discovered different via a different prefix.
type CollectionFieldMarker FieldMarker

//nolint:gocritic //needed to implement string interface
func (cfm CollectionFieldMarker) String() string {
	return fmt.Sprintf("CollectionFieldMarker{Name: %s Type: %v Description: %q Default: %v}",
		cfm.GetName(),
		cfm.Type,
		cfm.GetDescription(),
		cfm.Default,
	)
}

// defineCollectionFieldMarker will define a CollectionFieldMarker and add it a registry of markers.
func defineCollectionFieldMarker(registry *marker.Registry) error {
	collectionMarker, err := marker.Define(CollectionFieldMarkerPrefix, CollectionFieldMarker{})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	registry.Add(collectionMarker)

	return nil
}

// FieldMarkerProcessor interface methods.
func (cfm *CollectionFieldMarker) GetName() string {
	if cfm.Name == nil {
		return ""
	}

	return *cfm.Name
}

func (cfm *CollectionFieldMarker) GetDefault() interface{} {
	return cfm.Default
}

func (cfm *CollectionFieldMarker) GetDescription() string {
	if cfm.Description == nil {
		return ""
	}

	return *cfm.Description
}

func (cfm *CollectionFieldMarker) GetFieldType() FieldType {
	return cfm.Type
}

func (cfm *CollectionFieldMarker) GetReplaceText() string {
	if cfm.Replace == nil {
		return ""
	}

	return *cfm.Replace
}

func (cfm *CollectionFieldMarker) GetPrefix() string {
	return CollectionFieldPrefix
}

func (cfm *CollectionFieldMarker) GetSpecPrefix() string {
	return CollectionFieldSpecPrefix
}

func (cfm *CollectionFieldMarker) GetSourceCodeVariable() string {
	return cfm.sourceCodeVar
}

func (cfm *CollectionFieldMarker) GetOriginalValue() interface{} {
	return cfm.originalValue
}

func (cfm *CollectionFieldMarker) GetParent() string {
	if cfm.Parent == nil {
		return ""
	}

	return *cfm.Parent
}

func (cfm *CollectionFieldMarker) IsCollectionFieldMarker() bool {
	return true
}

func (cfm *CollectionFieldMarker) IsFieldMarker() bool {
	return false
}

func (cfm *CollectionFieldMarker) IsForCollection() bool {
	return cfm.forCollection
}

func (cfm *CollectionFieldMarker) IsArbitrary() bool {
	if cfm.Arbitrary == nil {
		return false
	}

	return *cfm.Arbitrary
}

func (cfm *CollectionFieldMarker) SetOriginalValue(value string) {
	if cfm.GetReplaceText() != "" {
		cfm.originalValue = cfm.GetReplaceText()

		return
	}

	cfm.originalValue = &value
}

func (cfm *CollectionFieldMarker) SetDescription(description string) {
	cfm.Description = &description
}

func (cfm *CollectionFieldMarker) SetForCollection(forCollection bool) {
	cfm.forCollection = forCollection
}
