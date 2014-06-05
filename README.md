go-meval
========

[![Build Status](https://drone.io/github.com/atuleu/go-meval/status.png)](https://drone.io/github.com/atuleu/go-meval/latest) 
[![Coverage Status](https://coveralls.io/repos/atuleu/go-meval/badge.png?branch=master)](https://coveralls.io/r/atuleu/go-meval?branch=master)


A mathematical expression  parser / evaluator in go.


It uses simply the shutting yard algorithm to parse a mathemical expression in an AST.

This AST can then be evaluated within a context, letting user define a dictionnary of expression that can refer other expression. It is not designed to be fast, but only to provide an easy way to mathematically parametrize your program at a configuration level.

## 1. Installation

```bash
go get github.com/atuleu/go-meval
```

## 2. Roadmap

Still in alpha release. No complex github issue for that task list

see [roadmap](roadmap.md)

# 3. Contribute 
