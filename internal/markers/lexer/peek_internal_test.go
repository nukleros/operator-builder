// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package lexer

import (
	"bytes"
	"testing"
)

func getTestLexer(buf string) *Lexer {
	return NewLexer(bytes.NewBufferString(buf))
}

func Test_lexer_peek(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		lexer *Lexer
		wantR rune
	}{
		{
			name:  "successfully peeks a rune",
			lexer: getTestLexer("Hello World"),
			wantR: 'H',
		},
		{
			name:  "successfully peeks a rune in one char string",
			lexer: getTestLexer("H"),
			wantR: 'H',
		},
		{
			name:  "successfully peeks new line rune in one char string",
			lexer: getTestLexer("\n"),
			wantR: '\n',
		},
		{
			name:  "successfully peeks eof rune in empty string",
			lexer: getTestLexer(""),
			wantR: -1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if gotR := tt.lexer.peek(); gotR != tt.wantR {
				t.Errorf("lexer.peek() = %v, want %v", gotR, tt.wantR)
			}
		})
	}
}

func Test_lexer_peekN(t *testing.T) {
	t.Parallel()

	type args struct {
		i int
	}

	tests := []struct {
		name  string
		args  args
		lexer *Lexer
		wantR []rune
	}{
		{
			name: "given 2, peeks 2 runes",
			args: args{
				i: 2,
			},
			lexer: getTestLexer("Hello World"),
			wantR: []rune{'H', 'e'},
		},
		{
			name: "given 2, peeks 2 runes in single char string",
			args: args{
				i: 2,
			},
			lexer: getTestLexer("H"),
			wantR: []rune{'H', -1},
		},
		{
			name: "successfully peeks new line rune in one char string",
			args: args{
				i: 2,
			},
			lexer: getTestLexer("H\n"),
			wantR: []rune{'H', '\n'},
		},
		{
			name: "given 2, peeks 1 rune in empty string",
			args: args{
				i: 2,
			},
			lexer: getTestLexer(""),
			wantR: []rune{-1},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotR := tt.lexer.peekN(tt.args.i)
			if len(gotR) != len(tt.wantR) {
				t.Fatalf("lexer.peekN(%v) = %v, want %v", tt.args.i, gotR, tt.wantR)
			}

			for i := range gotR {
				if gotR[i] != tt.wantR[i] {
					t.Errorf("lexer.peekN(%v) = %v, want %v", tt.args.i, gotR, tt.wantR)
				}
			}
		})
	}
}

func Test_lexer_peeked(t *testing.T) {
	t.Parallel()

	type args struct {
		token  string
		except []string
	}

	tests := []struct {
		name  string
		lexer *Lexer
		args  args
		want  bool
	}{
		{
			name:  "returns true if token is found and no exceptions",
			lexer: getTestLexer("Hello World"),
			args: args{
				"Hello",
				nil,
			},
			want: true,
		},
		{
			name:  "returns false if token is found and but followed by an exception",
			lexer: getTestLexer("HelloWorld"),
			args: args{
				"Hello",
				[]string{"W"},
			},
			want: false,
		},
		{
			name:  "returns false if token is not found",
			lexer: getTestLexer("HelloWorld"),
			args: args{
				"Goodbye",
				nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.lexer.peeked(tt.args.token, tt.args.except...); got != tt.want {
				t.Errorf("lexer.peeked(%v, %v) = %v, want %v", tt.args.token, tt.args.except, got, tt.want)
			}
		})
	}
}

func Test_lexer_peekedWhitespaced(t *testing.T) {
	t.Parallel()

	type args struct {
		tokens []string
	}

	tests := []struct {
		name  string
		lexer *Lexer
		args  args
		want  bool
	}{
		{
			name:  "returns true if token is found with whitespace",
			lexer: getTestLexer("  Hello World"),
			args: args{
				[]string{"Hello"},
			},
			want: true,
		},
		{
			name:  "returns true if token is found with no whitespace",
			lexer: getTestLexer("HelloWorld"),
			args: args{
				[]string{"Hello"},
			},
			want: true,
		},
		{
			name:  "returns false if token is not found",
			lexer: getTestLexer("HelloWorld"),
			args: args{
				[]string{"Goodbye"},
			},
			want: false,
		},
		{
			name:  "returns false if eof is reached",
			lexer: getTestLexer("    "),
			args: args{
				[]string{"Hello"},
			},
			want: false,
		},
		{
			name:  "returns true if token is found after new line",
			lexer: getTestLexer("    \nHello"),
			args: args{
				[]string{"Hello"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.lexer.peekedWhitespaced(tt.args.tokens...); got != tt.want {
				t.Errorf("lexer.peekedWhitespaced(%v) = %v, want %v", tt.args.tokens, got, tt.want)
			}
		})
	}
}
