package meval

import (
	. "gopkg.in/check.v1"
	"io"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type LexSuite struct{}

var _ = Suite(&LexSuite{})

func CheckAllToken(l *Lexer, tokens []Token, c *C) {
	for i, t := range tokens {
		lexed, err := l.Next()
		c.Assert(err, IsNil, Commentf("[%d:Check '%s']: got error: %s", i, t.Value, err))
		c.Check(lexed, Equals, t)
	}

	_, err := l.Next()
	c.Assert(err, Equals, io.EOF)
}

func (s *LexSuite) TestLexNumber(c *C) {
	tokens := []Token{
		NewToken(tokValue, "2455"),
		NewToken(tokValue, "-46"),
		NewToken(tokValue, "+89e67"),
		NewToken(tokValue, "+89.46E67"),
		NewToken(tokValue, "35i"),
		NewToken(tokValue, "35.46e-67"),
		NewToken(tokValue, "35.46e+067"),
		NewToken(tokValue, "0b111001"),
		NewToken(tokValue, "0xabcdef46"),
	}

	toLex := ""
	for i, t := range tokens {
		if i > 0 {
			toLex += " "
		}
		toLex += t.Value
	}

	CheckAllToken(NewLexer(toLex), tokens, c)
}

func (s *LexSuite) TestComplexLex(c *C) {
	toLex := " 45 + - */ +44 -36 (,)foo sqrt()-56e23^"

	tokens := []Token{
		NewToken(tokValue, "45"),
		NewToken(tokPlus, "+"),
		NewToken(tokMinus, "-"),
		NewToken(tokMult, "*"),
		NewToken(tokDivide, "/"),
		NewToken(tokValue, "+44"),
		NewToken(tokValue, "-36"),
		NewToken(tokOParen, "("),
		NewToken(tokComma, ","),
		NewToken(tokCParen, ")"),
		NewToken(tokIdent, "foo"),
		NewToken(tokIdent, "sqrt"),
		NewToken(tokOParen, "("),
		NewToken(tokCParen, ")"),
		NewToken(tokValue, "-56e23"),
		NewToken(tokPower, "^"),
	}

	CheckAllToken(NewLexer(toLex), tokens, c)
}
