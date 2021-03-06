package meval

import (
	"fmt"
	"math"
	"strings"
)

// Expression can be evaluated to float from a Context
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
	if c == nil {
		return math.NaN(), fmt.Errorf("'%s' referenced, but no Context providen",
			e.variable)
	}

	if bad, deps := c.testStack(e); bad == true {
		deps = append([]string{deps[len(deps)-1]},
			deps...)
		return math.NaN(), fmt.Errorf("Got cyclic dependency %s", strings.Join(deps, " -> "))
	}
	c.push(e)
	defer c.pop()
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

// NEvaluer is a function that takes a list of float and returns a
// float
type NEvaluer func([]float64) float64

type nExp struct {
	children []Expression
	card     int
	evaluer  NEvaluer
}

func (e *nExp) Eval(c Context) (float64, error) {
	if len(e.children) != e.card {
		return math.NaN(), fmt.Errorf("Bad AST, expression expected %d children, got %d", e.card, len(e.children))
	}
	values := make([]float64, e.card)
	for i, exp := range e.children {
		var err error
		if values[i], err = exp.Eval(c); err != nil {
			return math.NaN(), err
		}
	}
	return e.evaluer(values), nil
}
