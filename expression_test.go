package meval

import (
	"fmt"
	. "gopkg.in/check.v1"
	"math"
)

type ExprSuite struct {
	c *MapContext
}

var _ = Suite(&ExprSuite{
	c: NewMapContext(),
})

type ExpResult struct {
	Result float64
	Input  string
}

func (e ExpResult) String() string {
	return fmt.Sprintf(" %s == %f", e.Input, e.Result)
}

func (s *ExprSuite) SetUpSuite(c *C) {
	err := s.c.CompileAndAdd("foo", "3.0")
	c.Assert(err, IsNil, Commentf("Got error at SetUp: %s", err))
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
		res, err := ee.Eval(s.c)
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
	res, err := exp.Eval(s.c)
	c.Assert(err, Not(IsNil))
	c.Check(err.Error(),Equals,"Could not find 'does' in MapContext")
	c.Check(math.IsNaN(res), Equals, true)

	res, err =  exp.Eval(nil)
	c.Assert(err,Not(IsNil))
	c.Check(err.Error(),Equals,"'does' referenced, but no Context providen")
	c.Check(math.IsNaN(res), Equals, true)
}

func (s *ExprSuite) TestCyclicEvaluationNotPermitted(c *C) {
	//create a cycle of references
	s.c.CompileAndAdd("bar", "sqrt(baz)")
	s.c.CompileAndAdd("baz", "foobar * foobar")
	s.c.CompileAndAdd("foobar", "bar ^ 2.0")
	//we should clean in any case
	defer func() {
		s.c.Delete("bar")
		s.c.Delete("baz")
		s.c.Delete("foobar")
	}()

	exp, err := Compile("bar + 42.0")
	c.Assert(err, IsNil, Commentf("Got compilation error %s", err))
	res, err := exp.Eval(s.c)
	c.Assert(err, Not(IsNil))
	c.Check(err.Error(), Equals, "Got cyclic dependency bar -> baz -> foobar -> bar")
	c.Check(math.IsNaN(res), Equals, true)

}

func ExampleExpression_basic() {
	expr, err := Compile("1.0 + 2.0")
	if err != nil {
		fmt.Printf("Got error: %s", err)
		return
	}

	res, err := expr.Eval(nil)
	if err != nil {
		fmt.Printf("Got error: %s", err)
		return
	}
	fmt.Printf("%f", res)
	//Output: 3.000000
}
