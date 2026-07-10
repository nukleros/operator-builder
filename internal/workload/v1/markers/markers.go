// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package markers

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/nukleros/markers/inspect"
	markerparser "github.com/nukleros/markers/marker"
	"gopkg.in/yaml.v3"

	"github.com/nukleros/operator-builder/internal/utils"
)

var (
	ErrMissingReplaceText            = errors.New("marker is missing the requested replace text")
	ErrMissingParentOrName           = errors.New("missing either parent=value or name=value marker")
	ErrInvalidReplaceMarkerFieldType = errors.New("invalid marker type using replace")
	ErrInvalidParentField            = errors.New("invalid parent field")
)

// MarkerType defines the types of markers that are accepted by the parser.
type MarkerType int

const (
	FieldMarkerType MarkerType = iota
	CollectionMarkerType
	ResourceMarkerType
	UnknownMarkerType
)

// FieldMarkerProcessor is an interface that requires specific methods that are
// necessary for parsing a field marker or a collection field marker.
type FieldMarkerProcessor interface {
	GetName() string
	GetDefault() interface{}
	GetDescription() string
	GetFieldType() FieldType
	GetOriginalValue() interface{}
	GetParent() string
	GetReplaceText() string
	GetSpecPrefix() string
	GetSourceCodeVariable() string
	GetComments(exceptions ...string) []string

	IsCollectionFieldMarker() bool
	IsFieldMarker() bool
	IsForCollection() bool
	IsArbitrary() bool

	SetDescription(string)
	SetOriginalValue(string)
	SetForCollection(bool)
}

// MarkerProcessor is a more generic interface that requires specific methods that are
// necessary for parsing any type of marker.
type MarkerProcessor interface {
	GetName() string
	GetPrefix() string
	GetSpecPrefix() string
	GetParent() string
}

// MarkerCollection is an object that stores a set of markers.
type MarkerCollection struct {
	FieldMarkers           []*FieldMarker
	CollectionFieldMarkers []*CollectionFieldMarker
}

// ContainsMarkerType will determine if a given marker type exists within
// a set of marker types.
func ContainsMarkerType(s []MarkerType, e MarkerType) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

// InspectForYAML will inspect yamlContent for a set of markers.  It will find
// all of the markers within the yamlContent and return the resultant lines and
// any associated errors.
func InspectForYAML(yamlContent []byte, markerTypes ...MarkerType) ([]*yaml.Node, []*inspect.YAMLResult, error) {
	insp, err := initializeMarkerInspector(markerTypes...)
	if err != nil {
		return nil, nil, fmt.Errorf("%w; error initializing markers %v", err, markerTypes)
	}

	nodes, results, err := insp.InspectYAML(yamlContent, transformYAML)
	if err != nil {
		return nil, nil, fmt.Errorf("%w; error inspecting YAML for markers %v", err, markerTypes)
	}

	return nodes, results, nil
}

// initializeMarkerInspector will create a new registry and initialize an inspector
// for specific marker types.
func initializeMarkerInspector(markerTypes ...MarkerType) (*inspect.Inspector, error) {
	registry := markerparser.NewRegistry()

	var err error

	for i := range markerTypes {
		switch markerTypes[i] {
		case FieldMarkerType:
			err = defineFieldMarker(registry)
		case CollectionMarkerType:
			err = defineCollectionFieldMarker(registry)
		case ResourceMarkerType:
			err = defineResourceMarker(registry)
		}
	}

	if err != nil {
		return nil, err
	}

	return inspect.NewInspector(registry), nil
}

// transformYAML will transform a YAML result into the proper format for scaffolding
// resultant code and API definitions.
func transformYAML(results ...*inspect.YAMLResult) error {
	for _, result := range results {
		// convert to interface
		var marker FieldMarkerProcessor

		switch t := result.Object.(type) {
		case FieldMarker:
			sourceCodeVar, err := getSourceCodeVariable(&t)
			if err != nil {
				return err
			}

			t.sourceCodeVar = sourceCodeVar
			marker = &t
		case CollectionFieldMarker:
			sourceCodeVar, err := getSourceCodeVariable(&t)
			if err != nil {
				return err
			}

			t.sourceCodeVar = sourceCodeVar
			marker = &t
		default:
			continue
		}

		// ensure that either a parent or a name is set
		if marker.GetName() == "" && marker.GetParent() == "" {
			return fmt.Errorf("%w for marker %s", ErrMissingParentOrName, marker)
		}

		// get common variables and confirm that we are not working with a reserved marker
		if isReserved(marker.GetName()) {
			return fmt.Errorf("%s %w", marker.GetName(), ErrFieldMarkerReserved)
		}

		key, value := getKeyValue(result)

		setComments(marker, result, key, value)

		if err := setValue(marker, value); err != nil {
			return fmt.Errorf("%w; error setting value for marker %s", err, result.MarkerText)
		}

		result.Object = marker
	}

	return nil
}

// reservedMarkers represents a list of markers which cannot be used
// within a manifest.  They are reserved for internal purposes.  If any of the
// reservedMarkers are found, we will throw an error and notify the user.
func reservedMarkers() []string {
	return []string{
		"collection",
		"collection.name",
		"collection.namespace",
	}
}

// supportedParents represents a map of parent fields to their go variable values
// which are currently supported.
func supportedParents() map[string]string {
	return map[string]string{
		"metadata.name": "Name",
	}
}

// isReserved is a convenience method which returns whether or not a marker, given
// the fieldName as a string, is reserved for internal purposes.
func isReserved(fieldName string) bool {
	return validField(fieldName, reservedMarkers())
}

// isSupported is a convenience method which returns whether or not a marker, given
// the parentField as a string, is supported.
func isSupported(parentField string) bool {
	return supportedParents()[parentField] != ""
}

// validField determines if a field is valid based on a known list of valid fields.
func validField(field string, validFields []string) bool {
	for _, valid := range validFields {
		if utils.ToTitle(valid) == utils.ToTitle(field) {
			return true
		}
	}

	return false
}

const commentWrapWidth = 80

// commentsFromMarker builds the formatted comment slice for a field marker.
// It splits the description into word-wrapped blocks (at commentWrapWidth) and
// prepends a blank line so the description is visually separated from any
// preceding marker annotations (e.g. the "(Default: X)" that setDefault writes
// into api.Markers).  Any block line that begins with an exception prefix is
// emitted verbatim without word-wrapping.
//
// Note: the default annotation itself is intentionally NOT included here — it
// is already written to api.Markers by setDefault/setCommentsAndDefault.
func commentsFromMarker(description string, exceptions ...string) []string {
	description = strings.Trim(description, "\n")
	if description == "" {
		return nil
	}

	// Leading blank line separates the description from marker annotations
	// (e.g. +kubebuilder:default=X or (Default: X)) that precede it.
	comments := []string{""}

	for i, block := range buildDescriptionBlocks(strings.Split(description, "\n")) {
		blockLines := wrapCommentBlock(block, exceptions)
		if len(blockLines) == 0 {
			continue
		}

		if i > 0 {
			comments = append(comments, "")
		}

		comments = append(comments, blockLines...)
	}

	return normalizeCommentLines(comments)
}

// buildDescriptionBlocks splits trimmed non-blank lines into paragraph blocks,
// separated by blank (or whitespace-only) lines.
func buildDescriptionBlocks(rawLines []string) [][]string {
	var blocks [][]string

	var current []string

	for _, line := range rawLines {
		if trimmed := strings.TrimSpace(line); trimmed == "" {
			if len(current) > 0 {
				blocks = append(blocks, current)
				current = nil
			}
		} else {
			current = append(current, trimmed)
		}
	}

	if len(current) > 0 {
		blocks = append(blocks, current)
	}

	return blocks
}

// normalizeCommentLines collapses consecutive blank entries into one and removes
// any trailing blank entry.  Returns nil when nothing remains.
func normalizeCommentLines(comments []string) []string {
	normalized := comments[:0:0]

	for i, c := range comments {
		if c == "" && i > 0 && comments[i-1] == "" {
			continue
		}

		normalized = append(normalized, c)
	}

	for len(normalized) > 0 && normalized[len(normalized)-1] == "" {
		normalized = normalized[:len(normalized)-1]
	}

	if len(normalized) == 0 {
		return nil
	}

	return normalized
}

// wrapCommentBlock emits the lines of a description block as wrapped comment strings.
// Lines beginning with an exception prefix are passed through verbatim on their
// own line; all other lines are joined and word-wrapped at commentWrapWidth.
func wrapCommentBlock(lines, exceptions []string) []string {
	var result []string

	var pending []string

	flush := func() {
		if len(pending) == 0 {
			return
		}

		result = append(result, wrapCommentLine(strings.Join(pending, " "))...)
		pending = nil
	}

	for _, line := range lines {
		if isCommentLineException(line, exceptions) {
			flush()
			result = append(result, line)
		} else {
			pending = append(pending, line)
		}
	}

	flush()

	return result
}

var wordRe = regexp.MustCompile(`\S+`)

// wrapCommentLine wraps text at commentWrapWidth characters, breaking only at word
// boundaries.  Whitespace runs between words (e.g. double-spaces after a
// sentence-ending period) are preserved in the output.
func wrapCommentLine(text string) []string {
	if text == "" {
		return nil
	}
	locs := wordRe.FindAllStringIndex(text, -1)

	if len(locs) == 0 {
		return nil
	}

	var lines []string

	lineStart := locs[0][0]
	lineEnd := locs[0][1]

	for _, loc := range locs[1:] {
		wordEnd := loc[1]
		if wordEnd-lineStart > commentWrapWidth {
			lines = append(lines, text[lineStart:lineEnd])
			lineStart = loc[0]
		}

		lineEnd = wordEnd
	}

	lines = append(lines, text[lineStart:lineEnd])

	return lines
}

// isCommentLineException returns true when line begins with any of the exception prefixes.
func isCommentLineException(line string, exceptions []string) bool {
	for _, exc := range exceptions {
		if strings.HasPrefix(line, exc) {
			return true
		}
	}

	return false
}

// getSourceCodeFieldVariable gets a full variable name for a marker as it is intended to be
// passed into the generate package to generate the source code.  This includes particular
// tags that are needed by the generator to properly identify when a variable starts and ends.
func getSourceCodeFieldVariable(marker FieldMarkerProcessor) (string, error) {
	switch marker.GetFieldType() {
	case FieldString:
		return fmt.Sprintf("!!start %s !!end", marker.GetSourceCodeVariable()), nil
	case FieldInt:
		return fmt.Sprintf("!!start strconv.Itoa(%s) !!end", marker.GetSourceCodeVariable()), nil
	case FieldBool:
		return fmt.Sprintf("!!start strconv.FormatBool(%s) !!end", marker.GetSourceCodeVariable()), nil
	default:
		return "", fmt.Errorf("%w with field type %s", ErrInvalidReplaceMarkerFieldType, marker.GetFieldType())
	}
}

// getSourceCodeVariable gets a full variable name for a marker as it is intended to be
// scaffolded in the source code.
func getSourceCodeVariable(marker MarkerProcessor) (string, error) {
	if marker.GetParent() == "" {
		return fmt.Sprintf("%s.%s", marker.GetSpecPrefix(), utils.ToTitle(marker.GetName())), nil
	}

	if isSupported(marker.GetParent()) {
		return fmt.Sprintf("%s.%s", marker.GetPrefix(), supportedParents()[marker.GetParent()]), nil
	}

	return "", fmt.Errorf("%w %s. supported parent fields are: %v", ErrInvalidParentField, marker.GetParent(), supportedParents())
}

// getKeyValue gets the key and value from a YAML result.
func getKeyValue(result *inspect.YAMLResult) (key, value *yaml.Node) {
	if len(result.Nodes) > 1 {
		return result.Nodes[0], result.Nodes[1]
	}

	return result.Nodes[0], result.Nodes[0]
}

// setComments sets the comments for use by the resultant code.
func setComments(marker FieldMarkerProcessor, result *inspect.YAMLResult, key, value *yaml.Node) {
	// update the description to ensure new lines are commented
	if marker.GetDescription() != "" {
		marker.SetDescription(strings.TrimPrefix(marker.GetDescription(), "\n"))
		key.HeadComment = key.HeadComment + "\n# " + marker.GetDescription()
	}

	// set replace text to ensure that our markers are commented
	replaceText := strings.TrimSuffix(result.MarkerText, "\n")
	replaceText = strings.ReplaceAll(replaceText, "\n", "\n#")

	// set the append text to notify the user where a marker was originated from in their source code
	var appendText string
	switch t := marker.(type) {
	case *FieldMarker:
		appendText = "controlled by field: " + t.GetName()
	case *CollectionFieldMarker:
		appendText = "controlled by collection field: " + t.GetName()
	}

	// set the comments on the yaml nodes
	key.FootComment = ""
	key.HeadComment = strings.ReplaceAll(key.HeadComment, replaceText, appendText)
	value.LineComment = strings.ReplaceAll(value.LineComment, replaceText, appendText)
}

// setValue will set the value appropriately.  If the marker is arbitrary, no
// value need be set - return immediately.  If the marker has requested
// replacement text this is set.
func setValue(marker FieldMarkerProcessor, value *yaml.Node) error {
	if marker.IsArbitrary() {
		return nil
	}

	const varTag = "!!var"

	const strTag = "!!str"

	markerReplaceText := marker.GetReplaceText()

	marker.SetOriginalValue(value.Value)

	if markerReplaceText != "" {
		value.Tag = strTag

		re, err := regexp.Compile(markerReplaceText)
		if err != nil {
			return fmt.Errorf("unable to convert %s to regex, %w", markerReplaceText, err)
		}

		fieldVar, err := getSourceCodeFieldVariable(marker)
		if err != nil {
			return fmt.Errorf("unable to get source code field variable for marker %s, %w", marker, err)
		}

		if !strings.Contains(value.Value, markerReplaceText) {
			return fmt.Errorf("replace text=[%s] value=[%s], %w", markerReplaceText, value.Value, ErrMissingReplaceText)
		}

		value.Value = re.ReplaceAllString(value.Value, fieldVar)
	} else {
		value.Tag = varTag
		value.Value = marker.GetSourceCodeVariable()
	}

	return nil
}
