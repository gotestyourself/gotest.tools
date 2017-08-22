package icmd_test

import "github.com/gotestyourself/gotestyourself/icmd"

type fakeTesting struct{}

func (t fakeTesting) Fatalf(string, ...interface{}) {}

var t = fakeTesting{}

func ExampleRunCommand() {
	result := icmd.RunCommand("bash", "-c", "echo all good")
	result.Assert(t, icmd.Success)
}

func ExampleRunCmd() {
	result := icmd.RunCmd(icmd.Command("cat", "/does/not/exist"))
	result.Assert(t, icmd.Expected{
		ExitCode: 1,
		Err:      "cat: /does/not/exist: No such file or directory",
	})
}
