package icmd

type fakeTesting struct{}

func (t fakeTesting) Fatalf(string, ...interface{}) {}

var t = fakeTesting{}

func ExampleRunCommand() {
	result := RunCommand("bash", "-c", "echo all good")
	result.Assert(t, Success)
}

func ExampleRunCmd() {
	result := RunCmd(Command("cat", "/does/not/exist"))
	result.Assert(t, Expected{
		ExitCode: 1,
		Err:      "cat: /does/not/exist: No such file or directory",
	})
}
