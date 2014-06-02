package meval

import (
	"fmt"
	. "gopkg.in/check.v1"
	"math"
)

type MockContext struct{}

func (c *MockContext) GetExpression(name string) (Expression, error) {
	if name == "foo" {
		return &valueExp{3.0}, nil
	}
	return nil, fmt.Errorf("%s is not present in context", name)
}

type ExprSuite struct{}

var _ = Suite(&ExprSuite{})

type ExpResult struct {
	Result float64
	Input  string
}

func (e ExpResult) String() string {
	return fmt.Sprintf(" %s == %f", e.Input, e.Result)
}

func (s *ExprSuite) TestBasicEval(c *C) {

	exps := []ExpResult{
		{0.0, "0.0"},
		{1.12345, "1.12345"},
		{11.0, "1.0 + 10.0"},
		{1 / 10.0, "1.0 / 10.0"},
		{30.0, "3.0 * 10.0"},
		{-2.0, "8.0 - 10.0"},
		{9.0, "3 * foo"},
		{math.Pow(math.Cos(42.0)*3.14159+2, 2.45), "( cos(42.0) * 3.14159 + 2 ) ^2.45"},
	}

	for i, e := range exps {
		ee, err := Compile(e.Input)
		c.Check(err, IsNil, Commentf("[%d: %s]: got error at compilation : %s", i, e, err))
		if err != nil {
			continue
		}
		res, err := ee.Eval(&MockContext{})
		c.Check(err, IsNil, Commentf("[%d: %s]: got error at evaluation: %s", i, e, err))
		if err != nil {
			continue
		}
		c.Check(res, Equals, e.Result)
	}
}

func (s *ExprSuite) TestUnfoundVariableEvaluation(c *C) {
	exp, err := Compile("does * not + exist")
	c.Assert(err, IsNil)
	res, err := exp.Eval(&MockContext{})
	c.Check(err, Not(IsNil))
	c.Check(math.IsNaN(res), Equals, true)

}
