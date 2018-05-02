package main

import (
	"io/ioutil"
	"testing"

	"gotest.tools/assert"
	"gotest.tools/env"
	"gotest.tools/fs"
	"gotest.tools/golden"
)

func TestRun(t *testing.T) {
	setupLogging(&options{})
	dir := fs.NewDir(t, "test-run",
		fs.WithDir("src/example.com/example", fs.FromDir("testdata/full")))
	defer dir.Remove()

	defer env.Patch(t, "GOPATH", dir.Path())()
	err := run(options{
		pkgs: []string{"example.com/example"},
	})
	assert.NilError(t, err)

	raw, err := ioutil.ReadFile(dir.Join("src/example.com/example/some_test.go"))
	assert.NilError(t, err)
	golden.Assert(t, string(raw), "full-expected/some_test.go")
}
