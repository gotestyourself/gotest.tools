package patch

import "testing"

var t = &testing.T{}

// Patch an environment variable and defer to return to the previous state
func ExampleEnvVariable() {
	defer EnvVariable(t, "PATH", "/custom/path")()
}

// Patch all environment variables
func ExampleEnvironment() {
	defer Environment(t, map[string]string{
		"ONE": "FOO",
		"TWO": "BAR",
	})()
}
