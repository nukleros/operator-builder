// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package lexer

import "unicode"

// consume consumes as many runes as there are in the given string.
func (l *Lexer) consume(s string) {
	for range s {
		l.next()
	}
}

// consumed checks the input to see if it starts with the given token and does
// not start with any of the given exceptions. If so, it consumes the given
// token and returns true. Otherwise, it returns false.
func (l *Lexer) consumed(token string, except ...string) bool {
	if l.hasPrefix(token) {
		for _, e := range except {
			if l.hasPrefix(token + e) {
				return false
			}
		}

		l.consume(token)

		return true
	}

	return false
}

// consumedWhitespaces checks the input to see if, after whitespace is removed, it
// starts with one of the given tokens. If so, it consumes that
// token and any whitespace and returns true. Otherwise, it returns false.
func (l *Lexer) consumedWhitespaced(tokens ...string) bool {
	if l.peekedWhitespaced(tokens...) {
		r := make([]rune, l.width)

		l.consume(string(r))

		return true
	}

	return false
}

// consumeWhitespace consumes any leading whitespace.
func (l *Lexer) consumeWhitespace() {
	// consume whitespace
	for {
		r := l.next()

		if !unicode.IsSpace(r) {
			l.backup()

			break
		}
	}
}

// consumeUntil consumes tokens until it hits one of the exceptions provided,
// if no exception is provided it will consume tokens until end of input.
func (l *Lexer) consumeUntil(except ...rune) (consumed bool) {
	except = append(except, eof)

	for {
		le := l.next()
		for _, s := range except {
			if le == s {
				l.backup()

				return consumed
			}
		}

		consumed = true
	}
}
