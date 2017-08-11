package golden

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type FakeT struct {
	Failed bool
}

func (t *FakeT) Fatal(a ...interface{}) {
	t.Failed = true
}

func (t *FakeT) Fatalf(string, ...interface{}) {
	t.Failed = true
}

func (t *FakeT) Errorf(_ string, _ ...interface{}) {
}

func (t *FakeT) FailNow() {
	t.Failed = true
}

func TestGoldenGetInvalidFile(t *testing.T) {
	fakeT := new(FakeT)

	Get(fakeT, "/invalid/path")
	require.True(t, fakeT.Failed)
}

func TestGoldenGet(t *testing.T) {
	expected := "content\nline1\nline2"

	filename, clean := setupGoldenFile(t, expected)
	defer clean()

	fakeT := new(FakeT)

	actual := Get(fakeT, filename)
	assert.False(t, fakeT.Failed)
	assert.Equal(t, actual, []byte(expected))
}

func TestGoldenAssertInvalidContent(t *testing.T) {
	filename, clean := setupGoldenFile(t, "content")
	defer clean()

	fakeT := new(FakeT)

	success := Assert(fakeT, "foo", filename)
	assert.False(t, fakeT.Failed)
	assert.False(t, success)
}

func TestGoldenAssert(t *testing.T) {
	filename, clean := setupGoldenFile(t, "foo")
	defer clean()

	fakeT := new(FakeT)

	success := Assert(fakeT, "foo", filename)
	assert.False(t, fakeT.Failed)
	assert.True(t, success)
}

func setupGoldenFile(t *testing.T, content string) (string, func()) {
	_ = os.Mkdir("testdata", 0755)
	f, err := ioutil.TempFile("testdata", "")
	require.NoError(t, err, "fail to setup test golden file")
	err = ioutil.WriteFile(f.Name(), []byte(content), 0660)
	require.NoError(t, err, "fail to write test golden file with %q", content)
	_, name := filepath.Split(f.Name())
	t.Log(f.Name(), name)
	return name, func() {
		require.NoError(t, os.Remove(f.Name()))
	}
}
