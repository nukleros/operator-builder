package v1

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	markerStr    = "+workload"
	docMarkerStr = "+workload-docs:"
)

// SupportedMarkerDataTypes returns the supported data types that can be used in
// workload markers.
func SupportedMarkerDataTypes() []string {
	return []string{"bool", "string", "int", "int32", "int64", "float32", "float64"}
}

func processMarkers(workloadPath string, resources []string, collection bool) (*[]APISpecField, error) {
	var specFields []APISpecField

	for _, manifestFile := range resources {

		// capture entire resource manifest file content
		manifestContent, err := ioutil.ReadFile(filepath.Join(filepath.Dir(workloadPath), manifestFile))
		if err != nil {
			return nil, err
		}

		// extract all workload markers from yaml content
		markers, err := processManifest(string(manifestContent))
		if err != nil {
			return nil, err
		}

	MARKERS:
		for _, m := range markers {
			// define all the cases in which we skip processing this marker
			switch {
			case collection && !m.Collection:
				continue
			case !collection && m.Collection:
				continue
			}

			for i, r := range specFields {
				if r.ManifestFieldName == m.FieldName {
					if len(m.DocumentationLines) > 0 {
						specFields[i].DocumentationLines = m.DocumentationLines
					}

					continue MARKERS
				}
			}

			specField := APISpecField{
				FieldName:          strings.Title(m.FieldName),
				ManifestFieldName:  m.FieldName,
				DataType:           m.DataType,
				DocumentationLines: m.DocumentationLines,
				APISpecContent: fmt.Sprintf(
					"%s %s `json:\"%s\"`",
					strings.Title(m.FieldName),
					m.DataType,
					m.FieldName,
				),
			}

			zv, err := zeroValue(m.DataType)
			if err != nil {
				return nil, err
			}
			specField.ZeroVal = zv

			if m.Default != "" {
				specField.DefaultVal = m.Default
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
	var startIndex int
	var hasDocs bool

	lines := strings.Split(manifest, "\n")
	for i, line := range lines {
		if containsDocumentMarker(line) {
			hasDocs = true
			startIndex = i
		}

		if containsMarker(line) {
			marker, err := processMarkerLine(line)
			if err != nil {
				return nil, err
			}

			if hasDocs {
				marker.DocumentationLines = processDocLines(lines, startIndex, i-1)

				// reset the hasDocs variable
				hasDocs = false
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

	var fieldPath string
	if strings.Contains(line, "collection=true") {
		fieldPath = fmt.Sprintf("collection.Spec.%s", strings.Title(fieldName))
	} else {
		fieldPath = fmt.Sprintf("parent.Spec.%s", strings.Title(fieldName))
	}

	if strings.Contains(code, ":") {
		keyValSplit := strings.Split(code, ":")
		key := keyValSplit[0]
		processed = fmt.Sprintf("%s: %s,", key, fieldPath)
	} else {
		processed = fmt.Sprintf("%s,", fieldPath)
	}

	return processed
}

func processDocLines(lines []string, start, end int) []string {
	docLines := []string{}

	for i := start; i <= end; i++ {
		line := strings.TrimLeft(lines[i], " ")
		if strings.HasPrefix(line, "#") {
			docLine := strings.TrimLeft(line, "#")
			docLine = strings.TrimLeft(docLine, " ")
			docLine = strings.TrimLeft(docLine, docMarkerStr)
			docLine = strings.TrimLeft(docLine, " ")

			docLines = append(docLines, docLine)
		}
	}

	return docLines
}

func processMarkerLine(line string) (Marker, error) {
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
			manifestVal += v
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
		} else if strings.Contains(element, "collection=") {
			collectionVal := strings.Split(element, "=")[1]
			switch collectionVal {
			case "true":
				marker.Collection = true
			case "false":
				marker.Collection = false
			default:
				msg := fmt.Sprintf("collection value %s found - must be either 'true' or 'false'", collectionVal)
				return marker, errors.New(msg)
			}
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
		return "", fmt.Errorf("unsupported data type in workload marker; supported data types: %v", SupportedMarkerDataTypes())
	}
}

func containsMarker(line string) bool {
	return strings.Contains(line, markerStr+":")
}

func containsDocumentMarker(line string) bool {
	return strings.Contains(line, docMarkerStr)
}
