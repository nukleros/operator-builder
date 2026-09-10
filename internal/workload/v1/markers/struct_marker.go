// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package markers

import (
	"errors"
	"fmt"

	"github.com/nukleros/markers/marker"
)

var (
	ErrStructMarkerInvalidName    = errors.New("struct marker name is invalid")
	ErrStructMarkerMissingComment = errors.New("struct marker missing description")
)

const (
	StructMarkerPrefix = "+operator-builder:struct"

	// StructRootName is the sentinel name that targets the root Spec object.
	StructRootName = "."
)

// StructMarker associates a description with a generated struct type.  Place it
// in a resource YAML file above the map node whose name matches the dot-separated
// path given in name=.  The description is emitted as a comment above both the
// field declaration in the parent struct and the type declaration itself, so it
// appears in CRD documentation and in kubectl explain output.
//
// name="." targets the root Spec struct.
//
// A struct marker is only valid when at least one field marker whose name begins
// with the same path prefix is also present; without a field marker the named
// struct will not exist in the generated API types.
type StructMarker struct {
	Name        *string
	Description *string
}

// defineStructMarker registers a StructMarker with the marker registry.
func defineStructMarker(registry *marker.Registry) error {
	structMarker, err := marker.Define(StructMarkerPrefix, StructMarker{})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	registry.Add(structMarker)

	return nil
}

func (sm *StructMarker) GetName() string {
	if sm.Name == nil {
		return ""
	}

	return *sm.Name
}

func (sm *StructMarker) GetDescription() string {
	if sm.Description == nil {
		return ""
	}

	return *sm.Description
}

// GetComments returns word-wrapped comment lines derived from the description.
func (sm *StructMarker) GetComments() []string {
	return commentsFromMarker(sm.GetDescription())
}
