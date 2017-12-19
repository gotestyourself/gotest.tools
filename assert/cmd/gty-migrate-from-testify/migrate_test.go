package main

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/assert/cmp"
	"golang.org/x/tools/go/loader"
)

func TestMigrateFileReplacesTestingT(t *testing.T) {
	source := `
package foo

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSomething(t *testing.T) {
	a := assert.TestingT
	b := require.TestingT
	c := require.TestingT(t)
	if a == b {}
}

func do(t require.TestingT) {}
`
	migration := newMigrationFromSource(t, source)
	migrateFile(migration)

	expected := `package foo

import (
	"testing"

	"github.com/gotestyourself/gotestyourself/assert"
)

func TestSomething(t *testing.T) {
	a := assert.TestingT
	b := assert.TestingT
	c := assert.TestingT(t)
	if a == b {
	}
}

func do(t assert.TestingT) {}
`
	actual, err := formatFile(migration)
	assert.NilError(t, err)
	assert.Assert(t, cmp.EqualMultiLine(expected, string(actual)))
}

func newMigrationFromSource(t *testing.T, source string) migration {
	fileset := token.NewFileSet()
	nodes, err := parser.ParseFile(
		fileset,
		"foo.go",
		source,
		parser.AllErrors|parser.ParseComments)
	assert.NilError(t, err)

	opts := options{}
	conf := loader.Config{
		Fset:        fileset,
		ParserMode:  parser.ParseComments,
		Build:       buildContext(opts),
		AllowErrors: true,
	}
	conf.TypeChecker.Error = func(e error) {}
	conf.CreateFromFiles("foo.go", nodes)
	prog, err := conf.Load()
	assert.NilError(t, err)

	pkgInfo := prog.InitialPackages()[0]

	return migration{
		file:        pkgInfo.Files[0],
		fileset:     fileset,
		importNames: newImportNames(nodes.Imports, opts),
		pkgInfo:     pkgInfo,
	}
}

func TestMigrateFileWithNamedCmpPackage(t *testing.T) {
	source := `
package foo

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	assert.Equal(t, "a", "b")
}
`
	migration := newMigrationFromSource(t, source)
	migration.importNames.cmp = "is"
	migrateFile(migration)

	expected := `package foo

import (
	"testing"

	"github.com/gotestyourself/gotestyourself/assert"
	is "github.com/gotestyourself/gotestyourself/assert/cmp"
)

func TestSomething(t *testing.T) {
	assert.Check(t, is.Equal("a", "b"))
}
`
	actual, err := formatFile(migration)
	assert.NilError(t, err)
	assert.Assert(t, cmp.EqualMultiLine(expected, string(actual)))
}

func TestMigrateFileWithCommentsOnAssert(t *testing.T) {
	source := `
package foo

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	// This is going to fail
	assert.Equal(t, "a", "b")
}
`
	migration := newMigrationFromSource(t, source)
	migrateFile(migration)

	expected := `package foo

import (
	"testing"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/assert/cmp"
)

func TestSomething(t *testing.T) {
	// This is going to fail
	assert.Check(t, cmp.Equal("a", "b"))
}
`
	actual, err := formatFile(migration)
	assert.NilError(t, err)
	assert.Assert(t, cmp.EqualMultiLine(expected, string(actual)))
}

func TestMigrateFileConvertNilToNilError(t *testing.T) {
	source := `
package foo

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	var err error
	assert.Nil(t, err)
	require.Nil(t, err)
}
`
	migration := newMigrationFromSource(t, source)
	migrateFile(migration)

	expected := `package foo

import (
	"testing"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/assert/cmp"
)

func TestSomething(t *testing.T) {
	var err error
	assert.Check(t, cmp.NilError(err))
	assert.NilError(t, err)
}
`
	actual, err := formatFile(migration)
	assert.NilError(t, err)
	assert.Assert(t, cmp.EqualMultiLine(expected, string(actual)))
}
