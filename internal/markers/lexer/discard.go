// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package lexer

import (
	"unicode"
	"unicode/utf8"
)

// discard discards a single rune from the reader without reading into the buffer.
func (l *Lexer) discard() {
	l.discardN(1)
}

// discardN discards N runes from the reader without reading into the buffer.
func (l *Lexer) discardN(n int) {
	rs := l.peekN(n)

	for _, r := range rs {
		if r == eof {
			l.flush()

			return
		}

		w := utf8.RuneLen(r)
		w, _ = l.reader.Discard(w)

		l.pos.column += w

		if r == '\n' {
			l.resetPosition()
		}
	}

	l.start = l.pos
}

// discardUntil discards runes without reading into the buffer until it reaches one of the given token.
func (l *Lexer) discardUntil(tokens ...string) {
	for {
		for _, token := range tokens {
			if l.hasPrefix(token) {
				return
			}
		}

		l.discard()
	}
}

// stripWhitespace strips out whitespace
// it should only be called immediately after emitting a Lexeme.
func (l *Lexer) stripWhitespace() {
	// find whitespace
	for {
		nextRune := l.peek()
		if !unicode.IsSpace(nextRune) {
			break
		}

		l.discard()
	}
}

// flush clears the current buffer.
func (l *Lexer) flush() {
	l.buffer = ""
	l.start = l.pos
}
