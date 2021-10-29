// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
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

func (fm FieldMarker) String() string {
	return fmt.Sprintf("FieldMarker{Name: %s Type: %v Description: %q Default: %v}",
		fm.Name,
		fm.Type,
		*fm.Description,
		fm.Default,
	)
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

func InitializeMarkerInspector(markerTypes ...MarkerType) (*inspect.Inspector, error) {
	registry := marker.NewRegistry()

	fieldMarker, err := marker.Define("+operator-builder:field", FieldMarker{})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	collectionMarker, err := marker.Define("+operator-builder:collection:field", CollectionFieldMarker{})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	for _, markerType := range markerTypes {
		switch markerType {
		case FieldMarkerType:
			registry.Add(fieldMarker)
		case CollectionMarkerType:
			registry.Add(collectionMarker)
		}
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

		key.HeadComment = ""
		key.FootComment = ""
		value.LineComment = ""

		switch t := r.Object.(type) {
		case FieldMarker:
			if t.Description != nil {
				*t.Description = strings.TrimPrefix(*t.Description, "\n")
				key.HeadComment = "# " + *t.Description + ", controlled by " + t.Name
			}

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
				value.Value = fmt.Sprintf("parent.Spec." + strings.Title(t.Name))
			}

			r.Object = t

		case CollectionFieldMarker:
			if t.Description != nil {
				*t.Description = strings.TrimPrefix(*t.Description, "\n")
				key.HeadComment = "# " + *t.Description + ", controlled by " + t.Name
			}

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
				value.Value = fmt.Sprintf("collection.Spec." + strings.Title(t.Name))
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
