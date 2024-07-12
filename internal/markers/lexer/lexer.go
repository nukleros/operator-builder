// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package lexer

import (
	"bufio"
	"io"
)

// Lexer holds the state of the scanner.
type Lexer struct {
	name              string // name of Lexer used for error messages
	buffer            string
	start             position    // start position of this lexeme
	pos               position    // current position in input
	lineLens          map[int]int // length of each line
	width             int         // width of last rune read from input
	state             stateFn     // lexer state
	stack             []stateFn   // lexer stack
	items             chan Lexeme // channel of scanned Lexemes
	lastEmittedLexeme Lexeme      // type of last emitted Lexeme (or lexemEOF if no Lexeme has been emitted)
	reader            *bufio.Reader
}

// NewLexer creates a new lexer for the input reader.
func NewLexer(r io.Reader) *Lexer {
	const bufferSize = 3

	return &Lexer{
		name:     "Marker Lexer",
		start:    position{line: 1, column: 1},
		pos:      position{line: 1, column: 1},
		lineLens: make(map[int]int),
		state:    lex,
		stack:    make([]stateFn, 0),
		items:    make(chan Lexeme, bufferSize),
		reader:   bufio.NewReader(r),
	}
}

// Run runs the lexer until eof or until it encounters an error.
func (l *Lexer) Run() {
	for l.state != nil {
		l.state = l.state(l)
	}
	close(l.items)
}

// NextLexeme returns the next item from the input.
func (l *Lexer) NextLexeme() Lexeme {
	return <-l.items
}
