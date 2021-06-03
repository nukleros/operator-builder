package v1

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const markerStr = "+workload"

// SupportedMarkerDataTypes returns the supported data types that can be used in
// workload markers
func SupportedMarkerDataTypes() []string {
	return []string{"bool", "string", "int", "int32", "int64", "float32", "float64"}
}

func processMarkers(workloadPath string, resources []string) (*[]APISpecField, error) {

	var specFields []APISpecField

	for _, manifestFile := range resources {

		// capture entire resource manifest file content
		manifestContent, err := ioutil.ReadFile(filepath.Join(filepath.Dir(workloadPath), manifestFile))
		if err != nil {
			return &[]APISpecField{}, err
		}

		// extract all markers from yaml content
		markers, err := processManifest(string(manifestContent))
		if err != nil {
			return &[]APISpecField{}, err
		}

	MARKERS:
		for _, m := range markers {
			for _, r := range specFields {
				if r.ManifestFieldName == m.FieldName {
					continue MARKERS
				}
			}

			var specField APISpecField
			specField.FieldName = strings.Title(m.FieldName)
			specField.ManifestFieldName = m.FieldName
			specField.DataType = m.DataType
			if m.Default != "" {
				specField.DefaultVal = m.Default
			}
			zv, err := zeroValue(m.DataType)
			if err != nil {
				return nil, err
			}
			specField.ZeroVal = zv
			specField.ApiSpecContent = fmt.Sprintf(
				"%s %s `json:\"%s\"`",
				strings.Title(m.FieldName),
				m.DataType,
				m.FieldName,
			)
			if m.Default != "" {
				specField.SampleField = fmt.Sprintf("%s: %s", m.FieldName, m.Default)
			} else {
				specField.SampleField = fmt.Sprintf("%s: %s", m.FieldName, m.Value)
			}
			specFields = append(specFields, specField)
		}
	}

	return &specFields, nil
}

func processManifest(manifest string) ([]Marker, error) {

	var markers []Marker
	lines := strings.Split(string(manifest), "\n")
	for _, line := range lines {
		if containsMarker(line) {
			marker, err := processMarker(line)
			if err != nil {
				return nil, err
			}
			markers = append(markers, marker)
		}
	}

	return markers, nil
}

func processMarkedComments(line string) (processed string) {

	codeCommentSplit := strings.Split(line, "//")
	code := codeCommentSplit[0]
	comment := codeCommentSplit[1]
	commentSplit := strings.Split(comment, ":")
	fieldName := commentSplit[1]
	fieldPath := fmt.Sprintf("parent.Spec.%s", strings.Title(fieldName))

	if strings.Contains(code, ":") {
		keyValSplit := strings.Split(code, ":")
		key := keyValSplit[0]
		processed = fmt.Sprintf("%s: %s,", key, fieldPath)
	} else {
		processed = fmt.Sprintf("%s,", fieldPath)
	}

	return processed
}

func processMarker(line string) (Marker, error) {

	var marker Marker

	// count leading spaces
	var spaces int
	for _, char := range line {
		if char == ' ' {
			spaces++
		} else {
			break
		}
	}
	marker.LeadingSpaces = spaces

	commentedLine := strings.Split(line, "#")
	if len(commentedLine) != 2 {
		return marker, errors.New("+workload markers in static manifests must be commented out with a single '#' comment symbol")
	}

	// extract key and value from manifest
	keyVal := commentedLine[0]
	keyValSlice := strings.Split(keyVal, ":")
	manifestKey := strings.Replace(keyValSlice[0], "- ", "", 1)
	var manifestVal string
	var valElements int
	for _, v := range keyValSlice[1:] {
		valElements++
		if valElements > 1 {
			manifestVal = manifestVal + ":" + v
		} else {
			manifestVal = manifestVal + v
		}
	}
	marker.Key = strings.TrimSpace(manifestKey)
	marker.Value = strings.TrimSpace(manifestVal)

	// parse marker elements
	// marker elements are colon-separated
	markerLine := commentedLine[1]
	markerElements := strings.Split(markerLine, ":")
	for i, element := range markerElements {
		if strings.HasSuffix(element, `\`) {
			// backslash used to escape colons that are *not* delimeters
			// combine this element, without the backslash, with the next element
			element = strings.Split(element, `\`)[0] + ":" + markerElements[i+1]
		}

		if strings.Contains(element, markerStr) {
			continue
		} else if strings.HasSuffix(markerElements[i-1], `\`) {
			// this element has already been combined with the last one
			continue
		} else if strings.Contains(element, "type=") {
			marker.DataType = strings.Split(element, "=")[1]
		} else if strings.Contains(element, "default=") {
			marker.Default = strings.Split(element, "=")[1]
		} else {
			marker.FieldName = element
		}
	}

	return marker, nil

}

// zeroValue returns the zero value for the data type as a string.
// It is returned as a string to be used in a template for Go source code.
func zeroValue(val interface{}) (string, error) {
	switch val {
	case "bool":
		return "false", nil
	case "string":
		return "\"\"", nil
	case "int", "int32", "int64", "float32", "float64":
		return "0", nil
	default:
		return "", fmt.Errorf("unsupported data type in workload marker.  Support data types: %v", SupportedMarkerDataTypes())
	}
}

func containsMarker(line string) bool {
	return strings.Contains(line, markerStr)
}

func containsDefault(line string) bool {
	return strings.Contains(line, "default=")
}
