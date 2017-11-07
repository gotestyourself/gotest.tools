# Go Test Yourself

A collection of packages compatible with `go test` to support common testing
patterns.

[![GoDoc](https://godoc.org/github.com/gotestyourself/gotestyourself?status.svg)](https://godoc.org/github.com/gotestyourself/gotestyourself)
[![CircleCI](https://circleci.com/gh/gotestyourself/gotestyourself/tree/master.svg?style=shield)](https://circleci.com/gh/gotestyourself/gotestyourself/tree/master)
[![Go Reportcard](https://goreportcard.com/badge/github.com/gotestyourself/gotestyourself)](https://goreportcard.com/report/github.com/gotestyourself/gotestyourself)


## Packages

* [env](http://godoc.org/github.com/gotestyourself/gotestyourself/env) -
  test code that uses environment variables
* [fs](http://godoc.org/github.com/gotestyourself/gotestyourself/fs) -
  create test files and directories
* [golden](http://godoc.org/github.com/gotestyourself/gotestyourself/golden) -
  compare large multi-line strings
* [icmd](http://godoc.org/github.com/gotestyourself/gotestyourself/icmd) -
  execute binaries and test the output
* [poll](http://godoc.org/github.com/gotestyourself/gotestyourself/poll) -
  test asynchronous code by polling until a desired state is reached
* [skip](http://godoc.org/github.com/gotestyourself/gotestyourself/skip) -
  skip tests based on conditions
* [testsum](http://godoc.org/github.com/gotestyourself/gotestyourself/testsum) -
  a program to summarize `go test` output and test failures

## Related

* [testify/assert](https://godoc.org/github.com/stretchr/testify/assert) and 
  [testify/require](https://godoc.org/github.com/stretchr/testify/require) -
  assertion libraries with common assertions
* [maxbrunsfeld/counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) - generate fakes for interfaces
* [testify/suite](https://godoc.org/github.com/stretchr/testify/suite) - 
  group test into suites to share common setup/teardown logic
