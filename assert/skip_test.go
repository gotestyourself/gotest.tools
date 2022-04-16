package assert

import (
	"bytes"
	"fmt"
	"testing"

	"gotest.tools/v3/assert/cmp"
)

type fakeSkipT struct {
	reason string
	logs   []string
}

func (f *fakeSkipT) Skip(args ...interface{}) {
	buf := new(bytes.Buffer)
	for _, arg := range args {
		buf.WriteString(fmt.Sprintf("%s", arg))
	}
	f.reason = buf.String()
}

func (f *fakeSkipT) Log(args ...interface{}) {
	f.logs = append(f.logs, fmt.Sprintf("%s", args[0]))
}

func (f *fakeSkipT) Helper() {}

func version(v string) string {
	return v
}

func TestSkipIFCondition(t *testing.T) {
	skipT := &fakeSkipT{}
	apiVersion := "v1.4"
	SkipIf(skipT, apiVersion < version("v1.6"))

	Equal(t, `apiVersion < version("v1.6")`, skipT.reason)
	Assert(t, cmp.Len(skipT.logs, 0))
}

func TestSkipIfConditionWithMessage(t *testing.T) {
	skipT := &fakeSkipT{}
	apiVersion := "v1.4"
	SkipIf(skipT, apiVersion < "v1.6", "see notes")

	Equal(t, `apiVersion < "v1.6": see notes`, skipT.reason)
	Assert(t, cmp.Len(skipT.logs, 0))
}

func TestSkipIfConditionMultiline(t *testing.T) {
	skipT := &fakeSkipT{}
	apiVersion := "v1.4"
	SkipIf(
		skipT,
		apiVersion < "v1.6")

	Equal(t, `apiVersion < "v1.6"`, skipT.reason)
	Assert(t, cmp.Len(skipT.logs, 0))
}

func TestSkipIfConditionMultilineWithMessage(t *testing.T) {
	skipT := &fakeSkipT{}
	apiVersion := "v1.4"
	SkipIf(
		skipT,
		apiVersion < "v1.6",
		"see notes")

	Equal(t, `apiVersion < "v1.6": see notes`, skipT.reason)
	Assert(t, cmp.Len(skipT.logs, 0))
}

func TestSkipIfConditionNoSkip(t *testing.T) {
	skipT := &fakeSkipT{}
	SkipIf(skipT, false)

	Equal(t, "", skipT.reason)
	Assert(t, cmp.Len(skipT.logs, 0))
}

func SkipBecauseISaidSo() bool {
	return true
}

func TestSkipIf(t *testing.T) {
	skipT := &fakeSkipT{}
	SkipIf(skipT, SkipBecauseISaidSo)

	Equal(t, "SkipBecauseISaidSo", skipT.reason)
}

func TestSkipIfWithMessage(t *testing.T) {
	skipT := &fakeSkipT{}
	SkipIf(skipT, SkipBecauseISaidSo, "see notes")

	Equal(t, "SkipBecauseISaidSo: see notes", skipT.reason)
}

func TestSkipIf_InvalidCondition(t *testing.T) {
	skipT := &fakeSkipT{}
	Assert(t, cmp.Panics(func() {
		SkipIf(skipT, "just a string")
	}))
}

func TestSkipIfWithSkipResultFunc(t *testing.T) {
	t.Run("no extra message", func(t *testing.T) {
		skipT := &fakeSkipT{}
		SkipIf(skipT, alwaysSkipWithMessage)

		Equal(t, "alwaysSkipWithMessage: skip because I said so!", skipT.reason)
	})
	t.Run("with extra message", func(t *testing.T) {
		skipT := &fakeSkipT{}
		SkipIf(skipT, alwaysSkipWithMessage, "also %v", 4)

		Equal(t, "alwaysSkipWithMessage: skip because I said so!: also 4", skipT.reason)
	})
}

func alwaysSkipWithMessage() SkipResult {
	return skipResult{}
}

type skipResult struct{}

func (s skipResult) Skip() bool {
	return true
}

func (s skipResult) Message() string {
	return "skip because I said so!"
}
