package meval

import (
	. "gopkg.in/check.v1"
)

type ContextSuite struct {
	c *MapContext
}

var _ = Suite(&ContextSuite{
	c: NewMapContext(),
})

func (s *ContextSuite) TestCanStoreExpression(c *C) {
	err := s.c.CompileAndAdd("foo", "+0x")
	c.Assert(err, Not(IsNil))
}

func (s *ContextSuite) TestInvalidPushWillPanic(c *C) {
	didPanic := false
	defer func() {
		if r := recover(); r != nil {
			c.Check(r, Equals, "Should never happen")
			didPanic = true
		}
		c.Check(didPanic, Equals, true)
	}()
	s.c.pop()

}
