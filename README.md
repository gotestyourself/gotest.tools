# Go Test Yourself

A collection of packages compatible with `go test` to support common testing
patterns.

## Packages

* `fs` - create test files and directories
* `golden` - compare large multi-line strings
* `gotestsum` - a program to summarize `go test` output and test failures
* `icmd` - execute binaries and test the output
* `skip` - skip tests based on conditions
* `tags` - annotate tests with tags and run only tests matching those tags


## Related

* [testify/assert](https://godoc.org/github.com/stretchr/testify/assert) and 
  [testify/require](https://godoc.org/github.com/stretchr/testify/require) -
  an assertion library with common assertions
* [golang/mock](https://github.com/golang/mock) - generate mocks for interfaces
* [testify/suite](https://godoc.org/github.com/stretchr/testify/suite) - 
  group test into suites to share common setup/teardown logic
