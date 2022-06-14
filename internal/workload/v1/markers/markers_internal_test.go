// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package markers

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/inspect"
	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/parser"
)

func TestContainsMarkerType(t *testing.T) {
	t.Parallel()

	knownMarkerTypes := []MarkerType{
		FieldMarkerType,
		CollectionMarkerType,
		ResourceMarkerType,
	}

	type args struct {
		s []MarkerType
		e MarkerType
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ensure missing marker type returns false",
			args: args{
				s: knownMarkerTypes,
				e: UnknownMarkerType,
			},
			want: false,
		},
		{
			name: "ensure non-missing marker type returns true",
			args: args{
				s: knownMarkerTypes,
				e: FieldMarkerType,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ContainsMarkerType(tt.args.s, tt.args.e); got != tt.want {
				t.Errorf("ContainsMarkerType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceMarker_hasField(t *testing.T) {
	t.Parallel()

	testPath := "test.has.field"
	testEmpty := ""

	type fields struct {
		Field           *string
		CollectionField *string
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "resource marker with field returns true",
			fields: fields{
				Field: &testPath,
			},
			want: true,
		},
		{
			name: "resource marker with collection field returns true",
			fields: fields{
				CollectionField: &testPath,
			},
			want: true,
		},
		{
			name: "resource marker with empty field and collection field returns false",
			fields: fields{
				Field:           &testEmpty,
				CollectionField: &testEmpty,
			},
			want: false,
		},
		{
			name:   "resource marker without field or collection field returns false",
			fields: fields{},
			want:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := &ResourceMarker{
				Field:           tt.fields.Field,
				CollectionField: tt.fields.CollectionField,
			}
			if got := rm.hasField(); got != tt.want {
				t.Errorf("ResourceMarker.hasField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceMarker_hasValue(t *testing.T) {
	t.Parallel()

	testValue := "test.has.value"

	type fields struct {
		Value interface{}
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "resource marker with nil value returns false",
			fields: fields{
				Value: nil,
			},
			want: false,
		},
		{
			name: "resource marker with value returns true",
			fields: fields{
				Value: &testValue,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rm := &ResourceMarker{
				Value: tt.fields.Value,
			}
			if got := rm.hasValue(); got != tt.want {
				t.Errorf("ResourceMarker.hasValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isReserved(t *testing.T) {
	t.Parallel()

	type args struct {
		fieldName string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ensure reserved field returns true",
			args: args{
				fieldName: "collection.name",
			},
			want: true,
		},
		{
			name: "ensure reserved field as a title returns true",
			args: args{
				fieldName: "collection.Name",
			},
			want: true,
		},
		{
			name: "ensure non-reserved field returns false",
			args: args{
				fieldName: "collection.nonReserved",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isReserved(tt.args.fieldName); got != tt.want {
				t.Errorf("isReserved() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSourceCodeFieldVariable(t *testing.T) {
	t.Parallel()

	fieldMarkerTest := &FieldMarker{
		Name:          "field.marker",
		sourceCodeVar: "parent.Spec.Field.Marker",
		Type:          FieldString,
	}

	collectionFieldMarkerTest := &CollectionFieldMarker{
		Name:          "collection",
		sourceCodeVar: "collection.Spec.Collection",
		Type:          FieldString,
	}

	intTest := &FieldMarker{
		Name:          "field.integer",
		sourceCodeVar: "parent.Spec.Field.Integer",
		Type:          FieldInt,
	}

	failTest := &FieldMarker{
		Name:          "field.fail",
		sourceCodeVar: "parent.Spec.Field.Fail",
		Type:          FieldBool,
	}

	type args struct {
		marker FieldMarkerProcessor
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "ensure field marker returns a correct source code variable field",
			args: args{
				marker: fieldMarkerTest,
			},
			want:    "!!start parent.Spec.Field.Marker !!end",
			wantErr: false,
		},
		{
			name: "ensure collection field marker returns a correct source code variable field",
			args: args{
				marker: collectionFieldMarkerTest,
			},
			want:    "!!start collection.Spec.Collection !!end",
			wantErr: false,
		},
		{
			name: "ensure integer field marker returns a correct source code variable field",
			args: args{
				marker: intTest,
			},
			want:    "!!start rune(parent.Spec.Field.Integer) !!end",
			wantErr: false,
		},
		{
			name: "ensure unsupported field marker returns an error",
			args: args{
				marker: failTest,
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := getSourceCodeFieldVariable(tt.args.marker)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSourceCodeFieldVariable() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("getSourceCodeFieldVariable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSourceCodeVariable(t *testing.T) {
	t.Parallel()

	fieldMarkerTest := &FieldMarker{
		Name: "this.is.a.highly.nested.field",
	}

	collectionFieldMarkerTest := &CollectionFieldMarker{
		Name: "flat",
	}

	fieldMarkerField := "test.field.marker.field"
	collectionFieldMarkerField := "test.collection.field.marker.field"

	resourceMarkerFieldTest := &ResourceMarker{
		Field: &fieldMarkerField,
	}

	resourceMarkerCollectionFieldTest := &ResourceMarker{
		CollectionField: &collectionFieldMarkerField,
	}

	type args struct {
		marker MarkerProcessor
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ensure field marker returns a correct source code variable",
			args: args{
				marker: fieldMarkerTest,
			},
			want: "parent.Spec.This.Is.A.Highly.Nested.Field",
		},
		{
			name: "ensure collection field marker returns a correct source code variable",
			args: args{
				marker: collectionFieldMarkerTest,
			},
			want: "collection.Spec.Flat",
		},
		{
			name: "ensure resource marker with field marker field returns a correct source code variable",
			args: args{
				marker: resourceMarkerFieldTest,
			},
			want: "parent.Spec.Test.Field.Marker.Field",
		},
		{
			name: "ensure resource marker with collection field marker field returns a correct source code variable",
			args: args{
				marker: resourceMarkerCollectionFieldTest,
			},
			want: "collection.Spec.Test.Collection.Field.Marker.Field",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getSourceCodeVariable(tt.args.marker); got != tt.want {
				t.Errorf("getSourceCodeVariable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getKeyValue(t *testing.T) {
	t.Parallel()

	testYamlNode := &yaml.Node{
		Tag:   "testTag",
		Value: "testValue",
	}

	testOtherYamlNode := &yaml.Node{
		Tag:   "testTag2",
		Value: "testValue2",
	}

	type args struct {
		result *inspect.YAMLResult
	}

	tests := []struct {
		name      string
		args      args
		wantKey   *yaml.Node
		wantValue *yaml.Node
	}{
		{
			name: "ensure flat result returns same key and value",
			args: args{
				result: &inspect.YAMLResult{
					Nodes: []*yaml.Node{testYamlNode},
				},
			},
			wantKey:   testYamlNode,
			wantValue: testYamlNode,
		},
		{
			name: "ensure multiple result returns correct key and value",
			args: args{
				result: &inspect.YAMLResult{
					Nodes: []*yaml.Node{
						testYamlNode,
						testOtherYamlNode,
					},
				},
			},
			wantKey:   testYamlNode,
			wantValue: testOtherYamlNode,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotKey, gotValue := getKeyValue(tt.args.result)
			if !reflect.DeepEqual(gotKey, tt.wantKey) {
				t.Errorf("getKeyValue() gotKey = %v, want %v", gotKey, tt.wantKey)
			}
			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("getKeyValue() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func Test_setValue(t *testing.T) {
	t.Parallel()

	//nolint: goconst
	testInvalidReplaceText := "*&^%"
	testReplaceText := "<replace me>"

	type args struct {
		marker FieldMarkerProcessor
		value  *yaml.Node
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *yaml.Node
	}{
		{
			name: "ensure value is set appropriately when replace text is not requested",
			args: args{
				marker: &FieldMarker{
					Name:          "test.field",
					sourceCodeVar: "parent.Spec.Test.Field",
				},
				value: &yaml.Node{
					Tag:   "testTag",
					Value: "test <replace me> value",
				},
			},
			wantErr: false,
			want: &yaml.Node{
				Tag:   "!!var",
				Value: "parent.Spec.Test.Field",
			},
		},
		{
			name: "ensure value is set appropriately when replace text is requested",
			args: args{
				marker: &FieldMarker{
					Name:          "test.field",
					Replace:       &testReplaceText,
					sourceCodeVar: "parent.Spec.Test.Field",
				},
				value: &yaml.Node{
					Tag:   "testTag",
					Value: "test <replace me> value",
				},
			},
			wantErr: false,
			want: &yaml.Node{
				Tag:   "!!str",
				Value: "test !!start parent.Spec.Test.Field !!end value",
			},
		},
		{
			name: "ensure invalid replace text returns an error",
			args: args{
				marker: &FieldMarker{
					Name:          "test.field",
					Replace:       &testInvalidReplaceText,
					sourceCodeVar: "parent.Spec.Test.Field",
				},
				value: &yaml.Node{
					Tag:   "testTag",
					Value: "test <replace me> value",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := setValue(tt.args.marker, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("setValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil {
				assert.Equal(t, tt.want, tt.args.value)
			}
		})
	}
}

func Test_setComments(t *testing.T) {
	t.Parallel()

	testDescription := "\n this\n is\n a\n test"
	testHeadCommentDescription := "\n# this\n# is\n# a\n# test"
	testName := "test.comment.field"
	testMarkerPrefix := "+operator-builder:field:default=\"my-field\",type=string"
	testMarkerText := fmt.Sprintf("%s,name=%s,description=`%s`", testMarkerPrefix, testName, testDescription)
	testHeadComment := fmt.Sprintf("# %s,name=%s,description=`%s`", testMarkerPrefix, testName, testHeadCommentDescription)

	type args struct {
		marker FieldMarkerProcessor
		result *inspect.YAMLResult
		key    *yaml.Node
		value  *yaml.Node
	}

	tests := []struct {
		name      string
		args      args
		wantKey   *yaml.Node
		wantValue *yaml.Node
	}{
		{
			name: "ensure head comment is set correctly with a description",
			args: args{
				marker: &FieldMarker{
					Name:        testName,
					Description: &testHeadCommentDescription,
				},
				result: &inspect.YAMLResult{
					Result: &parser.Result{
						MarkerText: testMarkerText,
					},
				},
				key: &yaml.Node{
					HeadComment: testHeadComment,
				},
				value: &yaml.Node{
					LineComment: testHeadComment,
				},
			},
			wantKey: &yaml.Node{
				FootComment: "",
				HeadComment: "# controlled by field: test.comment.field\n# # this\n# is\n# a\n# test",
			},
		},
		{
			name: "ensure head comment is set correctly without a description",
			args: args{
				marker: &FieldMarker{
					Name: testName,
				},
				result: &inspect.YAMLResult{
					Result: &parser.Result{
						MarkerText: testMarkerText,
					},
				},
				key: &yaml.Node{
					HeadComment: testHeadComment,
				},
				value: &yaml.Node{
					LineComment: testHeadComment,
				},
			},
			wantKey: &yaml.Node{
				FootComment: "",
				HeadComment: "# controlled by field: test.comment.field",
			},
		},
		{
			name: "ensure line comment is set correctly with a description",
			args: args{
				marker: &CollectionFieldMarker{
					Name:        testName,
					Description: &testHeadCommentDescription,
				},
				result: &inspect.YAMLResult{
					Result: &parser.Result{
						MarkerText: testMarkerText,
					},
				},
				key: &yaml.Node{
					HeadComment: testHeadComment,
				},
				value: &yaml.Node{
					LineComment: testHeadComment,
				},
			},
			wantValue: &yaml.Node{
				LineComment: "# controlled by collection field: test.comment.field",
			},
		},
		{
			name: "ensure line comment is set correctly without a description",
			args: args{
				marker: &CollectionFieldMarker{
					Name: testName,
				},
				result: &inspect.YAMLResult{
					Result: &parser.Result{
						MarkerText: testMarkerText,
					},
				},
				key: &yaml.Node{
					HeadComment: testHeadComment,
				},
				value: &yaml.Node{
					LineComment: testHeadComment,
				},
			},
			wantValue: &yaml.Node{
				LineComment: "# controlled by collection field: test.comment.field",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			setComments(tt.args.marker, tt.args.result, tt.args.key, tt.args.value)
			if tt.wantKey != nil {
				assert.Equal(t, tt.wantKey, tt.args.key)
			}
			if tt.wantValue != nil {
				assert.Equal(t, tt.wantValue, tt.args.value)
			}
		})
	}
}

func Test_transformYAML(t *testing.T) {
	t.Parallel()

	badReplaceText := "*&^%"

	type args struct {
		results []*inspect.YAMLResult
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ensure valid marker does not return error",
			args: args{
				results: []*inspect.YAMLResult{
					{
						Result: &parser.Result{
							MarkerText: "test",
							Object: FieldMarker{
								Name: "real.field",
							},
						},
						Nodes: []*yaml.Node{
							{
								Tag:   "test",
								Value: "test",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "ensure invalid object skips and returns no error",
			args: args{
				results: []*inspect.YAMLResult{
					{
						Result: &parser.Result{
							MarkerText: "test",
							Object:     "this is a string no a marker",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "ensure invalid field marker returns an error",
			args: args{
				results: []*inspect.YAMLResult{
					{
						Result: &parser.Result{
							MarkerText: "test",
							Object: FieldMarker{
								Name: "collection.name",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "ensure invalid collection field marker returns an error",
			args: args{
				results: []*inspect.YAMLResult{
					{
						Result: &parser.Result{
							MarkerText: "test",
							Object: CollectionFieldMarker{
								Name: "collection.name",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "ensure failure while attempting to set value",
			args: args{
				results: []*inspect.YAMLResult{
					{
						Result: &parser.Result{
							MarkerText: "test",
							Object: CollectionFieldMarker{
								Name:    "real.field",
								Replace: &badReplaceText,
							},
						},
						Nodes: []*yaml.Node{
							{
								Tag:   "test",
								Value: "test",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := transformYAML(tt.args.results...); (err != nil) != tt.wantErr {
				t.Errorf("transformYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
