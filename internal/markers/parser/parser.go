// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package parser

import (
	"bytes"

	"github.com/nukleros/operator-builder/internal/markers/lexer"
)

type stateFn func(*Parser) stateFn

type Parser struct {
	name              string
	scopeBuffer       string
	lexer             *lexer.Lexer
	registry          Registry
	currentLexeme     lexer.Lexeme
	currentDefinition Definition
	peekCount         int
	peekStack         [3]lexer.Lexeme
	stack             []stateFn
	state             stateFn
	items             chan *Result
}

func NewParser(input string, registry Registry) *Parser {
	const bufferSize = 3

	p := &Parser{
		name:        "Marker Parser",
		scopeBuffer: "",
		registry:    registry,
		lexer:       lexer.NewLexer(bytes.NewBufferString(input)),
		currentLexeme: lexer.Lexeme{
			Type:  lexer.LexemeError,
			Value: "",
		},
		peekStack: [3]lexer.Lexeme{},
		peekCount: 0,
		stack:     make([]stateFn, 0),
		state:     startParse,
		items:     make(chan *Result, bufferSize),
	}

	go p.lexer.Run()

	return p
}

func (p *Parser) Run() {
	for p.state != nil {
		p.state = p.state(p)
	}
	close(p.items)
}

func (p *Parser) NextItem() *Result {
	return <-p.items
}

func (p *Parser) Parse() []*Result {
	var results []*Result

	go p.Run()

	for it := p.NextItem(); it != nil; it = p.NextItem() {
		results = append(results, it)
	}

	return results
}
