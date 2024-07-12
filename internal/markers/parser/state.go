// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"errors"
	"strconv"

	"github.com/nukleros/operator-builder/internal/markers/lexer"
)

func startParse(p *Parser) stateFn {
	switch {
	case p.peeked(lexer.LexemeComment):
		p.discard()

		return parse
	case p.consumed(lexer.LexemeMarkerStart):
		return parseMarkerStart
	case p.consumed(lexer.LexemeEOF):
		return nil
	default:
		return parse
	}
}

func parse(p *Parser) stateFn {
	switch {
	case p.peeked(lexer.LexemeComment):
		p.discard()

		return parse
	case p.consumed(lexer.LexemeMarkerStart):
		return parseMarkerStart
	case p.consumed(lexer.LexemeEOF):
		return nil
	case p.consumed(lexer.LexemeError):
		return p.error(errors.New(p.currentLexeme.Value)) //nolint:goerr113
	default:
		p.next()
		p.scopeBuffer = ""

		return parse
	}
}

func parseMarkerStart(p *Parser) stateFn {
	if p.consumed(lexer.LexemeScope) {
		return parseScope
	}

	return parse
}

func parseScope(p *Parser) stateFn {
	if p.consumed(lexer.LexemeSeparator) {
		return parseSeparator
	}

	return parse
}

func parseSeparator(p *Parser) stateFn {
	switch {
	case p.consumed(lexer.LexemeScope):
		return parseScope
	case p.peeked(lexer.LexemeArg):
		if found := p.loadDefinition(); found {
			return parseArg
		}
	}

	p.flush()

	return parse
}

func parseArg(p *Parser) stateFn {
	if p.consumed(lexer.LexemeArg) {
		if found := p.currentDefinition.LookupArgument(p.currentLexeme.Value); found {
			argName := p.currentLexeme.Value

			if p.peeked(lexer.LexemeArgAssignment) {
				p.next()
			}

			return parseArgValue(p, argName)
		}
	}

	return parse
}

func parseArgValue(p *Parser, argName string) stateFn {
	stripQuotes(p)

	switch {
	case p.peeked(lexer.LexemeSyntheticBoolLiteral):
		lx := p.peek()

		b, err := strconv.ParseBool(lx.Value)
		if err != nil {
			return p.error(err)
		}

		if err := p.currentDefinition.SetArgument(argName, b); err != nil {
			return p.error(err)
		}

		p.discard()
	case p.consumed(lexer.LexemeBoolLiteral):
		b, err := strconv.ParseBool(p.currentLexeme.Value)
		if err != nil {
			return p.error(err)
		}

		if err := p.currentDefinition.SetArgument(argName, b); err != nil {
			return p.error(err)
		}
	case p.consumed(lexer.LexemeIntegerLiteral):
		v, err := strconv.Atoi(p.currentLexeme.Value)
		if err != nil {
			return p.error(err)
		}

		if err := p.currentDefinition.SetArgument(argName, v); err != nil {
			return p.error(err)
		}
	case p.consumed(lexer.LexemeFloatLiteral):
		const floatSize = 32

		v, err := strconv.ParseFloat(p.currentLexeme.Value, floatSize)
		if err != nil {
			return p.error(err)
		}

		if err := p.currentDefinition.SetArgument(argName, v); err != nil {
			return p.error(err)
		}
	case p.consumed(lexer.LexemeStringLiteral):
		if err := p.currentDefinition.SetArgument(argName, p.currentLexeme.Value); err != nil {
			return p.error(err)
		}

		stripQuotes(p)

	default:
		return parse
	}

	return parseMoreArgs
}

func parseMoreArgs(p *Parser) stateFn {
	switch {
	case p.consumed(lexer.LexemeArgDelimiter):
		return parseArg
	case p.consumed(lexer.LexemeMarkerEnd):
		err := p.emit()
		if err != nil {
			return p.error(err)
		}

		return parse
	default:
		return parse
	}
}

func stripQuotes(p *Parser) {
	if p.peeked(lexer.LexemeQuote) {
		p.next()
	}
}
