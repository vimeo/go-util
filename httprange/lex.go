package httprange

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// itemKind is what's emitted by our lexer
type itemKind int

//go:generate stringer -type=itemKind ${GOFILE}

const (
	itemError itemKind = iota
	itemEOF
	itemUnit
	itemStart
	itemEnd
	itemLength

	eof rune = 0
)

type stateFn func(*lexer) stateFn

type token struct {
	kind itemKind
	tok  string
}

func (t token) String() string {
	return fmt.Sprintf("token(%s, %q)", t.kind, t.tok)
}

func lexResponse(input string) *lexer {
	return &lexer{
		input: input,
		state: startResponse,
	}
}

func lexRequest(input string) *lexer {
	return &lexer{
		input: input,
		state: startRequest,
	}
}

// A lexer based on Rob Pike's talk about the text/template lexer.
type lexer struct {
	input string
	start int
	pos   int
	width int

	state stateFn
	item  token
}

// Step is the only thing the parser should need to call.
func (l *lexer) step() token {
	if l.state == nil {
		return token{itemEOF, ""}
	}
	l.state = l.state(l)
	return l.item
}

// The rest of the lexer methods are for the stateFns to call.

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) chomp() {
	l.next()
	l.ignore()
}

func (l *lexer) emit(k itemKind) {
	l.item = token{k, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) error(s string) stateFn {
	l.item = token{itemError, fmt.Sprintf("%s (got %q)", s, l.input[l.start:l.pos])}
	return nil
}

// These are the states the lexer can be in.
//
// Generic ranges have two parts: the "unit" and the "response"
//
// Byte ranges have those, but the response is defined to be a range and a length.
// One of the range or the length can be "*", and if the range is not, it has a
// start and end component. RFC7233 has the formal grammar.
//
//Here's a dot description if you want a visual:
/*
dot -Tpng <<EOF | page
	digraph rangelexer {
		startResponse -> byteStart [label="\" \""];
		byteStart -> byteLen [label="\"/\""];
		byteStart -> byteEnd [label="\"-\""];
		byteEnd -> byteLen [label="\"/\""];
		startResponse [label="\N\n[^ ]+"];
		byteLen [label="\N\n\\*|[0-9]+"];
		byteStart [label="\N\n\\*|[0-9]+"];
		byteEnd [label="\N\n\[0-9]+"];
	}
EOF
*/

func startResponse(l *lexer) stateFn {
outer:
	for {
		switch l.next() {
		case ' ':
			l.backup()
			break outer
		case eof:
			return l.error("wanted a space")
		}
	}
	l.emit(itemUnit)
	l.chomp()
	return byteStart
}

func byteStart(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == '*':
			l.emit(itemStart)
			return byteLen
		case unicode.IsNumber(r):
		case r == '-':
			l.backup()
			l.emit(itemStart)
			return byteEnd
		case r == '/':
			return l.error("wanted an int-like until '-'")
		default:
			return l.error("wanted an int-like")
		}
	}
}

func byteEnd(l *lexer) stateFn {
	if r := l.next(); r != '-' {
		return l.error("wanted a '-'")
	}
	l.ignore()
	for {
		switch r := l.next(); {
		case unicode.IsNumber(r):
		case r == '/':
			l.backup()
			if l.pos == l.start {
				return l.error("wanted an int-like")
			}
			l.emit(itemEnd)
			return byteLen
		case r == eof:
			return l.error("wanted an int-like until '/', hit eof")
		default:
			return l.error("wanted an int-like")
		}
	}
}

func byteLen(l *lexer) stateFn {
	if r := l.next(); r != '/' {
		return l.error("wanted a '/'")
	}
	l.ignore()
	for {
		switch r := l.next(); {
		case unicode.IsNumber(r):
		case r == '*':
			fallthrough
		case r == eof:
			if l.pos == l.start {
				return l.error("wanted int-like until eof")
			}
			l.emit(itemLength)
			return nil
		default:
			return l.error("wanted int-like until end of string")
		}
	}
}

// These are request states.
/*
dot -Tpng <<EOF | page
	digraph requestlexer {
		startRequest -> byteRangeSet [label="\"=\""];
		byteRangeSet -> byteSuffixRange [label="\"-\""];
		byteRangeSet -> firstByte [label="[^-]"];
		firstByte -> lastByte [label="\"-\""];
		lastByte -> byteRangeSet [label="\",\""];
		byteSuffixRange -> byteRangeSet [label="\",\""];
		byteRangeSet -> eof;
	}
EOF
*/

func startRequest(l *lexer) stateFn {
	for {
		switch l.next() {
		case '=':
			l.backup()
			l.emit(itemUnit)
			l.chomp()
			return byteRangeSet
		case eof:
			return l.error("wanted a '='")
		}
	}
}

func byteRangeSet(l *lexer) stateFn {
	switch l.peek() {
	case eof:
		l.emit(itemEOF)
		return nil
	case '-':
		return byteSuffixRange(l)
	}
	return firstByte(l)
}

func byteSuffixRange(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == '-':
		case unicode.IsNumber(r):
		case r == ',':
			l.backup()
			fallthrough
		case r == eof:
			l.emit(itemStart)
			return lastByte
		default:
			return l.error("wanted int-like until ',' or eof")
		}
	}
}

func firstByte(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case unicode.IsNumber(r):
		case r == '-':
			l.backup()
			l.emit(itemStart)
			l.chomp()
			return lastByte
		default:
			return l.error("wanted int-like until '-'")
		}
	}
}

func lastByte(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case unicode.IsNumber(r):
		case r == ',':
			l.backup()
			l.emit(itemEnd)
			l.chomp()
			return byteRangeSet
		case r == eof:
			l.emit(itemEnd)
			return byteRangeSet
		default:
			return l.error("wanted int-like until ',' or eof")
		}
	}
}
