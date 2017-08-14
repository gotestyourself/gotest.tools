package skip

var t = &fakeSkipT{}
var apiVersion = ""

type env struct{}

func (e env) hasFeature(_ string) bool { return false }

var testEnv = env{}

func MissingFeature() bool { return false }

func ExampleIf() {
	//   --- SKIP: TestName (0.00s)
	//           skip.go:19: MissingFeature
	If(t, MissingFeature)

	//   --- SKIP: TestName (0.00s)
	//           skip.go:19: MissingFeature: coming soon
	If(t, MissingFeature, "coming soon")
}

func ExampleIfCondition() {
	//   --- SKIP: TestName (0.00s)
	//           skip.go:19: apiVersion < version("v1.24")
	IfCondition(t, apiVersion < version("v1.24"))

	//   --- SKIP: TestName (0.00s)
	//           skip.go:19: !textenv.hasFeature("build"): coming soon
	IfCondition(t, !testEnv.hasFeature("build"), "coming soon")
}
