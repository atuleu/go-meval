package meval

import "fmt"

type callStack interface {
	push(e *refExp)
	pop()
	testStack(e *refExp) (bool, []string)
}

// A Context is a kind of dictionnary of expression. You can pass it
// to Eval.
type Context interface {
	// Returns an expression from a given name.
	GetExpression(string) (Expression, error)
	callStack
}

// CallStack is meant to be embedded in any type that want to
// implement Context
type CallStack struct {
	stack []*refExp
}

func (c *CallStack) push(e *refExp) {
	c.stack = append(c.stack, e)
}

func (c *CallStack) pop() {
	if len(c.stack) == 0 {
		panic("Should never happen")
	}
	c.stack = c.stack[0 : len(c.stack)-1]
}

func (c *CallStack) testStack(e *refExp) (bool, []string) {
	res := false
	var deps []string
	for _, ee := range c.stack {
		if e == ee {
			res = true
		}
		if res == true {
			deps = append(deps, ee.variable)
		}
	}
	return res, deps
}

// MapContext represents the most simple context, aka a dictionnary of
// expressions.
type MapContext struct {
	// In order to be a Context, one should embed a CallStack
	CallStack

	exprs map[string]Expression
}

// NewMapContext creates a MapContext
func NewMapContext() *MapContext {
	return &MapContext{exprs: make(map[string]Expression)}
}

// GetExpression returns an Expression storred in the MapContext, or nil
func (c *MapContext) GetExpression(name string) (Expression, error) {
	if e, ok := c.exprs[name]; ok == true {
		return e, nil
	}
	return nil, fmt.Errorf("Could not find '%s' in MapContext", name)
}

// Add adds a new expression to the MapContext
func (c *MapContext) Add(name string, e Expression) {
	c.exprs[name] = e
}

// CompileAndAdd compiles and adds a new expression to the MapContext.
//
// It returns the same errors than Compile()
func (c *MapContext) CompileAndAdd(name, input string) error {
	if e, err := Compile(input); err != nil {
		return err
	} else {
		c.Add(name, e)
	}
	return nil
}

// Delete deletes the given expression from the MapContext if it
// exists.
func (c *MapContext) Delete(name string) {
	delete(c.exprs, name)
}
