// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package lexer

import (
	"strconv"
	"unicode"
)

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*Lexer) stateFn

// lex scans until a comment or makerStart Lexeme is found.
func lex(l *Lexer) stateFn {
	l.stripWhitespace()

	switch {
	case l.isEmpty():
		if !l.emptyStack() {
			return l.pop()
		}

		l.emitSynthetic(LexemeEOF, "")

		return nil
	case l.consumedWhitespaced(golangComment, yamlComment):
		return lexCommentStart
	case l.consumed(markerStart):
		return lexMarkerStart
	default:
		l.discard()

		return lex
	}
}

// lexCommentStart scans a comment delimiter. the Left comment Lexeme is known to be present.
func lexCommentStart(l *Lexer) stateFn {
	l.emit(LexemeComment)

	return lexComment
}

// lexComment scans a comment.
func lexComment(l *Lexer) stateFn {
	switch {
	case l.consumed(markerStart):
		return lexMarkerStart
	case l.peeked("\n"), l.isEmpty():
		return lex
	default:
		l.discard()

		return lexComment
	}
}

// lexMarkerStart scans a markerStart Lexeme. the markerStart Lexeme is known to be present.
func lexMarkerStart(l *Lexer) stateFn {
	if unicode.IsLetter(l.peek()) {
		l.emit(LexemeMarkerStart)

		return lexMarker
	}

	return lexComment
}

// lexMarker scans a makerScope. the markerStart Lexeme is known to be present.
func lexMarker(l *Lexer) stateFn {
	exceptions := []rune{
		':', '=', ' ', '"', '\'', '`',
		',', '+', '{', '}', '[', ']',
		'(', ')', ';', '\n', eof,
	}

	if markerScope := l.consumeUntil(exceptions...); !markerScope {
		l.backup()

		l.flush()

		return lexComment
	}

	switch {
	case l.peeked(markerSeparator):
		l.emit(LexemeScope)
		l.consume(markerSeparator)
		l.emit(LexemeSeparator)

		return lexMarker
	case l.peeked(" "), l.peeked("\n"), l.peek() == eof:
		if l.lastEmittedLexeme.Type != LexemeSeparator {
			return l.warningf(`marker without scope found`)
		}

		l.emit(LexemeArg)
		l.emitSynthetic(LexemeSyntheticBoolLiteral, "true")
		l.emitSynthetic(LexemeMarkerEnd, "\n")

		return lexComment
	case l.peeked(argAssignment):
		if l.lastEmittedLexeme.Type != LexemeSeparator {
			return l.warningf(`marker without scope found`)
		}

		l.emit(LexemeArg)
		l.consume(argAssignment)
		l.emit(LexemeArgAssignment)

		return lexArgValueInitial
	default:
		return l.warningf("invalid marker found")
	}
}

func lexArgs(l *Lexer) stateFn {
	exceptions := []rune{
		':', '=', ' ', '"', '\'', '`',
		',', '+', '{', '}', '[', ']',
		'(', ')', ';', '\n', eof,
	}

	if argName := l.consumeUntil(exceptions...); !argName {
		l.backup()

		l.flush()

		l.emitSynthetic(LexemeMarkerEnd, "\n")

		return lex
	}

	l.emit(LexemeArg)

	switch {
	case l.consumed(argAssignment):
		l.emit(LexemeArgAssignment)

		return lexArgValueInitial
	case l.peeked(" "), l.peeked("\n"), l.peek() == eof:
		l.emitSynthetic(LexemeSyntheticBoolLiteral, "true")
		l.emitSynthetic(LexemeMarkerEnd, "\n")

		return lexComment
	case l.peeked(argDelimiter):
		l.emitSynthetic(LexemeSyntheticBoolLiteral, "true")

		return lexMoreArgs
	default:
		return l.errorf("malformed argument: %s", l.buffer)
	}
}

func lexArgValueInitial(l *Lexer) stateFn {
	if nextState, present := lexStringLiteral(l, lexMoreArgs); present {
		return nextState
	}

	if nextState, present := lexNumericLiteral(l, lexMoreArgs); present {
		return nextState
	}

	if nextState, present := lexBooleanLiteral(l, lexMoreArgs); present {
		return nextState
	}

	if nextState, present := lexNakedStringLiteral(l, lexMoreArgs); present {
		return nextState
	}

	return l.errorf("malformed argument: %s", l.buffer)
}

func lexStringLiteral(l *Lexer, nextState stateFn) (stateFn, bool) {
	var quote string

	switch string(l.peek()) {
	case singleQuote:
		quote = singleQuote
	case doubleQuote:
		quote = doubleQuote
	case literalQuote:
		quote = literalQuote
	default:
		return nil, false
	}

	l.consume(quote)
	l.emit(LexemeQuote)

	pos := l.pos
	context := l.context()

	for {
		switch {
		case l.peek() == eof:
			return l.rawErrorf(`unmatched string delimiter %s at position %+v, following %q`, quote, pos, context), true
		case l.peeked("\n"):
			if quote == literalQuote {
				l.next()

				if l.peekedWhitespaced(golangComment, yamlComment) {
					l.discardUntil(golangComment, yamlComment)
					l.discard()
				}
			} else {
				return l.rawErrorf(`unmatched string delimiter %s at position %+v, following %q`, quote, pos, context), true
			}
		case l.peeked(quote):
			l.emit(LexemeStringLiteral)
			l.consume(quote)
			l.emit(LexemeQuote)

			return nextState, true
		default:
			l.next()
		}
	}
}

func lexNumericLiteral(l *Lexer, nextState stateFn) (stateFn, bool) {
	n := l.peek()
	if l.peekedOneOf('.', '-') || unicode.IsNumber(l.peek()) {
		float := n == '.'

		for {
			l.next()

			if l.peekedOneOf('.', 'e', 'E', '-') {
				float = true

				continue
			}

			if !unicode.IsNumber(l.peek()) {
				break
			}
		}

		l.push(nextState)

		if float {
			return lexFloatLiteral, true
		}

		return lexIntegerLiteral, true
	}

	return nil, false
}

func lexFloatLiteral(l *Lexer) stateFn {
	// validate float
	const floatBitSize = 64

	if _, err := strconv.ParseFloat(l.value(), floatBitSize); err != nil {
		return l.rawErrorf("invalid float literal %q: %s before position %d", l.value(), err, l.pos)
	}

	l.emit(LexemeFloatLiteral)

	return l.pop()
}

func lexIntegerLiteral(l *Lexer) stateFn {
	// validate integer
	if _, err := strconv.Atoi(l.value()); err != nil {
		return l.rawErrorf("invalid integer literal %q: %s before position %d", l.value(), err, l.pos)
	}

	l.emit(LexemeIntegerLiteral)

	return l.pop()
}

func lexBooleanLiteral(l *Lexer, nextState stateFn) (stateFn, bool) {
	if l.consumedWhitespaced("true") || l.consumedWhitespaced("false") {
		l.emit(LexemeBoolLiteral)

		return nextState, true
	}

	return nil, false
}

func lexNakedStringLiteral(l *Lexer, nextState stateFn) (stateFn, bool) {
	exceptions := []rune{
		':', '=', ' ', '"', '\'', '`',
		',', '+', '{', '}', '[', ']',
		'(', ')', '\n', eof,
	}

	if argValue := l.consumeUntil(exceptions...); !argValue {
		return nil, false
	}

	l.emit(LexemeStringLiteral)

	return nextState, true
}

func lexMoreArgs(l *Lexer) stateFn {
	switch {
	case l.consumed(argDelimiter):
		l.emit(LexemeArgDelimiter)

		return lexArgs
	case l.peeked(" "), l.peeked("\n"), l.peek() == eof:
		l.emitSynthetic(LexemeMarkerEnd, "\n")

		return lexComment
	default:
		return l.errorf("malformed argument: %s", l.buffer)
	}
}
