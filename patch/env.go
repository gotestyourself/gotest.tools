/*Package patch provides functions to patch environment variables
 */
package patch

import (
	"os"
	"strings"

	"github.com/stretchr/testify/require"
)

// EnvVariable changes the value of an environment variable, and returns a
// function which will reset the environment back to the previous state.
func EnvVariable(t require.TestingT, key, value string) func() {
	oldValue, ok := os.LookupEnv(key)
	require.NoError(t, os.Setenv(key, value))
	return func() {
		if !ok {
			require.NoError(t, os.Unsetenv(key))
			return
		}
		require.NoError(t, os.Setenv(key, oldValue))
	}
}

// Environment sets the environment to env, and returns a function which will
// reset the environment back to the previous state
func Environment(t require.TestingT, env map[string]string) func() {
	oldEnv := os.Environ()
	os.Clearenv()

	for key, value := range env {
		require.NoError(t, os.Setenv(key, value))
	}
	return func() {
		os.Clearenv()
		for _, oldVar := range oldEnv {
			parts := strings.SplitN(oldVar, "=", 2)
			require.Len(t, parts, 2)
			require.NoError(t, os.Setenv(parts[0], parts[1]))
		}
	}
}
