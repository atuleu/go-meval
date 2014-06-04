package meval

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"strconv"
)

type outQueue struct {
	q []Expression
}

func (o *outQueue) unsafePop() Expression {
	expr := o.q[len(o.q)-1]
	o.q = o.q[0 : len(o.q)-1]
	return expr
}

func (o *outQueue) push(e Expression) {
	o.q = append(o.q, e)
}

func (o *outQueue) size() int {
	return len(o.q)
}

type queuePoper func(*outQueue) Expression

type operatorType uint

const (
	opStandard operatorType = iota
	opFunction
	opLeftParenthesis
)

type operator struct {
	oType            operatorType
	precedence, card int
	leftAssociative  bool
	poper            queuePoper
}

type opStack struct {
	s []operator
}

func (o *opStack) unsafePop() operator {
	op := o.s[len(o.s)-1]
	o.s = o.s[0 : len(o.s)-1]
	return op
}

func (o *opStack) push(op operator) {
	o.s = append(o.s, op)
}

func (o *opStack) size() int {
	return len(o.s)
}

func (o *opStack) unsafeTop() operator {
	return o.s[len(o.s)-1]
}

type function struct {
	card    int
	evaluer NEvaluer
}

func operatorFromFunction(f function) operator {
	return operator{
		oType: opFunction,
		card:  f.card,
		poper: func(out *outQueue) Expression {
			//pop from the queue, is done before
			res := &nExp{
				card:     f.card,
				children: make([]Expression, f.card),
				evaluer:  f.evaluer,
			}
			for i := 0; i < f.card; i++ {
				res.children[i] = out.unsafePop()
			}
			return res
		},
	}
}

var operators = make(map[TokenType]operator)
var functions = make(map[string]function)

func poperForBinaryOperator(evaluer binaryEvaluer) queuePoper {
	return func(output *outQueue) Expression {
		return &binaryExp{
			evaluer:    evaluer,
			rightChild: output.unsafePop(),
			leftChild:  output.unsafePop(),
		}
	}
}

func registerOperator(t TokenType,
	precedence int,
	leftAssociative bool,
	evaluer binaryEvaluer) {
	operators[t] = operator{
		oType:           opStandard,
		poper:           poperForBinaryOperator(evaluer),
		precedence:      precedence,
		leftAssociative: leftAssociative,
		card:            2,
	}
}

// RegisterFunction register a new function with the given
// cardinality. The list of float passed to the evaluer function is
// asserted to be cardinatily
func RegisterFunction(name string, cardinality uint, evaluer NEvaluer) {
	functions[name] = function{
		card:    int(cardinality),
		evaluer: evaluer,
	}
}

func init() {
	registerOperator(TokPlus, 2, true, func(a float64, b float64) float64 { return a + b })
	registerOperator(TokMinus, 2, true, func(a float64, b float64) float64 { return a - b })
	registerOperator(TokMult, 3, true, func(a float64, b float64) float64 { return a * b })
	registerOperator(TokDivide, 3, true, func(a float64, b float64) float64 { return a / b })
	registerOperator(TokPower, 4, false, func(a float64, b float64) float64 { return math.Pow(a, b) })

	RegisterFunction("pi", 0, func(a []float64) float64 { return math.Pi })
	RegisterFunction("rand", 0, func(a []float64) float64 { return rand.Float64() })

	RegisterFunction("sin", 1, func(a []float64) float64 { return math.Sin(a[0]) })
	RegisterFunction("cos", 1, func(a []float64) float64 { return math.Cos(a[0]) })
	RegisterFunction("tan", 1, func(a []float64) float64 { return math.Tan(a[0]) })
	RegisterFunction("asin", 1, func(a []float64) float64 { return math.Asin(a[0]) })
	RegisterFunction("acos", 1, func(a []float64) float64 { return math.Acos(a[0]) })
	RegisterFunction("atan", 1, func(a []float64) float64 { return math.Atan(a[0]) })
	RegisterFunction("sqrt", 1, func(a []float64) float64 { return math.Sqrt(a[0]) })
	RegisterFunction("exp", 1, func(a []float64) float64 { return math.Exp(a[0]) })
	RegisterFunction("ln", 1, func(a []float64) float64 { return math.Log(a[0]) })
	RegisterFunction("log10", 1, func(a []float64) float64 { return math.Log10(a[0]) })
	RegisterFunction("ceil", 1, func(a []float64) float64 { return math.Ceil(a[0]) })
	RegisterFunction("floor", 1, func(a []float64) float64 { return math.Floor(a[0]) })

}

func popOperatorFromStack(output *outQueue, stack *opStack) error {
	if stack.size() == 0 {
		return fmt.Errorf("Internal expression compilation error, stack should not be emptyin popOperatorFromStack")
	}
	op := stack.unsafePop()
	if output.size() < op.card {
		return fmt.Errorf("Evaluation stack error, need %d element, but only %d provided",
			op.card,
			output.size())
	}
	//will pop the stack and push it
	output.push(op.poper(output))
	return nil
}

func buildAST(input string) (Expression, error) {
	l := NewLexer(input)

	output := outQueue{}
	stack := opStack{}

	for {
		t, err := l.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		if t.Type == TokValue {
			if value, err := strconv.ParseFloat(t.Value, 64); err != nil {
				return nil, err
			} else {
				output.push(&valueExp{value: value})
			}
			continue
		}

		// checks for a function or a number
		if t.Type == TokIdent {
			if fn, ok := functions[t.Value]; ok == true {
				stack.push(operatorFromFunction(fn))
			} else {
				output.push(&refExp{variable: t.Value})
			}
			continue
		}

		if t.Type == TokComma {
			for stack.size() > 0 && stack.unsafeTop().oType != opLeftParenthesis {
				if err := popOperatorFromStack(&output, &stack); err != nil {
					return nil, err
				}
			}

			if stack.size() == 0 {
				return nil, fmt.Errorf("Misplaced comma or mismatched parenthese in %s", input)
			}
			continue
		}

		//get the operator from the token

		if op1, ok := operators[t.Type]; ok == true {
			for stack.size() > 0 {
				op2 := stack.unsafeTop()
				if op2.oType != opStandard {
					break
				}

				if op1.precedence < op2.precedence {
					popOperatorFromStack(&output, &stack)
					continue
				}

				if op1.leftAssociative && op1.precedence == op2.precedence {
					popOperatorFromStack(&output, &stack)
					continue
				}

				break
			}
			stack.push(op1)
			continue
		}

		if t.Type == TokOParen {
			stack.push(operator{
				oType: opLeftParenthesis,
				poper: func(*outQueue) Expression {
					return nil
				},
			})
			continue
		}

		if t.Type == TokCParen {
			for stack.size() > 0 && stack.unsafeTop().oType != opLeftParenthesis {
				if err := popOperatorFromStack(&output, &stack); err != nil {
					return nil, err
				}
			}
			if stack.size() == 0 {
				return nil, fmt.Errorf("Mismatched parenthese in %s", input)
			}
			stack.unsafePop()
			// pop the next if this is a function
			if stack.size() > 0 && stack.unsafeTop().oType == opFunction {
				if err := popOperatorFromStack(&output, &stack); err != nil {
					return nil, err
				}
			}
			continue
		}

		return nil, fmt.Errorf("Operator %s is not yet implemented", t.Value)

	}

	for stack.size() > 0 {
		if stack.unsafeTop().oType == opLeftParenthesis {
			return nil, fmt.Errorf("Mismatched parenthese in %s", input)
		}
		if err := popOperatorFromStack(&output, &stack); err != nil {
			return nil, err
		}
	}

	if output.size() != 1 {
		return nil, fmt.Errorf("Evaluation stack error, still got %d element instead of 1 %s",
			output.size(),
			output.q)
	}

	return output.unsafePop(), nil
}
