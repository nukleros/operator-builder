// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package parser

import "github.com/nukleros/operator-builder/internal/markers/lexer"

func (p *Parser) consumed(lxt lexer.LexemeType) bool {
	if p.peek().Type == lxt {
		p.next()

		return true
	}

	return false
}
