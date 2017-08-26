package env

import (
	"os"
	"testing"

	"sort"

	"github.com/gotestyourself/gotestyourself/skip"
	"github.com/stretchr/testify/assert"
)

func TestPatchFromUnset(t *testing.T) {
	key, value := "FOO_IS_UNSET", "VALUE"
	revert := Patch(t, key, value)

	assert.Equal(t, value, os.Getenv(key))
	revert()
	_, isSet := os.LookupEnv(key)
	assert.False(t, isSet)
}

func TestPatch(t *testing.T) {
	skip.IfCondition(t, os.Getenv("PATH") == "")
	oldVal := os.Getenv("PATH")

	key, value := "PATH", "NEWVALUE"
	revert := Patch(t, key, value)

	assert.Equal(t, value, os.Getenv(key))
	revert()
	assert.Equal(t, oldVal, os.Getenv(key))
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
	assert.Equal(t, []string{"FIRST=STARS", "THEN=MOON"}, actual)

	revert()
	assert.Equal(t, sorted(oldEnv), sorted(os.Environ()))
}

func sorted(source []string) []string {
	sort.Strings(source)
	return source
}

func TestToMap(t *testing.T) {
	source := []string{"key=value", "novaluekey"}
	actual := ToMap(source)
	expected := map[string]string{"key": "value", "novaluekey": ""}
	assert.Equal(t, expected, actual)
}
