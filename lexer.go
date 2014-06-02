package meval

import (
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

type TokenType int

const (
	tokPlus TokenType = iota
	tokMinus
	tokMult
	tokDivide
	tokIdent
	tokOParen
	tokCParen
	tokValue
	tokComma
	tokPower
)

type Token struct {
	Type  TokenType
	Value string
}

func NewToken(t TokenType, value string) Token {
	return Token{Type: t, Value: value}
}

type Lexer struct {
	input      string
	tokens     chan Token
	errors     chan error
	action     lActionFn
	start, pos int
	width      int
}

type lActionFn func(l *Lexer) lActionFn

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		start:  0,
		pos:    0,
		width:  0,
		errors: make(chan error, 2),
		tokens: make(chan Token, 2),
		action: lexWS,
	}
}

func (l *Lexer) Next() (Token, error) {
	for {
		select {
		case err := <-l.errors:
			if err != io.EOF {
				return Token{}, err
			}
		case t := <-l.tokens:
			return t, nil
		default:
			if l.action == nil {
				return Token{}, io.EOF
			}
			l.action = l.action(l)
		}
	}
	panic("Should not be reached")
}

const eof rune = 0

// Actions

func lexHexadecimal(l *Lexer) lActionFn {
	l.acceptRun(numeric)
	if l.accept("abcedf") {
		l.acceptRun(numeric + "abcdef")
	} else if l.accept("ABCDEF") {
		l.acceptRun(numeric + "ABCDEF")
	}
	return lexNumberEndCheck
}

func lexBinary(l *Lexer) lActionFn {
	l.acceptRun("01")
	return lexNumberEndCheck
}

func lexNumberEndCheck(l *Lexer) lActionFn {
	if l.accept(alphabetic + numeric) {
		return l.errorf("Bad number syntax %q", l.current())
	}

	l.emit(tokValue)

	return lexWS
}

func lexNumber(l *Lexer) lActionFn {
	l.backup()
	var asPM = false
	if l.accept("+-") {
		asPM = true
		if l.accept(numeric) == false {
			if l.current() == "+" {
				l.emit(tokPlus)
			} else {
				l.emit(tokMinus)
			}
			return lexWS
		}
	}

	if l.accept("0") {
		if l.accept("xX") {
			if asPM == true {
				return l.errorf("Bad number syntax %q", l.current())
			}
			return lexHexadecimal
		}

		if l.accept("bB") {
			if asPM == true {
				return l.errorf("Bad number syntax %q", l.current())
			}
			return lexBinary
		}
	}

	l.acceptRun(numeric)

	if l.accept(".") {
		l.acceptRun(numeric)
	}

	if l.accept("eE") {
		l.accept("+-")
	}

	l.acceptRun(numeric)

	l.accept("i")

	return lexNumberEndCheck
}

func lexIdentifier(l *Lexer) lActionFn {
	l.acceptRun(alphabetic + numeric + "_")
	l.emit(tokIdent)
	return lexWS
}

func lexWS(l *Lexer) lActionFn {
	var ru rune
	for {
		ru = l.next()
		if unicode.IsSpace(ru) == false {
			break
		}
	}
	//we peek the last char
	l.backup()

	//we ignore all data
	l.ignore()

	if ru == eof {
		return nil
	}

	if l.accept(numeric + "+-") {
		return lexNumber
	}

	if l.accept(alphabetic + "_") {
		return lexIdentifier
	}

	//check for runes
	ru = l.next() //we know it is not eof
	if action, ok := runeToken[ru]; ok == true {
		return action
	}

	return l.errorf("Got unexpected run %c", ru)
}

// static data

var numeric = "0123456789"

var alphabetic = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var runeToken = make(map[rune]lActionFn)

func registerRuneToken(ru rune, t TokenType) {
	runeToken[ru] = func(l *Lexer) lActionFn {
		l.emit(t)
		return lexWS
	}
}

func init() {
	registerRuneToken('+', tokPlus)
	registerRuneToken('-', tokMinus)
	registerRuneToken('*', tokMult)
	registerRuneToken('^', tokPower)
	registerRuneToken('/', tokDivide)
	registerRuneToken('(', tokOParen)
	registerRuneToken(')', tokCParen)
	registerRuneToken(',', tokComma)
}

// helpers

func (l *Lexer) current() string {
	return l.input[l.start:l.pos]
}

func (l *Lexer) emit(t TokenType) {
	l.tokens <- NewToken(t, l.current())
	l.ignore()
}

func (l *Lexer) error(err error) lActionFn {
	if len(l.errors) == 0 {
		l.errors <- err
	}
	return nil
}

func (l *Lexer) errorf(format string, args ...interface{}) lActionFn {
	return l.error(fmt.Errorf(format, args...))
}

func (l *Lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	var ru rune
	ru, l.width =
		utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return ru
}

func (l *Lexer) backup() {
	l.pos -= l.width
}

func (l *Lexer) peek() rune {
	ru := l.next()
	l.backup()
	return ru
}

func (l *Lexer) ignore() {
	l.start = l.pos
}

func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}
