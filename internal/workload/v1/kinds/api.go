// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package kinds

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/nukleros/operator-builder/internal/utils"
	"github.com/nukleros/operator-builder/internal/workload/v1/markers"
)

var ErrOverwriteExistingValue = errors.New("an attempt to overwrite existing value was made")

type APIFields struct {
	Name         string
	StructName   string
	manifestName string
	Type         markers.FieldType
	Tags         string
	Comments     []string
	Markers      []string
	Children     []*APIFields
	Default      string
	Sample       string
}

func (api *APIFields) AddField(path string, fieldType markers.FieldType, comments []string, sample interface{}, hasDefault bool) error {
	obj := api

	parts := strings.Split(path, ".")

	last := parts[len(parts)-1]

	for _, part := range parts[:len(parts)-1] {
		foundMatch := false

		if obj.Children != nil {
			for i := range obj.Children {
				if obj.Children[i].manifestName == part {
					if obj.Children[i].Type != markers.FieldStruct {
						return fmt.Errorf("%w for api field %s", ErrOverwriteExistingValue, path)
					}

					foundMatch = true
					obj = obj.Children[i]

					break
				}
			}
		}

		if !foundMatch {
			child := obj.newChild(part, markers.FieldStruct, sample)

			child.Markers = append(child.Markers, "+kubebuilder:validation:Optional")

			child.generateStructName(path)

			obj.Children = append(obj.Children, child)
			obj = child
		}
	}

	newChild := obj.newChild(last, fieldType, sample)

	newChild.setCommentsAndDefault(comments, sample, hasDefault)

	for _, child := range obj.Children {
		if child.manifestName == last {
			if !child.isEqual(newChild) {
				return fmt.Errorf("%w for api field %s", ErrOverwriteExistingValue, path)
			}

			child.setCommentsAndDefault(comments, sample, hasDefault)

			return nil
		}
	}

	obj.Children = append(obj.Children, newChild)

	return nil
}

func (api *APIFields) GenerateAPISpec(kind string) string {
	var buf bytes.Buffer

	mustWrite(buf.WriteString(fmt.Sprintf(`
// %[1]sSpec defines the desired state of %[1]s.
type %[1]sSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

`, kind)))

	for _, child := range api.Children {
		child.generateAPISpecField(&buf, kind)
	}

	mustWrite(buf.WriteString("}\n\n"))

	for _, child := range api.Children {
		if child.Children != nil {
			child.generateAPIStruct(&buf, kind)
		}
	}

	return buf.String()
}

func (api *APIFields) GenerateSampleSpec(requiredOnly bool) string {
	var buf bytes.Buffer

	indent := 0

	api.generateSampleSpec(&buf, indent, requiredOnly)

	return buf.String()
}

func (api *APIFields) generateSampleSpec(b io.StringWriter, indent int, requiredOnly bool) {
	mustWrite(b.WriteString(fmt.Sprintf("%s%s\n", strings.Repeat("  ", indent), api.Sample)))

	for _, child := range api.Children {
		if child.needsGenerate(requiredOnly) {
			child.generateSampleSpec(b, indent+1, requiredOnly)
		}
	}
}

func (api *APIFields) needsGenerate(requiredOnly bool) bool {
	// if required fields only are not requested, return true immediately
	if !requiredOnly {
		return true
	}

	// traverse the api tree for any default value
	return api.hasRequiredField()
}

func (api *APIFields) hasRequiredField() bool {
	if len(api.Children) == 0 && api.Default == "" {
		return true
	}

	for _, child := range api.Children {
		if child.hasRequiredField() {
			return true
		}
	}

	return false
}

func (api *APIFields) generateAPISpecField(b io.StringWriter, kind string) {
	typeName := api.Type.String()
	if api.Type == markers.FieldStruct {
		typeName = kind + api.StructName
	}

	for _, m := range api.Markers {
		mustWrite(b.WriteString(fmt.Sprintf("// %s\n", m)))
	}

	for _, c := range api.Comments {
		mustWrite(b.WriteString(fmt.Sprintf("// %s\n", c)))
	}

	mustWrite(b.WriteString(fmt.Sprintf("%s %s %s\n\n", api.Name, typeName, api.Tags)))
}

func (api *APIFields) generateAPIStruct(b io.StringWriter, kind string) {
	if api.Type == markers.FieldStruct {
		mustWrite(b.WriteString(fmt.Sprintf("type %s %s{\n", kind+api.StructName, api.Type.String())))

		for _, child := range api.Children {
			child.generateAPISpecField(b, kind)
		}

		mustWrite(b.WriteString("}\n\n"))

		for _, child := range api.Children {
			child.generateAPIStruct(b, kind)
		}
	}
}

func (api *APIFields) generateStructName(path string) {
	var buf bytes.Buffer

	mustWrite(buf.WriteString("Spec"))

	for _, part := range strings.Split(path, ".") {
		mustWrite(buf.WriteString(utils.ToTitle(part)))

		if part == api.manifestName {
			break
		}
	}

	api.StructName = buf.String()
}

func (api *APIFields) isEqual(input *APIFields) bool {
	if api.Type != input.Type {
		return false
	}

	if api.Default == "" || api.Default == input.Default || input.Default == "" {
		if len(api.Comments) == 0 || len(input.Comments) == 0 {
			return true
		}

		if len(api.Comments) == len(input.Comments) {
			return reflect.DeepEqual(api.Comments, input.Comments)
		}
	}

	return false
}

// getSampleValue exists to solve the problem of the sample value being a brittle, generic interface
// which can change when we move from proper typed objects to pointers.  This function serves to
// solve both use cases.
func (api *APIFields) getSampleValue(sampleVal interface{}) string {
	switch t := sampleVal.(type) {
	case *string:
		if api.Type == markers.FieldString {
			return fmt.Sprintf(`%q`, *t)
		}

		return *t
	case *int:
		return fmt.Sprintf(`%v`, *t)
	case *bool:
		return fmt.Sprintf(`%v`, *t)
	case string:
		if api.Type == markers.FieldString {
			return fmt.Sprintf(`%q`, t)
		}

		return t
	default:
		return fmt.Sprintf(`%v`, t)
	}
}

func (api *APIFields) setSample(sampleVal interface{}) {
	switch api.Type {
	case markers.FieldStruct:
		api.Sample = fmt.Sprintf("%s:", api.manifestName)
	default:
		api.Sample = fmt.Sprintf("%s: %v", api.manifestName, api.getSampleValue(sampleVal))
	}
}

func (api *APIFields) setDefault(sampleVal interface{}) {
	api.Default = api.getSampleValue(sampleVal)
	api.appendMarkers(
		fmt.Sprintf("+kubebuilder:default=%s", api.Default),
		"+kubebuilder:validation:Optional",
		fmt.Sprintf("(Default: %s)", api.Default),
	)
	api.setSample(sampleVal)
}

func (api *APIFields) appendMarkers(apiMarkers ...string) {
	if len(api.Markers) == 0 {
		api.Markers = append(api.Markers, apiMarkers...)
	}
}

func (api *APIFields) setCommentsAndDefault(comments []string, sampleVal interface{}, hasDefault bool) {
	if hasDefault {
		api.setDefault(sampleVal)
	} else {
		api.appendMarkers("+kubebuilder:validation:Required")
	}

	if len(comments) > 0 {
		api.Comments = comments
	}
}

func (api *APIFields) newChild(name string, fieldType markers.FieldType, sample interface{}) *APIFields {
	child := &APIFields{
		Name:         utils.ToTitle(name),
		manifestName: name,
		Type:         fieldType,
		Tags:         fmt.Sprintf("`json:%q`", fmt.Sprintf("%s,%s", name, "omitempty")),
		Comments:     []string{},
		Markers:      []string{},
	}

	child.setSample(sample)

	return child
}

func mustWrite(n int, err error) {
	if err != nil {
		panic(err)
	}
}
