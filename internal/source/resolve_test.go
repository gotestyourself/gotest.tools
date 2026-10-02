package source_test

import (
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/internal/source"
)

func TestResolveSourceFile_AlreadyAbsolute(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "test.go")
	assert.NilError(t, os.WriteFile(tempFile, []byte("package foo"), 0o644))

	resolved, err := source.ResolveSourceFile(tempFile)
	assert.NilError(t, err)
	assert.Equal(t, tempFile, resolved)
}

func TestResolveSourceFile_TrimpathInModule(t *testing.T) {
	trimmed := "gotest.tools/v3/internal/source/resolve_test.go"
	resolved, err := source.ResolveSourceFile(trimmed)
	assert.NilError(t, err)

	info, err := os.Stat(resolved)
	assert.NilError(t, err)
	assert.Assert(t, !info.IsDir())
	assert.Equal(t, "resolve_test.go", filepath.Base(resolved))
}

func TestResolveSourceFile_TrimpathWithChdir(t *testing.T) {
	origWd, err := os.Getwd()
	assert.NilError(t, err)
	tempDir := t.TempDir()
	assert.NilError(t, os.Chdir(tempDir))
	defer func() { _ = os.Chdir(origWd) }()

	trimmed := "gotest.tools/v3/internal/source/resolve.go"
	resolved, err := source.ResolveSourceFile(trimmed)
	assert.NilError(t, err)

	info, err := os.Stat(resolved)
	assert.NilError(t, err)
	assert.Assert(t, !info.IsDir())
	assert.Equal(t, "resolve.go", filepath.Base(resolved))
}

func TestResolveSourceFile_ModuleDirEnv(t *testing.T) {
	tempDir := t.TempDir()
	goModContent := "module custom.module\n\ngo 1.18\n"
	assert.NilError(t, os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0o644))
	fakeSource := filepath.Join(tempDir, "pkg", "foo.go")
	assert.NilError(t, os.MkdirAll(filepath.Dir(fakeSource), 0o755))
	assert.NilError(t, os.WriteFile(fakeSource, []byte("package pkg"), 0o644))

	t.Setenv("GOTESTTOOLS_MODULE_DIR", tempDir)

	resolved, err := source.ResolveSourceFile("custom.module/pkg/foo.go")
	assert.NilError(t, err)
	assert.Equal(t, fakeSource, resolved)
}

func TestResolveSourceFile_NotFound(t *testing.T) {
	_, err := source.ResolveSourceFile("nonexistent.module/not_found.go")
	assert.ErrorContains(t, err, "failed to find source file")
	assert.ErrorContains(t, err, "-trimpath")
	assert.ErrorContains(t, err, "GOTESTTOOLS_MODULE_DIR")
}

func TestParseModulePath(t *testing.T) {
	tempDir := t.TempDir()

	writeGoMod := func(name, content string) string {
		p := filepath.Join(tempDir, name)
		assert.NilError(t, os.WriteFile(p, []byte(content), 0o644))
		return p
	}

	t.Run("standard module directive", func(t *testing.T) {
		p := writeGoMod("go.mod.std", "module example.com/foo/bar\n\ngo 1.20\n")
		mod, err := source.ParseModulePath(p)
		assert.NilError(t, err)
		assert.Equal(t, "example.com/foo/bar", mod)
	})

	t.Run("with inline comment", func(t *testing.T) {
		p := writeGoMod("go.mod.comment", "module example.com/foo/bar // some comment\n\ngo 1.20\n")
		mod, err := source.ParseModulePath(p)
		assert.NilError(t, err)
		assert.Equal(t, "example.com/foo/bar", mod)
	})

	t.Run("with quotes", func(t *testing.T) {
		p := writeGoMod("go.mod.quotes", "module \"example.com/foo/bar\"\n\ngo 1.20\n")
		mod, err := source.ParseModulePath(p)
		assert.NilError(t, err)
		assert.Equal(t, "example.com/foo/bar", mod)
	})

	t.Run("multiline module block", func(t *testing.T) {
		p := writeGoMod("go.mod.block", "module (\n\t// comment\n\texample.com/foo/bar\n)\n\ngo 1.20\n")
		mod, err := source.ParseModulePath(p)
		assert.NilError(t, err)
		assert.Equal(t, "example.com/foo/bar", mod)
	})

	t.Run("no module directive", func(t *testing.T) {
		p := writeGoMod("go.mod.empty", "// comment only\ngo 1.20\n")
		_, err := source.ParseModulePath(p)
		assert.ErrorContains(t, err, "module directive not found")
	})
}

func TestFindModule(t *testing.T) {
	tempDir := t.TempDir()
	modDir := filepath.Join(tempDir, "myrepo")
	subDir := filepath.Join(modDir, "pkg", "sub")
	assert.NilError(t, os.MkdirAll(subDir, 0o755))
	assert.NilError(t, os.WriteFile(filepath.Join(modDir, "go.mod"), []byte("module example.com/myrepo\n"), 0o644))

	root, path := source.FindModule(subDir)
	assert.Equal(t, modDir, root)
	assert.Equal(t, "example.com/myrepo", path)

	noRoot, noPath := source.FindModule(tempDir)
	assert.Equal(t, "", noRoot)
	assert.Equal(t, "", noPath)
}
