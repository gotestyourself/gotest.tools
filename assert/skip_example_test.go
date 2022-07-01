package assert_test

import (
	"gotest.tools/v3/assert"
)

var apiVersion = ""

type env struct{}

func (e env) hasFeature(_ string) bool { return false }

var testEnv = env{}

func MissingFeature() bool { return false }

func ExampleSkipIf() {
	assert.SkipIf(t, MissingFeature)
	//   --- SKIP: TestName (0.00s)
	//           skip.go:18: MissingFeature

	assert.SkipIf(t, MissingFeature, "coming soon")
	//   --- SKIP: TestName (0.00s)
	//           skip.go:22: MissingFeature: coming soon
}

func ExampleSkipIf_withExpression() {
	assert.SkipIf(t, apiVersion < version("v1.24"))
	//   --- SKIP: TestName (0.00s)
	//           skip.go:28: apiVersion < version("v1.24")

	assert.SkipIf(t, !testEnv.hasFeature("build"), "coming soon")
	//   --- SKIP: TestName (0.00s)
	//           skip.go:32: !textenv.hasFeature("build"): coming soon
}

func version(v string) string {
	return v
}
