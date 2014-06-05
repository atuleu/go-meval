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
		{4.0, "1.0 + 2.0 + 1.0"},
		{1.12345, "1.12345"},
		{11.0, "1.0 + 10.0"},
		{1 / 10.0, "1.0 / 10.0"},
		{30.0, "3.0 * 10.0"},
		{-2.0, "8.0 - 10.0"},
		{9.0, "3 * foo"},
		{27.0, "3 ^ 3 ^1"},
		{math.Pow(math.Cos(42.0)*3.14159+2, 2.45), "( cos(42.0) * 3.14159 + 2 ) ^2.45"},
		{0.0, "sin(0.0)"},
		{0.0, "asin(0.0)"},
		{1.0, "cos(0.0)"},
		{0.0, "acos(1.0)"},
		{0.0, "tan(0.0)"},
		{0.0, "atan(0.0)"},
		{0.0, "sqrt(0.0)"},
		{math.Exp(0.0), "exp(0.0)"},
		{0.0, "ln(1.0)"},
		{1.0, "log(10.0)"},
		{2.0, "ceil(1.5)"},
		{1.0, "floor(1.5)"},
		{math.Pi / 4, "atan2(1.0,1.0)"},
		{math.Pi / 4, "pi() / 4"},
	}

	for i, e := range exps {
		ee, err := Compile(e.Input)
		if c.Check(err, IsNil, Commentf("[%d: %s]: got error at compilation : %s", i, e, err)) == false {
			continue
		}
		res, err := ee.Eval(s.c)
		if c.Check(err, IsNil, Commentf("[%d: %s]: got error at evaluation: %s", i, e, err)) == false {
			continue
		}
		c.Check(res, Equals, e.Result)
	}
}

func (s *ExprSuite) TestUnfoundVariableEvaluation(c *C) {
	exp, err := Compile("1.0 * does * not + exist")
	c.Assert(err, IsNil)
	res, err := exp.Eval(s.c)
	c.Assert(err, Not(IsNil))
	c.Check(err.Error(), Equals, "Could not find 'does' in MapContext")
	c.Check(math.IsNaN(res), Equals, true)

	res, err = exp.Eval(nil)
	c.Assert(err, Not(IsNil))
	c.Check(err.Error(), Equals, "'does' referenced, but no Context providen")
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

type CompileError struct {
	input, error string
}

func (s *ExprSuite) TestCompilationError(c *C) {
	registerRuneToken('%', TokPlus+100)
	tests := []CompileError{
		{"( 2.0 ))", "Mismatched parenthese in ( 2.0 ))"},
		{"(( foo )", "Mismatched parenthese in (( foo )"},
		{"( +0x ))", "Bad number syntax \"+0x\""},
		{"( +0x ))", "Bad number syntax \"+0x\""},
		{"sin(0.0),", "Misplaced comma or mismatched parenthese in sin(0.0),"},
		{"atan2(0.0 +,2.1)", "Evaluation stack error for '+', need 2 element, but only 1 provided"},
		{"sin(0.0 , 0.3)", "Evaluation stack error, still got 2 element instead of 1 at the final state"},
		{"5 % 3", "Operator '%' is not yet implemented"},
		{"5 + ", "Evaluation stack error for '+', need 2 element, but only 1 provided"},
		{"sin()", "Evaluation stack error for 'sin()', need 1 element, but only 0 provided"},
		{"(5 + )", "Evaluation stack error for '+', need 2 element, but only 1 provided"},
	}

	for i, t := range tests {
		_, err := Compile(t.input)
		if c.Check(err, Not(IsNil)) == false {
			continue
		}
		c.Check(err.Error(), Equals, t.error, Commentf("[%d] : %s]", i, t.input))
	}
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

type FunctionAndResult struct {
	input  string
	result float64
}
