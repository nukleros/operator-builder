package v1

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/vmware-tanzu-labs/object-code-generator-for-k8s/pkg/generate"
	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/inspect"
	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/marker"
	"github.com/vmware-tanzu-labs/operator-builder/internal/utils"
	"gopkg.in/yaml.v3"
)

// SupportedMarkerDataTypes returns the supported data types that can be used in
// workload markers.
func SupportedMarkerDataTypes() []string {
	return []string{"bool", "string", "int", "int32", "int64", "float32", "float64"}
}

func processMarkers(workloadPath string, resources []string, collection bool) (*SourceCodeTemplateData, error) {
	results := &SourceCodeTemplateData{
		SourceFile:    new([]SourceFile),
		RBACRule:      new([]RBACRule),
		OwnershipRule: new([]OwnershipRule),
	}

	specFields := make(map[string]*APISpecField)

	for _, manifestFile := range resources {
		// capture entire resource manifest file content
		manifestContent, err := ioutil.ReadFile(filepath.Join(filepath.Dir(workloadPath), manifestFile))
		if err != nil {
			return nil, err
		}

		insp, err := InitializeMarkerInspector()
		if err != nil {
			return nil, err
		}

		nodes, markerResults, err := insp.InspectYAML(manifestContent, TransformYAML)
		if err != nil {
			return nil, err
		}

		buf := bytes.Buffer{}

		for _, node := range nodes {
			m, err := yaml.Marshal(node)
			if err != nil {
				return nil, err
			}

			buf.WriteString("---\n")
			buf.Write(m)
		}

		manifestContent = buf.Bytes()

		for _, markerResult := range markerResults {
			switch r := markerResult.Object.(type) {
			case FieldMarker:
				if collection {
					continue
				}

				specField := &APISpecField{
					FieldName:         strings.ToTitle(r.Name),
					ManifestFieldName: r.Name,
					DataType:          r.Type.String(),
					APISpecContent: fmt.Sprintf(
						"%s %s `json:\"%s\"`",
						strings.Title(r.Name),
						r.Type,
						r.Name,
					),
				}

				if r.Description != nil {
					specField.DocumentationLines = strings.Split(*r.Description, "\n")
				}

				zv, err := zeroValue(r.Type.String())
				if err != nil {
					return nil, err
				}

				specField.ZeroVal = zv

				if r.Default != nil {
					if specField.DataType == "string" {
						specField.DefaultVal = fmt.Sprintf("%q", r.Default)
						specField.SampleField = fmt.Sprintf("%s: %q", r.Name, r.Default)
					} else {
						specField.DefaultVal = fmt.Sprintf("%v", r.Default)
						specField.SampleField = fmt.Sprintf("%s: %v", r.Name, r.Default)
					}
				} else {
					if specField.DataType == "string" {
						specField.SampleField = fmt.Sprintf("%s: %q", r.Name, r.originalValue)
					} else {
						specField.SampleField = fmt.Sprintf("%s: %v", r.Name, r.originalValue)
					}
				}

				specFields[r.Name] = specField
			case CollectionFieldMarker:
				if !collection {
					continue
				}

				specField := &APISpecField{
					FieldName:         strings.ToTitle(r.Name),
					ManifestFieldName: r.Name,
					DataType:          r.Type.String(),
					APISpecContent: fmt.Sprintf(
						"%s %s `json:\"%s\"`",
						strings.Title(r.Name),
						r.Type,
						r.Name,
					),
				}

				if r.Description != nil {
					specField.DocumentationLines = strings.Split(*r.Description, "\n")
				}

				zv, err := zeroValue(r.Type.String())
				if err != nil {
					return nil, err
				}

				specField.ZeroVal = zv

				if r.Default != nil {
					if specField.DataType == "string" {
						specField.DefaultVal = fmt.Sprintf("%q", r.Default)
						specField.SampleField = fmt.Sprintf("%s: %q", r.Name, r.Default)
					} else {
						specField.DefaultVal = fmt.Sprintf("%v", r.Default)
						specField.SampleField = fmt.Sprintf("%s: %v", r.Name, r.Default)
					}
				} else {
					if specField.DataType == "string" {
						specField.SampleField = fmt.Sprintf("%s: %q", r.Name, r.originalValue)
					} else {
						specField.SampleField = fmt.Sprintf("%s: %v", r.Name, r.originalValue)
					}
				}

				specFields[r.Name] = specField
			default:
				continue
			}
		}

		if collection {
			continue
		}

		// determine sourceFile filename
		var sourceFile SourceFile
		sourceFile.Filename = filepath.Base(manifestFile)                // get filename from path
		sourceFile.Filename = strings.Split(sourceFile.Filename, ".")[0] // strip ".yaml"
		sourceFile.Filename += ".go"                                     // add correct file ext
		sourceFile.Filename = utils.ToFileName(sourceFile.Filename)      // kebab-case to snake_case

		var childResources []ChildResource

		manifests := extractManifests(manifestContent)

		for _, manifest := range manifests {
			// unmarshal yaml to get attributes
			var manifestMetadata struct {
				Kind       string
				APIVersion string
				Metadata   struct {
					Name string
				}
			}

			var rawContent interface{}

			err = yaml.Unmarshal([]byte(manifest), &manifestMetadata)
			if err != nil {
				return nil, err
			}

			err = yaml.Unmarshal([]byte(manifest), &rawContent)
			if err != nil {
				return nil, err
			}

			// generate a unique name for the resource using the kind and name
			resourceUniqueName := strings.Replace(strings.Title(manifestMetadata.Metadata.Name), "-", "", -1)
			resourceUniqueName = strings.Replace(resourceUniqueName, ".", "", -1)
			resourceUniqueName = strings.Replace(resourceUniqueName, ":", "", -1)
			resourceUniqueName = fmt.Sprintf("%s%s", manifestMetadata.Kind, resourceUniqueName)

			// determine resource group and version
			resourceVersion, resourceGroup := versionGroupFromAPIVersion(manifestMetadata.APIVersion)

			// determine group and resource for RBAC rule generation
			rbacRulesForManifest(manifestMetadata.Kind, resourceGroup, rawContent, results.RBACRule)

			// determine group and kind for ownership rule generation
			newOwnershipRule := OwnershipRule{
				Version: manifestMetadata.APIVersion,
				Kind:    manifestMetadata.Kind,
				CoreAPI: isCoreAPI(resourceGroup),
			}

			ownershipExists := versionKindRecorded(results.OwnershipRule, &newOwnershipRule)
			if !ownershipExists {
				*results.OwnershipRule = append(*results.OwnershipRule, newOwnershipRule)
			}

			resource := ChildResource{
				Name:       manifestMetadata.Metadata.Name,
				UniqueName: resourceUniqueName,
				Group:      resourceGroup,
				Version:    resourceVersion,
				Kind:       manifestMetadata.Kind,
			}

			// generate the object source code
			resourceDefinition, err := generate.Generate([]byte(manifest), "resourceObj")
			if err != nil {
				return nil, err
			}

			// add the source code to the resource
			resource.SourceCode = resourceDefinition
			resource.StaticContent = manifest

			childResources = append(childResources, resource)
		}

		sourceFile.Children = childResources
		*results.SourceFile = append(*results.SourceFile, sourceFile)
	}

	for _, v := range specFields {
		results.SpecField = append(results.SpecField, v)
	}

	return results, nil
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

func InitializeMarkerInspector() (*inspect.Inspector, error) {
	registry := marker.NewRegistry()

	fieldMarker, err := marker.Define("+operator-builder:field", FieldMarker{})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	collectionMarker, err := marker.Define("+operator-builder:collection:field", CollectionFieldMarker{})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	registry.Add(fieldMarker)
	registry.Add(collectionMarker)

	return inspect.NewInspector(registry), nil
}

func TransformYAML(results ...*inspect.YAMLResult) error {
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

			value.Tag = "!!var"
			value.Value = fmt.Sprintf("parent.Spec." + strings.Title(t.Name))

			r.Object = t

		case CollectionFieldMarker:
			if t.Description != nil {
				*t.Description = strings.TrimPrefix(*t.Description, "\n")
				key.HeadComment = "# " + *t.Description + ", controlled by " + t.Name
			}

			t.originalValue = value.Value

			value.Tag = "!!var"
			value.Value = fmt.Sprintf("collection.Spec." + strings.Title(t.Name))

			r.Object = t
		}
	}

	return nil
}

type FieldType int

const (
	FieldUnknownType FieldType = iota
	FieldString
	FieldInt
	FieldBool
)

func (f *FieldType) UnmarshalMarkerArg(in string) error {
	types := map[string]FieldType{
		"":       FieldUnknownType,
		"string": FieldString,
		"int":    FieldInt,
		"bool":   FieldBool,
	}

	if t, ok := types[in]; ok {
		if t == FieldUnknownType {
			return fmt.Errorf("unable to parse %s into FieldType", in)
		}

		*f = t

		return nil
	}

	return fmt.Errorf("unable to parse %s into FieldType", in)
}

func (f FieldType) String() string {
	types := map[FieldType]string{
		FieldUnknownType: "",
		FieldString:      "string",
		FieldInt:         "int",
		FieldBool:        "bool",
	}

	return types[f]
}

type FieldMarker struct {
	Name          string
	Type          FieldType
	Description   *string
	Default       interface{} `marker:",optional"`
	originalValue interface{}
}

type CollectionFieldMarker FieldMarker

func (fm FieldMarker) String() string {
	return fmt.Sprintf("FieldMarker{Name: %s Type: %v Description: %q Default: %v}",
		fm.Name,
		fm.Type,
		*fm.Description,
		fm.Default,
	)
}
