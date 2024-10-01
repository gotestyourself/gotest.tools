package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/env"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/golden"
)

func goGet(t *testing.T, pkg string) {
	t.Log("go get", pkg)
	cmd := exec.Command("go", "get", pkg)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestRun(t *testing.T) {
	setupLogging(&options{})
	dir := fs.NewDir(t, "test-run",
		fs.WithDir("src/example.com/example", fs.FromDir("testdata/full")))
	defer dir.Remove()

	defer env.Patch(t, "GO111MODULE", "off")()
	defer env.Patch(t, "GOPATH", dir.Path())()

	// Fetch dependencies in GOPATH mode
	// Check list in testdata/full/some_test.go
	goGet(t, "gopkg.in/check.v1")
	goGet(t, "github.com/stretchr/testify/assert")
	goGet(t, "github.com/stretchr/testify/require")

	err := run(options{
		pkgs:             []string{"example.com/example"},
		showLoaderErrors: true,
	})
	assert.NilError(t, err)

	raw, err := os.ReadFile(dir.Join("src/example.com/example/some_test.go"))
	assert.NilError(t, err)
	golden.Assert(t, string(raw), "full-expected/some_test.go")
}

func TestSetupFlags(t *testing.T) {
	flags, opts := setupFlags("testing")
	assert.Assert(t, flags.Usage != nil)

	err := flags.Parse([]string{
		"--dry-run",
		"--debug",
		"--cmp-pkg-import-alias=foo",
		"--print-loader-errors",
	})
	assert.NilError(t, err)
	expected := &options{
		dryRun:           true,
		debug:            true,
		cmpImportName:    "foo",
		showLoaderErrors: true,
	}
	assert.DeepEqual(t, opts, expected, cmpOptions)
}

var cmpOptions = cmp.AllowUnexported(options{})
