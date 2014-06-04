package meval

import (
	"fmt"
	"math"
	"strings"
)

type callStack interface {
	push(e *refExp)
	pop()
	testStack(e *refExp) (bool,[]string)
}

// A Context is a kind of dictionnary of expression. You can pass it
// to Eval
type Context interface {
	// Returns an expression from a given name.
	GetExpression(string) (Expression, error)
	callStack
}

type CallStack struct {
	stack []*refExp
}

func (c *CallStack)push(e *refExp) {
	c.stack = append(c.stack,e)
}

func (c *CallStack)pop() {
	if len(c.stack) == 0 {
		panic("Should never happen")
	}
	c.stack = c.stack[0:len(c.stack)-1]
}

func (c *CallStack)testStack(e *refExp) (bool, []string) {
	res := false
	var deps []string = nil
	for _,ee := range c.stack {
		if e == ee {
			res = true
		}
		if res == true {
			deps = append(deps,ee.variable)
		}
	}
	return res,deps
}

// An expression can be evaluated
// 
// BUG(tuleu): the expression evaluation mechanism is so simple that
// call loop can be easily implemented !
type Expression interface {
	// Evaluates The expression. You can pass a context to refer other
	// expression.  You can also pass nil as a context.
	Eval(Context) (float64, error)
}

// Compile a new expression from an input string
func Compile(input string) (Expression, error) {
	return buildAST(input)
}

//rest of the stuff is pretty private
type refExp struct {
	variable string
}


// TODO: This should be resilient to cyclic call ... this is not the
// case right now
func (e *refExp) Eval(c Context) (float64, error) {
	if bad, deps := c.testStack(e); bad == true {
		deps = append([]string{deps[len(deps)-1]},
			deps...)
		return math.NaN(), fmt.Errorf("Got cyclic dependency %s",strings.Join(deps," -> "))
	}
	c.push(e)
	defer c.pop()
	if c == nil {
		return math.NaN(), fmt.Errorf("Variable %s referenced, but no Context providen",
			e.variable)
	}
	expr, err := c.GetExpression(e.variable)
	if err != nil {
		return math.NaN(), err
	}
	return expr.Eval(c)
}

type valueExp struct {
	value float64
}

func (e *valueExp) Eval(Context) (float64, error) {
	return e.value, nil
}

type unaryEvaluer func(float64) float64

type unaryExp struct {
	child   Expression
	evaluer unaryEvaluer
}

func (e *unaryExp) Eval(c Context) (float64, error) {
	value, err := e.child.Eval(c)
	if err != nil {
		return math.NaN(), err
	}
	return e.evaluer(value), nil
}

type binaryEvaluer func(float64, float64) float64

type binaryExp struct {
	leftChild, rightChild Expression
	evaluer               binaryEvaluer
}

func (e *binaryExp) Eval(c Context) (float64, error) {
	valueLeft, err := e.leftChild.Eval(c)
	if err != nil {
		return math.NaN(), err
	}

	valueRight, err := e.rightChild.Eval(c)
	if err != nil {
		return math.NaN(), err
	}
	return e.evaluer(valueLeft, valueRight), nil
}
