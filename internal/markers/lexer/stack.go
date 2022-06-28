// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package lexer

// push pushes a state function on the stack which will be resumed when parsing terminates.
func (l *Lexer) push(state stateFn) {
	l.stack = append(l.stack, state)
}

// pop pops a state function from the stack. If the stack is empty, returns an error function.
func (l *Lexer) pop() stateFn {
	if len(l.stack) == 0 {
		return l.errorf("syntax error")
	}

	index := len(l.stack) - 1
	element := l.stack[index]
	l.stack = l.stack[:index]

	return element
}

// empty returns true if and only if the stack of state functions is empty.
func (l *Lexer) emptyStack() bool {
	return len(l.stack) == 0
}

