// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// peek returns the next rune in the input but without consuming it.
// it is equivalent to calling next() followed by backup().
func (l *Lexer) peek() rune {
	return l.peekN(1)[0]
}

// peek returns the next N runes in the input but without consuming them.
// it is equivalent to calling next() N times followed by backup() N times.
func (l *Lexer) peekN(n int) []rune {
	const maxByteSize = 4 // unicode rune can be up to 4 bytes

	var rs []rune

	l.width = 0

	b, _ := l.reader.Peek(n * maxByteSize)
	for n > len(rs) {
		if len(b) == 0 {
			rs = append(rs, eof)

			return rs
		}

		r, w := utf8.DecodeRune(b)
		b = b[w:]

		l.width += w

		rs = append(rs, r)
	}

	return rs
}

// peeked checks the input to see if it starts with the given token and does
// not start with any of the given exceptions. If so, it returns true
// Otherwise, it returns false.
func (l *Lexer) peeked(token string, except ...string) bool {
	if l.hasPrefix(token) {
		for _, e := range except {
			if l.hasPrefix(token + e) {
				return false
			}
		}

		return true
	}

	return false
}

// peekedWhitespaced checks the input to see if, after whitespace is removed, it
// starts with one of the given tokens. If so, it returns true. Otherwise, it returns false.
func (l *Lexer) peekedWhitespaced(tokens ...string) bool {
	for _, token := range tokens {
		// skip past whitespace
		i := 0

		for ; ; i++ {
			r := l.peekN(i + 1)
			if r[i] == eof {
				return false
			}

			if !unicode.IsSpace(r[i]) {
				break
			}
		}

		peeked := l.peekN(i + len(token))[i:]

		if strings.HasPrefix(string(peeked), token) {
			return true
		}
	}

	return false
}

// hasPrefix checks to see if the input has a prefix of the given string.
func (l *Lexer) hasPrefix(p string) bool {
	r := l.peekN(len(p))

	return strings.HasPrefix(string(r), p)
}

// peekedOneOf checks the input to see if it starts with one of the given tokens.
// If so, it returns true otherwise it returns false.
func (l *Lexer) peekedOneOf(tokens ...rune) bool {
	for _, token := range tokens {
		if l.peeked(string(token)) {
			return true
		}
	}

	return false
}

