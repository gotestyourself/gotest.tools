package skip_test

import (
	"testing"

	"github.com/gotestyourself/gotestyourself/skip"
)

var apiVersion = ""

type env struct{}

func (e env) hasFeature(_ string) bool { return false }

var testEnv = env{}

func MissingFeature() bool { return false }

func ExampleIf() {
	//   --- SKIP: TestName (0.00s)
	//           skip.go:19: MissingFeature
	skip.If(&testing.T{}, MissingFeature)

	//   --- SKIP: TestName (0.00s)
	//           skip.go:19: MissingFeature: coming soon
	skip.If(&testing.T{}, MissingFeature, "coming soon")
}

func ExampleIfCondition() {
	//   --- SKIP: TestName (0.00s)
	//           skip.go:19: apiVersion < version("v1.24")
	skip.IfCondition(&testing.T{}, apiVersion < version("v1.24"))

	//   --- SKIP: TestName (0.00s)
	//           skip.go:19: !textenv.hasFeature("build"): coming soon
	skip.IfCondition(&testing.T{}, !testEnv.hasFeature("build"), "coming soon")
}

func version(v string) string {
	return v
}
