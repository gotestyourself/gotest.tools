/*Package env provides functions to test code that read environment variables
or the current working directory.
*/
package env

import (
	"os"
	"strings"

	"github.com/gotestyourself/gotestyourself/assert"
)

// Patch changes the value of an environment variable, and returns a
// function which will reset the the value of that variable back to the
// previous state.
func Patch(t assert.TestingT, key, value string) func() {
	assert := assert.New(t)
	oldValue, ok := os.LookupEnv(key)
	assert.NilError(os.Setenv(key, value))
	return func() {
		if !ok {
			assert.NilError(os.Unsetenv(key))
			return
		}
		assert.NilError(os.Setenv(key, oldValue))
	}
}

// PatchAll sets the environment to env, and returns a function which will
// reset the environment back to the previous state.
func PatchAll(t assert.TestingT, env map[string]string) func() {
	assert := assert.New(t)
	oldEnv := os.Environ()
	os.Clearenv()

	for key, value := range env {
		assert.NilError(os.Setenv(key, value))
	}
	return func() {
		os.Clearenv()
		for key, oldVal := range ToMap(oldEnv) {
			assert.NilError(os.Setenv(key, oldVal))
		}
	}
}

// ToMap takes a list of strings in the format returned by os.Environ() and
// returns a mapping of keys to values.
func ToMap(env []string) map[string]string {
	result := map[string]string{}
	for _, raw := range env {
		parts := strings.SplitN(raw, "=", 2)
		switch len(parts) {
		case 1:
			result[raw] = ""
		case 2:
			result[parts[0]] = parts[1]
		}
	}
	return result
}

// ChangeWorkingDir to the directory, and return a function which restores the
// previous working directory.
func ChangeWorkingDir(t assert.TestingT, dir string) func() {
	cwd, err := os.Getwd()
	assert.NilError(t, err)
	assert.NilError(t, os.Chdir(dir))
	return func() {
		assert.NilError(t, os.Chdir(cwd))
	}
}
