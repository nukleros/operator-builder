// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package parser

// next returns the next token.
func (p *Parser) next() {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		for i := 2; i > 0; i-- {
			p.peekStack[i] = p.peekStack[i-1]
		}

		p.peekStack[0] = p.lexer.NextLexeme()
	}

	p.scopeBuffer += p.peekStack[p.peekCount].String()
	p.currentLexeme = p.peekStack[p.peekCount]
}

// discard discards the next token without consuming it.
func (p *Parser) discard() {
	if p.peekCount > 1 {
		for i := p.peekCount; i > 0; i-- {
			p.peekStack[i] = p.peekStack[i-1]
		}
	}

	p.peekStack[0] = p.lexer.NextLexeme()
}

func (p *Parser) flush() {
	p.scopeBuffer = ""
	p.currentDefinition = nil
}
