package patch

import (
	"os"
	"testing"

	"sort"

	"github.com/gotestyourself/gotestyourself/skip"
	"github.com/stretchr/testify/assert"
)

func TestEnvVariableFromUnset(t *testing.T) {
	key, value := "FOO_IS_UNSET", "VALUE"
	revert := EnvVariable(t, key, value)

	assert.Equal(t, value, os.Getenv(key))
	revert()
	_, isSet := os.LookupEnv(key)
	assert.False(t, isSet)
}

func TestEnvVariable(t *testing.T) {
	skip.IfCondition(t, os.Getenv("PATH") == "")
	oldVal := os.Getenv("PATH")

	key, value := "PATH", "NEWVALUE"
	revert := EnvVariable(t, key, value)

	assert.Equal(t, value, os.Getenv(key))
	revert()
	assert.Equal(t, oldVal, os.Getenv(key))
}

func TestEnvironment(t *testing.T) {
	oldEnv := os.Environ()
	newEnv := map[string]string{
		"FIRST": "STARS",
		"THEN":  "MOON",
	}

	revert := Environment(t, newEnv)

	actual := os.Environ()
	sort.Strings(actual)
	assert.Equal(t, []string{"FIRST=STARS", "THEN=MOON"}, actual)

	revert()
	assert.Equal(t, oldEnv, os.Environ())
}
