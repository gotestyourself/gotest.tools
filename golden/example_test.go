package golden

var t = &FakeT{}

func ExampleAssert() {
	Assert(t, "foo", "foo-content.golden")
}

func ExampleAssertBytes() {
	AssertBytes(t, []byte("foo"), "foo-content.golden")
}
