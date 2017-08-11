package fs

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromDir(t *testing.T) {
	dir := NewDir(t, "test-from-dir", FromDir("testdata/copy-test"))
	defer dir.Remove()

	assertFileWithContent(t, dir.Join("1"), "1\n")
	assertFileWithContent(t, dir.Join("a/1"), "1\n")
	assertFileWithContent(t, dir.Join("a/2"), "2\n")
	assertFileWithContent(t, dir.Join("a/b/1"), "1\n")
}

func assertFileWithContent(t *testing.T, path, content string) {
	actual, err := ioutil.ReadFile(path)
	require.NoError(t, err, "file %s does not exist", path)

	assert.Equal(t, content, string(actual), "file %s")
}
