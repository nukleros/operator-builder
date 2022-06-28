// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package parser

import "github.com/vmware-tanzu-labs/operator-builder/internal/markers/lexer"

func (p *Parser) consumed(lxt lexer.LexemeType) bool {
	if p.peek().Type == lxt {
		p.next()

		return true
	}

	return false
}

