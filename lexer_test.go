package meval

import (
	"fmt"
	"io"
	"testing"

	. "gopkg.in/check.v1"
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

type ValueAndError struct {
	value, error string
}

func (s *LexSuite) TestReportBadNumberSyntax(c *C) {
	tests := []ValueAndError{
		{"+0xA234", "+0x"},
		{"-0b0", "-0b"},
		{"0B02", "0B02"},
		{"0xAbE", "0xAb"},
	}
	for _, t := range tests {
		l := NewLexer(t.value)
		_, err := l.Next()
		if c.Check(err, Not(IsNil)) == false {
			continue
		}
		c.Check(err.Error(), Equals, fmt.Sprintf("Bad number syntax %q", t.error))
	}
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

func (s *LexSuite) TestReportUnknownToken(c *C) {
	l := NewLexer("@")
	_, err := l.Next()
	c.Assert(err, Not(IsNil))
	c.Check(err.Error(), Equals, "Got unexpected rune @")
}

func (s *LexSuite) TestShouldForbidInvalidOperatorToken(c *C) {
	err := registerOpToken("aa", tokUserStart)
	c.Assert(err, Not(IsNil))
	c.Check(err.Error(), Equals, "Invalid operator syntax \"aa\"")

	hasPanic := false
	defer func() {
		if r := recover(); r != nil {
			hasPanic = true
			c.Check(r, Equals, "Cannot register operator token : Invalid operator syntax \"aa\"")
		}

		c.Check(hasPanic, Equals, true)

	}()

	mustRegisterOpToken("aa", tokUserStart)
}

func (s *LexSuite) TestReportsInvalidTokenOperator(c *C) {
	// Adds a special token for @@
	err := registerOpToken("@@", tokUserStart)
	c.Assert(err, IsNil, Commentf("Cannot register valid token @@ : %s", err))

	l := NewLexer("@@ @@@@ @")

	t, err := l.Next()
	c.Assert(err, IsNil)
	c.Check(t.Type, Equals, tokUserStart)

	t, err = l.Next()
	c.Assert(err, IsNil)
	c.Check(t.Type, Equals, tokUserStart)

	t, err = l.Next()
	c.Assert(err, IsNil)
	c.Check(t.Type, Equals, tokUserStart)

	t, err = l.Next()
	c.Assert(err, Not(IsNil))
	c.Check(err.Error(), Equals, "Invalid token \"@\" found")

}
