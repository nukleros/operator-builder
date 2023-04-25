// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package lexer

import (
	"errors"
	"io"
)

// position marks the position of the lexer in the file.
type position struct {
	line   int
	column int
}

// next returns the next rune in the input.
func (l *Lexer) next() (r rune) {
	var err error

	r, l.width, err = l.reader.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			l.width = 0

			return eof
		}
	}

	if r == '\n' {
		l.resetPosition()
	} else {
		l.pos.column += l.width
	}

	l.buffer += string(r)

	return r
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *Lexer) backup() {
	if l.width != 0 {
		l.pos.column -= l.width
		if l.buffer != "" {
			l.buffer = l.buffer[:len(l.buffer)-1]
		}
	}

	if l.pos.column == 0 && l.pos.line > 1 {
		l.pos.line--
		l.pos.column = l.lineLens[l.pos.line]
	}

	_ = l.reader.UnreadRune()
}

// resetPosition resets the position for a new line.
func (l *Lexer) resetPosition() {
	l.lineLens[l.pos.line] = l.pos.column

	l.pos.line++
	l.pos.column = 1
}

// isEmpty returns true if the next rune is eof.
func (l *Lexer) isEmpty() bool {
	return l.peek() == eof
}
