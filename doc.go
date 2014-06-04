// Copyright 2014 Alexandre Tuleu
// This file is part of go-meval.
//
// go-meval is free software: you can redistribute it and/or modify it
// under the terms of the GNU Lesser General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-meval is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public
// License along with go-meval.  If not, see
// <http://www.gnu.org/licenses/>.

/*
Package go-meval provides a mathematical expresison parser. It can
parse mathematical expresison into an AST and evaluate it. It also
support the concept of variable through Context : An Expression can
refer any other expression defined in a Context. For simple case use,
a nil Context could be used (any expression refering another
expression will fail at evaluation).

Basics

An expression can be parsed using Compile, and evaluated with a nil
context. See #Expression basic example.

Context

One can use context, in order for your expression to refer to others,
aka variable.

TODO(tuleu): document a context and how to use it

TODO(tuleu) package global example

*/
package meval
