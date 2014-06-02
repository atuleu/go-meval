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
		NewToken(TokValue, "2455"),
		NewToken(TokValue, "-46"),
		NewToken(TokValue, "+89e67"),
		NewToken(TokValue, "+89.46E67"),
		NewToken(TokValue, "35i"),
		NewToken(TokValue, "35.46e-67"),
		NewToken(TokValue, "35.46e+067"),
		NewToken(TokValue, "0b111001"),
		NewToken(TokValue, "0xabcdef46"),
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
		NewToken(TokValue, "45"),
		NewToken(TokPlus, "+"),
		NewToken(TokMinus, "-"),
		NewToken(TokMult, "*"),
		NewToken(TokDivide, "/"),
		NewToken(TokValue, "+44"),
		NewToken(TokValue, "-36"),
		NewToken(TokOParen, "("),
		NewToken(TokComma, ","),
		NewToken(TokCParen, ")"),
		NewToken(TokIdent, "foo"),
		NewToken(TokIdent, "sqrt"),
		NewToken(TokOParen, "("),
		NewToken(TokCParen, ")"),
		NewToken(TokValue, "-56e23"),
		NewToken(TokPower, "^"),
	}

	CheckAllToken(NewLexer(toLex), tokens, c)
}
