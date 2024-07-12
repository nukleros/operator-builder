// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package lexer

// emit passes a Lexeme back to the client.
func (l *Lexer) emit(typ LexemeType) {
	lx := Lexeme{
		Type:  typ,
		Value: l.value(),
		Pos:   l.start,
	}

	l.lastEmittedLexeme = lx

	l.buffer = ""
	l.start = l.pos

	l.items <- lx
}

// emitSynthetic passes a Lexeme back to the client which wasn't encountered in the input.
// The lexing position is not modified.
func (l *Lexer) emitSynthetic(typ LexemeType, val string) {
	lx := Lexeme{
		Type:  typ,
		Value: val,
	}

	l.lastEmittedLexeme = lx

	l.items <- lx
}

// value returns the portion of the current Lexeme scanned so far.
func (l *Lexer) value() string {
	return l.buffer
}
