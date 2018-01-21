package env

import (
	"os"
	"testing"

	"sort"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/gotestyourself/gotestyourself/skip"
)

func TestPatchFromUnset(t *testing.T) {
	key, value := "FOO_IS_UNSET", "VALUE"
	revert := Patch(t, key, value)

	assert.Assert(t, value == os.Getenv(key))
	revert()
	_, isSet := os.LookupEnv(key)
	assert.Assert(t, !isSet)
}

func TestPatch(t *testing.T) {
	skip.If(t, os.Getenv("PATH") == "")
	oldVal := os.Getenv("PATH")

	key, value := "PATH", "NEWVALUE"
	revert := Patch(t, key, value)

	assert.Assert(t, value == os.Getenv(key))
	revert()
	assert.Assert(t, oldVal == os.Getenv(key))
}

func TestPatchAll(t *testing.T) {
	oldEnv := os.Environ()
	newEnv := map[string]string{
		"FIRST": "STARS",
		"THEN":  "MOON",
	}

	revert := PatchAll(t, newEnv)

	actual := os.Environ()
	sort.Strings(actual)
	assert.Assert(t, cmp.DeepEqual([]string{"FIRST=STARS", "THEN=MOON"}, actual))

	revert()
	assert.Assert(t, cmp.DeepEqual(sorted(oldEnv), sorted(os.Environ())))
}

func sorted(source []string) []string {
	sort.Strings(source)
	return source
}

func TestToMap(t *testing.T) {
	source := []string{"key=value", "novaluekey"}
	actual := ToMap(source)
	expected := map[string]string{"key": "value", "novaluekey": ""}
	assert.Assert(t, cmp.DeepEqual(expected, actual))
}
