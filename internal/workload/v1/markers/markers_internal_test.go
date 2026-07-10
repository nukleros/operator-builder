// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package markers

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/nukleros/markers/inspect"
	"github.com/nukleros/markers/parser"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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

	fieldMarkerTestString := "field.marker"
	collectionFieldMarkerTestString := "collection"
	intMarkerTest := "field.integer"
	failTestString := "field.fail"

	fieldMarkerTest := &FieldMarker{
		Name:          &fieldMarkerTestString,
		sourceCodeVar: "parent.Spec.Field.Marker",
		Type:          FieldString,
	}

	collectionFieldMarkerTest := &CollectionFieldMarker{
		Name:          &collectionFieldMarkerTestString,
		sourceCodeVar: "collection.Spec.Collection",
		Type:          FieldString,
	}

	intTest := &FieldMarker{
		Name:          &intMarkerTest,
		sourceCodeVar: "parent.Spec.Field.Integer",
		Type:          FieldInt,
	}

	failTest := &FieldMarker{
		Name:          &failTestString,
		sourceCodeVar: "parent.Spec.Field.Fail",
		Type:          FieldUnknownType,
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
			want:    "!!start strconv.Itoa(parent.Spec.Field.Integer) !!end",
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

	highlyNested := "this.is.a.highly.nested.field"
	flat := "flat"

	fieldMarkerTest := &FieldMarker{
		Name: &highlyNested,
	}

	collectionFieldMarkerTest := &CollectionFieldMarker{
		Name: &flat,
	}

	validParent := "metadata.name"
	invalidParent := "metadata.namespace"

	fieldMarkerParentTest := &FieldMarker{
		Parent: &validParent,
	}

	fieldMarkerInvalidParentTest := &FieldMarker{
		Parent: &invalidParent,
	}

	collectionFieldMarkerParentTest := &CollectionFieldMarker{
		Parent: &validParent,
	}

	collectionFieldMarkerInvalidParentTest := &CollectionFieldMarker{
		Parent: &invalidParent,
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
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "ensure field marker returns a correct source code variable",
			args: args{
				marker: fieldMarkerTest,
			},
			want:    "parent.Spec.This.Is.A.Highly.Nested.Field",
			wantErr: false,
		},
		{
			name: "ensure collection field marker returns a correct source code variable",
			args: args{
				marker: collectionFieldMarkerTest,
			},
			want:    "collection.Spec.Flat",
			wantErr: false,
		},
		{
			name: "ensure resource marker with field marker field returns a correct source code variable",
			args: args{
				marker: resourceMarkerFieldTest,
			},
			want:    "parent.Spec.Test.Field.Marker.Field",
			wantErr: false,
		},
		{
			name: "ensure resource marker with collection field marker field returns a correct source code variable",
			args: args{
				marker: resourceMarkerCollectionFieldTest,
			},
			want:    "collection.Spec.Test.Collection.Field.Marker.Field",
			wantErr: false,
		},
		{
			name: "ensure field marker with parent returns a correct source code variable",
			args: args{
				marker: fieldMarkerParentTest,
			},
			want:    "parent.Name",
			wantErr: false,
		},
		{
			name: "ensure field marker with invalid parent returns an error",
			args: args{
				marker: fieldMarkerInvalidParentTest,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "ensure collection field marker with parent returns a correct source code variable",
			args: args{
				marker: collectionFieldMarkerParentTest,
			},
			want:    "collection.Name",
			wantErr: false,
		},
		{
			name: "ensure collection field marker with invalid parent returns an error",
			args: args{
				marker: collectionFieldMarkerInvalidParentTest,
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := getSourceCodeVariable(tt.args.marker)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSourceCodeVariable() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
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

	testInvalidReplaceText := "*&^%"
	testReplaceText := "<replace me>"
	testField := "test.field.set"

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
					Name:          &testField,
					sourceCodeVar: "parent.Spec.Test.Field.Set",
					Type:          FieldString,
				},
				value: &yaml.Node{
					Tag:   "testTag",
					Value: "test <replace me> value",
				},
			},
			wantErr: false,
			want: &yaml.Node{
				Tag:   "!!var",
				Value: "parent.Spec.Test.Field.Set",
			},
		},
		{
			name: "ensure value is set appropriately when replace text is requested",
			args: args{
				marker: &FieldMarker{
					Name:          &testField,
					Replace:       &testReplaceText,
					sourceCodeVar: "parent.Spec.Test.Field.Set",
					Type:          FieldString,
				},
				value: &yaml.Node{
					Tag:   "testTag",
					Value: "test <replace me> value",
				},
			},
			wantErr: false,
			want: &yaml.Node{
				Tag:   "!!str",
				Value: "test !!start parent.Spec.Test.Field.Set !!end value",
			},
		},
		{
			name: "ensure invalid replace text returns an error",
			args: args{
				marker: &FieldMarker{
					Name:          &testField,
					Replace:       &testInvalidReplaceText,
					sourceCodeVar: "parent.Spec.Test.Field.Set",
					Type:          FieldString,
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
					Name:        &testName,
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
					Name: &testName,
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
					Name:        &testName,
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
					Name: &testName,
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
	realField := "real.field"
	collectionName := "collection.name"
	testString := "transformTest"

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
							MarkerText: testString,
							Object: FieldMarker{
								Name: &realField,
							},
						},
						Nodes: []*yaml.Node{
							{
								Tag:   testString,
								Value: testString,
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
							MarkerText: testString,
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
							MarkerText: testString,
							Object: FieldMarker{
								Name: &collectionName,
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
							MarkerText: testString,
							Object: CollectionFieldMarker{
								Name: &collectionName,
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
							MarkerText: testString,
							Object: CollectionFieldMarker{
								Name:    &realField,
								Replace: &badReplaceText,
							},
						},
						Nodes: []*yaml.Node{
							{
								Tag:   testString,
								Value: testString,
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

func Test_isCommentLineException(t *testing.T) {
	t.Parallel()

	type args struct {
		line       string
		exceptions []string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ensure line matching exception prefix returns true",
			args: args{
				line:       "+kubebuilder:validation:Optional",
				exceptions: []string{"+kubebuilder:"},
			},
			want: true,
		},
		{
			name: "ensure line not matching any exception returns false",
			args: args{
				line:       "this is a normal description line",
				exceptions: []string{"+kubebuilder:"},
			},
			want: false,
		},
		{
			name: "ensure empty exceptions list always returns false",
			args: args{
				line:       "+kubebuilder:validation:Optional",
				exceptions: []string{},
			},
			want: false,
		},
		{
			name: "ensure empty line with no exceptions returns false",
			args: args{
				line:       "",
				exceptions: []string{"+kubebuilder:"},
			},
			want: false,
		},
		{
			name: "ensure line matching one of multiple exceptions returns true",
			args: args{
				line:       "+operator-builder:field",
				exceptions: []string{"+kubebuilder:", "+operator-builder:"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isCommentLineException(tt.args.line, tt.args.exceptions); got != tt.want {
				t.Errorf("isCommentLineException() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wrapCommentLine(t *testing.T) {
	t.Parallel()

	type args struct {
		text string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "ensure empty string returns nil",
			args: args{text: ""},
			want: nil,
		},
		{
			name: "ensure whitespace-only string returns nil",
			args: args{text: "   "},
			want: nil,
		},
		{
			name: "ensure short line is returned as a single element",
			args: args{text: "short description"},
			want: []string{"short description"},
		},
		{
			name: "ensure line exactly at wrap width is not split",
			args: args{
				// 80 characters exactly
				text: "aaaaaaaaaa bbbbbbbbbb cccccccccc dddddddddd eeeeeeeeee ffffffffff gggggggggg h",
			},
			want: []string{"aaaaaaaaaa bbbbbbbbbb cccccccccc dddddddddd eeeeeeeeee ffffffffff gggggggggg h"},
		},
		{
			name: "ensure line exceeding wrap width is split at word boundary",
			args: args{
				text: "This is a description that is long enough to exceed the eighty character wrap width limit.",
			},
			want: []string{
				"This is a description that is long enough to exceed the eighty character wrap",
				"width limit.",
			},
		},
		{
			name: "ensure single word longer than wrap width is returned on one line",
			args: args{
				text: "us-central1-docker.pkg.dev/wanaware-core-dev/function-integrator/function-integrator:latest",
			},
			want: []string{
				"us-central1-docker.pkg.dev/wanaware-core-dev/function-integrator/function-integrator:latest",
			},
		},
		{
			name: "ensure double-space between words is preserved in output",
			args: args{
				text: "End of sentence.  Start of next sentence.",
			},
			want: []string{"End of sentence.  Start of next sentence."},
		},
		{
			name: "ensure long text is split into multiple lines",
			args: args{
				// "fourteen" ends at exactly column 80, so it stays on the first line;
				// the split happens before "fifteen".
				text: "one two three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen sixteen",
			},
			want: []string{
				"one two three four five six seven eight nine ten eleven twelve thirteen fourteen",
				"fifteen sixteen",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := wrapCommentLine(tt.args.text)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wrapCommentLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wrapCommentBlock(t *testing.T) {
	t.Parallel()

	type args struct {
		lines      []string
		exceptions []string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "ensure empty block returns nil",
			args: args{
				lines:      []string{},
				exceptions: []string{"+kubebuilder:"},
			},
			want: nil,
		},
		{
			name: "ensure block with single short line is returned as-is",
			args: args{
				lines:      []string{"a short description"},
				exceptions: []string{"+kubebuilder:"},
			},
			want: []string{"a short description"},
		},
		{
			name: "ensure multiple lines are joined and word-wrapped together",
			args: args{
				lines:      []string{"first line", "second line", "third line"},
				exceptions: []string{"+kubebuilder:"},
			},
			want: []string{"first line second line third line"},
		},
		{
			name: "ensure exception line is emitted verbatim without joining",
			args: args{
				lines:      []string{"+kubebuilder:validation:Optional"},
				exceptions: []string{"+kubebuilder:"},
			},
			want: []string{"+kubebuilder:validation:Optional"},
		},
		{
			name: "ensure exception lines are flushed between non-exception lines",
			args: args{
				lines:      []string{"before exception", "+kubebuilder:validation:Optional", "after exception"},
				exceptions: []string{"+kubebuilder:"},
			},
			want: []string{"before exception", "+kubebuilder:validation:Optional", "after exception"},
		},
		{
			name: "ensure all-exception block emits each line verbatim",
			args: args{
				lines:      []string{"+kubebuilder:default=true", "+kubebuilder:validation:Optional"},
				exceptions: []string{"+kubebuilder:"},
			},
			want: []string{"+kubebuilder:default=true", "+kubebuilder:validation:Optional"},
		},
		{
			name: "ensure long joined line is word-wrapped at wrap width",
			args: args{
				lines: []string{
					"This is the first part of a description",
					"that when joined together exceeds the eighty character wrap width boundary.",
				},
				exceptions: []string{"+kubebuilder:"},
			},
			want: []string{
				"This is the first part of a description that when joined together exceeds the",
				"eighty character wrap width boundary.",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := wrapCommentBlock(tt.args.lines, tt.args.exceptions)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wrapCommentBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_commentsFromMarker(t *testing.T) {
	t.Parallel()

	type args struct {
		description string
		exceptions  []string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "ensure empty description returns nil",
			args: args{
				description: "",
				exceptions:  []string{"+kubebuilder:"},
			},
			want: nil,
		},
		{
			name: "ensure newline-only description returns nil",
			args: args{
				description: "\n\n\n",
				exceptions:  []string{"+kubebuilder:"},
			},
			want: nil,
		},
		{
			name: "ensure single-paragraph description has leading blank separator",
			args: args{
				description: "\n a short description",
				exceptions:  []string{"+kubebuilder:"},
			},
			want: []string{"", "a short description"},
		},
		{
			name: "ensure two paragraphs are separated by exactly one blank line",
			args: args{
				description: "\n first paragraph\n\n second paragraph",
				exceptions:  []string{"+kubebuilder:"},
			},
			want: []string{"", "first paragraph", "", "second paragraph"},
		},
		{
			name: "ensure double blank line between paragraphs is normalized to one blank",
			args: args{
				description: "\n first paragraph\n\n\n second paragraph",
				exceptions:  []string{"+kubebuilder:"},
			},
			want: []string{"", "first paragraph", "", "second paragraph"},
		},
		{
			name: "ensure description with leading and trailing newlines is trimmed",
			args: args{
				description: "\n\n description text \n\n",
				exceptions:  []string{"+kubebuilder:"},
			},
			want: []string{"", "description text"},
		},
		{
			name: "ensure exception lines in description are passed through verbatim",
			args: args{
				description: "\n description text\n +kubebuilder:validation:Optional",
				exceptions:  []string{"+kubebuilder:"},
			},
			want: []string{"", "description text", "+kubebuilder:validation:Optional"},
		},
		{
			name: "ensure three-paragraph description produces correct structure",
			args: args{
				description: "\n first paragraph\n\n second paragraph\n\n third paragraph",
				exceptions:  []string{"+kubebuilder:"},
			},
			want: []string{"", "first paragraph", "", "second paragraph", "", "third paragraph"},
		},
		{
			name: "ensure long description line is word-wrapped",
			args: args{
				description: "\n This is a description that is long enough to exceed the eighty character wrap width limit set by commentWrapWidth.",
				exceptions:  []string{"+kubebuilder:"},
			},
			want: []string{
				"",
				"This is a description that is long enough to exceed the eighty character wrap",
				"width limit set by commentWrapWidth.",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := commentsFromMarker(tt.args.description, tt.args.exceptions...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("commentsFromMarker() = %v, want %v", got, tt.want)
			}
		})
	}
}
