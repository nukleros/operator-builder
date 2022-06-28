// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package lexer

import (
	"testing"
)

func Test_lexer_consume(t *testing.T) {
	t.Parallel()

	type args struct {
		s string
	}

	tests := []struct {
		name    string
		lexer   *Lexer
		args    args
		wantBuf string
		wantPos position
	}{
		{
			name:    "consumes tokens correctly",
			lexer:   getTestLexer("Hello World"),
			args:    args{s: "Hello"},
			wantBuf: "Hello",
			wantPos: position{line: 1, column: 6},
		},
		{
			name:    "consumes tokens across lines correctly",
			lexer:   getTestLexer("Hello \nWorld"),
			args:    args{s: "Hello \nWorld"},
			wantBuf: "Hello \nWorld",
			wantPos: position{line: 2, column: 6},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.lexer.consume(tt.args.s)
			if tt.lexer.buffer != tt.wantBuf {
				t.Errorf("lexer.consume(%v).buffer = %v, want %v", tt.args.s, tt.lexer.buffer, tt.wantBuf)
			}
			if tt.lexer.pos != tt.wantPos {
				t.Errorf("lexer.consume(%v).pos = %v, want %v", tt.args.s, tt.lexer.pos, tt.wantPos)
			}
		})
	}
}

func Test_lexer_consumed(t *testing.T) {
	t.Parallel()

	type args struct {
		token  string
		except []string
	}

	tests := []struct {
		name    string
		lexer   *Lexer
		args    args
		want    bool
		wantBuf string
		wantPos position
	}{
		{
			name:  "consumes tokens correctly",
			lexer: getTestLexer("Hello World"),
			args: args{
				token:  "Hello",
				except: nil,
			},
			want:    true,
			wantBuf: "Hello",
			wantPos: position{line: 1, column: 6},
		},
		{
			name:  "Does not consume token if followed by exception",
			lexer: getTestLexer("HelloWorld"),
			args: args{
				token:  "Hello",
				except: []string{"W"},
			},
			want:    false,
			wantBuf: "",
			wantPos: position{line: 1, column: 1},
		},
		{
			name:  "Does not consume token if not found",
			lexer: getTestLexer("Hello World"),
			args: args{
				token:  "GoodBye",
				except: nil,
			},
			want:    false,
			wantBuf: "",
			wantPos: position{line: 1, column: 1},
		},
		{
			name:  "consumes tokens across lines correctly",
			lexer: getTestLexer("Hello \nWorld"),
			args: args{
				token:  "Hello \nWorld",
				except: nil,
			},
			want:    true,
			wantBuf: "Hello \nWorld",
			wantPos: position{line: 2, column: 6},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.lexer.consumed(tt.args.token, tt.args.except...); got != tt.want {
				t.Errorf("lexer.consumed(%v, %v) = %v, want %v", tt.args.token, tt.args.except, got, tt.want)
			}

			if tt.lexer.buffer != tt.wantBuf {
				t.Errorf("lexer.consumed(%v, %v).buffer = %v, want %v", tt.args.token, tt.args.except, tt.lexer.buffer, tt.wantBuf)
			}

			if tt.lexer.pos != tt.wantPos {
				t.Errorf("lexer.consumed(%v, %v).pos = %v, want %v", tt.args.token, tt.args.except, tt.lexer.pos, tt.wantPos)
			}
		})
	}
}

func Test_lexer_consumedWhitespaced(t *testing.T) {
	t.Parallel()

	type args struct {
		tokens []string
	}

	tests := []struct {
		name    string
		lexer   *Lexer
		args    args
		want    bool
		wantBuf string
		wantPos position
	}{
		{
			name:  "consumes tokens correctly",
			lexer: getTestLexer("Hello World"),
			args: args{
				tokens: []string{"Hello"},
			},
			want:    true,
			wantBuf: "Hello",
			wantPos: position{line: 1, column: 6},
		},
		{
			name:  "consumes tokens separated by whitespace",
			lexer: getTestLexer("    Hello World"),
			args: args{
				tokens: []string{"Hello", "World"},
			},
			want:    true,
			wantBuf: "    Hello",
			wantPos: position{line: 1, column: 10},
		},
		{
			name:  "Does not consume token if not found",
			lexer: getTestLexer("Hello World"),
			args: args{
				tokens: []string{"GoodBye"},
			},
			want:    false,
			wantBuf: "",
			wantPos: position{line: 1, column: 1},
		},
		{
			name:  "consumes tokens across lines correctly",
			lexer: getTestLexer("   \nWorld"),
			args: args{
				tokens: []string{"World"},
			},
			want:    true,
			wantBuf: "   \nWorld",
			wantPos: position{line: 2, column: 6},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.lexer.consumedWhitespaced(tt.args.tokens...); got != tt.want {
				t.Errorf("lexer.consumedWhitespaced(%v) = %v, want %v", tt.args.tokens, got, tt.want)
			}

			if tt.lexer.buffer != tt.wantBuf {
				t.Errorf("lexer.consumedWhitespaced(%v).buffer = %v, want %v", tt.args.tokens, tt.lexer.buffer, tt.wantBuf)
			}

			if tt.lexer.pos != tt.wantPos {
				t.Errorf("lexer.consumedWhitespaced(%v).pos = %v, want %v", tt.args.tokens, tt.lexer.pos, tt.wantPos)
			}
		})
	}
}

func Test_lexer_consumeWhitespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		lexer   *Lexer
		wantBuf string
		wantPos position
	}{
		{
			name:    "consumes whitespace correctly",
			lexer:   getTestLexer("   \n\tHello World"),
			wantBuf: "   \n\t",
			wantPos: position{line: 2, column: 2},
		},
		{
			name:    "does not consume non whitespace chars",
			lexer:   getTestLexer("Hello World"),
			wantBuf: "",
			wantPos: position{line: 1, column: 1},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.lexer.consumeWhitespace()
			if tt.lexer.buffer != tt.wantBuf {
				t.Errorf("lexer.consumeWhitespace.buffer = %v, want %v", tt.lexer.buffer, tt.wantBuf)
			}

			if tt.lexer.pos != tt.wantPos {
				t.Errorf("lexer.consumedWhitespace.pos = %v, want %v", tt.lexer.pos, tt.wantPos)
			}
		})
	}
}

func TestLexer_consumeUntil(t *testing.T) {
	t.Parallel()

	type args struct {
		except []rune
	}

	tests := []struct {
		name         string
		lexer        *Lexer
		args         args
		wantConsumed bool
		wantBuf      string
		wantPos      position
	}{
		{
			name:  "consumes until it hits exception and returns true",
			lexer: getTestLexer("Hello+World"),
			args: args{
				except: []rune{'\n', '+'},
			},
			wantConsumed: true,
			wantBuf:      "Hello",
			wantPos:      position{line: 1, column: 6},
		},
		{
			name:  "does not consume and returns false if exception hit before first consume",
			lexer: getTestLexer("Hello World"),
			args: args{
				except: []rune{'H'},
			},
			wantConsumed: false,
			wantBuf:      "",
			wantPos:      position{line: 1, column: 1},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if gotConsumed := tt.lexer.consumeUntil(tt.args.except...); gotConsumed != tt.wantConsumed {
				t.Errorf("Lexer.consumeUntil(%v) = %v, want %v", tt.args.except, gotConsumed, tt.wantConsumed)
			}

			if tt.lexer.buffer != tt.wantBuf {
				t.Errorf("lexer.consumedUntil(%v).buffer = %v, want %v", tt.args.except, tt.lexer.buffer, tt.wantBuf)
			}

			if tt.lexer.pos != tt.wantPos {
				t.Errorf("lexer.consumedUntil(%v).pos = %v, want %v", tt.args.except, tt.lexer.pos, tt.wantPos)
			}
		})
	}
}

