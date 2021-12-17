// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package lexer

type LexemeType int

const (
	LexemeError LexemeType = iota
	LexemeComment
	LexemeMarkerStart
	LexemeScope
	LexemeSeparator
	LexemeArg
	LexemeArgAssignment
	LexemeArgDelimiter
	LexemeStringLiteral
	LexemeFloatLiteral
	LexemeIntegerLiteral
	LexemeSyntheticBoolLiteral
	LexemeBoolLiteral
	LexemeQuote
	LexemeSliceBegin
	LexemeSliceEnd
	LexemeSliceDelimiter
	LexemeNakedSliceDelimiter
	LexemeMarkerEnd
	LexemeWarning
	LexemeEOF
)

type Lexeme struct {
	Type  LexemeType
	Value string
	Pos   position
}

const eof = -1

const (
	golangComment   = "//"
	yamlComment     = "#"
	markerStart     = "+"
	markerSeparator = ":"
	argAssignment   = "="
	argDelimiter    = ","
	literalQuote    = "`"
	doubleQuote     = `"`
	singleQuote     = `'`
)

func (l Lexeme) String() string {
	return l.Value
}
