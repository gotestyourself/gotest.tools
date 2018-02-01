package icmd

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/gotestyourself/gotestyourself/skip"
)

func TestRunCommandSuccess(t *testing.T) {
	skip.If(t, runtime.GOOS == "windows", "ls not available")

	result := RunCommand("ls")
	result.Assert(t, Success)
}

func TestRunCommandWithCombined(t *testing.T) {
	skip.If(t, runtime.GOOS == "windows", "ls not available")

	result := RunCommand("ls", "-a")
	result.Assert(t, Expected{})

	assert.Assert(t, cmp.Contains(result.Combined(), "\n..\n"))
	assert.Assert(t, cmp.Contains(result.Stdout(), "\n..\n"))
}

func TestRunCommandWithTimeoutFinished(t *testing.T) {
	skip.If(t, runtime.GOOS == "windows", "ls not available")

	result := RunCmd(Cmd{
		Command: []string{"ls", "-a"},
		Timeout: 50 * time.Millisecond,
	})
	result.Assert(t, Expected{Out: ".."})
}

func TestRunCommandWithTimeoutKilled(t *testing.T) {
	skip.If(t, runtime.GOOS == "windows", "sh not available")

	command := []string{"sh", "-c", "while true ; do echo 1 ; sleep .5 ; done"}
	result := RunCmd(Cmd{Command: command, Timeout: 1250 * time.Millisecond})
	result.Assert(t, Expected{Timeout: true})

	ones := strings.Split(result.Stdout(), "\n")
	assert.Assert(t, cmp.Len(ones, 4))
}

func TestRunCommandWithErrors(t *testing.T) {
	result := RunCommand("doesnotexists")
	expected := `exec: "doesnotexists": executable file not found`
	result.Assert(t, Expected{Out: None, Err: None, ExitCode: 127, Error: expected})
}

func TestRunCommandWithStdoutStderr(t *testing.T) {
	result := RunCommand("echo", "hello", "world")
	result.Assert(t, Expected{Out: "hello world\n", Err: None})
}

func TestRunCommandWithStdoutStderrError(t *testing.T) {
	expected := "ls: unrecognized option"

	result := RunCommand("ls", "-z")
	result.Assert(t, Expected{
		Out:      None,
		Err:      expected,
		ExitCode: 1,
		Error:    "exit status 1",
	})
}
