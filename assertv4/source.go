package assertv4

func fail(t TestingT, msg string, args ...any) bool {
	t.Logf(msg, args...)
	t.FailNow()
	return false
}
