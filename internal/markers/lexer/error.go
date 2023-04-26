// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package lexer

import "fmt"

// context returns the last emitted Lexeme (if any) followed by the portion
// of the current Lexeme scanned so far.
func (l *Lexer) context() string {
	return l.lastEmittedLexeme.Value + l.buffer
}

// errorf returns an error Lexeme with context and terminates the scan.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Lexeme{
		Type:  LexemeError,
		Value: fmt.Sprintf("%s at position: %+v, following %q", fmt.Sprintf(format, args...), l.pos, l.context()),
		Pos:   l.pos,
	}

	return nil
}

// rawErrorf returns an error Lexeme with no context and terminates the scan.
func (l *Lexer) rawErrorf(format string, args ...interface{}) stateFn {
	l.items <- Lexeme{
		Type:  LexemeError,
		Value: fmt.Sprintf(format, args...),
		Pos:   l.pos,
	}

	return nil
}

// warningf returns an warning Lexeme with context and continues the scan.
func (l *Lexer) warningf(format string, args ...interface{}) stateFn {
	l.items <- Lexeme{
		Type:  LexemeWarning,
		Value: fmt.Sprintf("%s at position: %+v, following %q", fmt.Sprintf(format, args...), l.pos, l.context()),
		Pos:   l.pos,
	}

	return lexComment
}
