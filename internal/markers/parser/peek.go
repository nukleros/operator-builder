// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package parser

import "github.com/nukleros/operator-builder/internal/markers/lexer"

func (p *Parser) peek() lexer.Lexeme {
	if p.peekCount > 0 {
		return p.peekStack[p.peekCount-1]
	}

	p.peekCount = 1

	for i := 2; i > 0; i-- {
		p.peekStack[i] = p.peekStack[i-1]
	}

	p.peekStack[0] = p.lexer.NextLexeme()

	return p.peekStack[0]
}

func (p *Parser) peeked(lxt lexer.LexemeType) bool {
	return p.peek().Type == lxt
}
