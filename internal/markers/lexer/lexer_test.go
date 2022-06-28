// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package lexer_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/lexer"
)

func GetTestLexer(buf string) *lexer.Lexer {
	return lexer.NewLexer(bytes.NewBufferString(buf))
}

func TestLexer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []lexer.Lexeme
		focus    bool // if true, run only tests with focus set to true
	}{
		{
			name:  "marker start",
			input: "+test:flag",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "test"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "flag"},
				{Type: lexer.LexemeSyntheticBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "invalid marker start",
			input: "++",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "math operation",
			input: "2+2=4",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker flag with no scope",
			input: "+hello",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeWarning, Value: `marker without scope found at position: {line:1 column:7}, following "+hello"`},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker flag with scope",
			input: "+hello:world",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "hello"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "world"},
				{Type: lexer.LexemeSyntheticBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker flag with two scopes",
			input: "+hello:new:world",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "hello"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeScope, Value: "new"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "world"},
				{Type: lexer.LexemeSyntheticBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker arg with no scope",
			input: "+planet=earth",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeWarning, Value: `marker without scope found at position: {line:1 column:8}, following "+planet"`},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker arg with scope",
			input: "+galaxy:planet=earth",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "planet"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "earth"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker arg with two scopes",
			input: "+galaxy:planet:name=earth",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeScope, Value: "planet"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "earth"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker with two args",
			input: "+planet:name=earth,solar-system=milky-way",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "planet"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "earth"},
				{Type: lexer.LexemeArgDelimiter, Value: ","},
				{Type: lexer.LexemeArg, Value: "solar-system"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "milky-way"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker with two scopes and two args",
			input: "+galaxy:planet:name=earth,solar-system=milky-way",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeScope, Value: "planet"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "earth"},
				{Type: lexer.LexemeArgDelimiter, Value: ","},
				{Type: lexer.LexemeArg, Value: "solar-system"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "milky-way"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker with two scopes and two args one of which is a flag",
			input: "+galaxy:planet:name=earth,current-location",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeScope, Value: "planet"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "earth"},
				{Type: lexer.LexemeArgDelimiter, Value: ","},
				{Type: lexer.LexemeArg, Value: "current-location"},
				{Type: lexer.LexemeSyntheticBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker with single quoted string arg",
			input: "+galaxy:name=milkyway,description='our home system'",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "milkyway"},
				{Type: lexer.LexemeArgDelimiter, Value: ","},
				{Type: lexer.LexemeArg, Value: "description"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeQuote, Value: "'"},
				{Type: lexer.LexemeStringLiteral, Value: "our home system"},
				{Type: lexer.LexemeQuote, Value: "'"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker with double quoted string arg",
			input: `+galaxy:name=milkyway,description="our home system"`,
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "milkyway"},
				{Type: lexer.LexemeArgDelimiter, Value: ","},
				{Type: lexer.LexemeArg, Value: "description"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeQuote, Value: "\""},
				{Type: lexer.LexemeStringLiteral, Value: "our home system"},
				{Type: lexer.LexemeQuote, Value: "\""},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker with literal quoted string arg",
			input: "+galaxy:name=milkyway,description=`our home system`",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "milkyway"},
				{Type: lexer.LexemeArgDelimiter, Value: ","},
				{Type: lexer.LexemeArg, Value: "description"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeQuote, Value: "`"},
				{Type: lexer.LexemeStringLiteral, Value: "our home system"},
				{Type: lexer.LexemeQuote, Value: "`"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name: "marker with literal quoted multi-line string arg",
			input: `+galaxy:name=milkyway,description=` + "`" + `our home system
			this is where planet earth is located` + "`",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "milkyway"},
				{Type: lexer.LexemeArgDelimiter, Value: ","},
				{Type: lexer.LexemeArg, Value: "description"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeQuote, Value: "`"},
				{Type: lexer.LexemeStringLiteral, Value: "our home system\n\t\t\tthis is where planet earth is located"},
				{Type: lexer.LexemeQuote, Value: "`"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name: "marker with literal quoted multi-line string arg in yaml comment",
			input: `# +galaxy:name=milkyway,description=` + "`" + `our home system
			#this is where planet earth is located` + "`",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "#"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "milkyway"},
				{Type: lexer.LexemeArgDelimiter, Value: ","},
				{Type: lexer.LexemeArg, Value: "description"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeQuote, Value: "`"},
				{Type: lexer.LexemeStringLiteral, Value: "our home system\nthis is where planet earth is located"},
				{Type: lexer.LexemeQuote, Value: "`"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker in go comment no space",
			input: "//+hello:world",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "//"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "hello"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "world"},
				{Type: lexer.LexemeSyntheticBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker in go comment with white space",
			input: "//     +hello:world",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "//"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "hello"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "world"},
				{Type: lexer.LexemeSyntheticBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker in yaml comment no space",
			input: "#+hello:world",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "#"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "hello"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "world"},
				{Type: lexer.LexemeSyntheticBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker in yaml comment with white space",
			input: "#     +hello:world",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "#"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "hello"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "world"},
				{Type: lexer.LexemeSyntheticBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},

		{
			name: "marker with two args in context",
			input: `#+planet:name=earth,solar-system=milky-way
			plant: earth
			`,
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "#"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "planet"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "earth"},
				{Type: lexer.LexemeArgDelimiter, Value: ","},
				{Type: lexer.LexemeArg, Value: "solar-system"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "milky-way"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "fun with rich",
			input: `#+beetle-:dung:mature=0`,
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "#"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "beetle-"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeScope, Value: "dung"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "mature"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeIntegerLiteral, Value: "0"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "kubebuilder marker",
			input: `# +kubebuilder:validation:Enum=aws;azure;vmware`,
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "#"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "kubebuilder"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeScope, Value: "validation"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "Enum"},
				{Type: lexer.LexemeArgAssignment, Value: "="},
				{Type: lexer.LexemeStringLiteral, Value: "aws;azure;vmware"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
	}

	focused := false

	for _, tt := range tests {
		if tt.focus {
			focused = true

			break
		}
	}

	for _, tt := range tests {
		tt := tt
		if focused && !tt.focus {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			l := GetTestLexer(tt.input)
			go l.Run()
			actual := []lexer.Lexeme{}
			for {
				lexeme := l.NextLexeme()
				testLexeme := lexer.Lexeme{
					Type:  lexeme.Type,
					Value: lexeme.Value,
				}

				actual = append(actual, testLexeme)
				if lexeme.Type == lexer.LexemeEOF {
					break
				}
			}
			require.Equal(t, tt.expected, actual)
		})
	}

	if focused {
		t.Fatalf("testcase(s) still focussed")
	}
}

