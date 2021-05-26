package v1

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func (w *Workload) GetSpecFields(workloadPath string) (*[]APISpecField, error) {

	var specFields []APISpecField

	for _, manifestFile := range w.Spec.Resources {

		// capture entire resource manifest file content
		manifestContent, err := ioutil.ReadFile(filepath.Join(filepath.Dir(workloadPath), manifestFile))
		if err != nil {
			return nil, err
		}

		// extract all markers from yaml content
		markers, err := processManifest(string(manifestContent))
		if err != nil {
			return nil, err
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
