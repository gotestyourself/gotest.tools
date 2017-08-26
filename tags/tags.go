/*Package tags provides functions for skipping tests based on tags specified
on the command line.

Tags are set using the -test.run-tags flag:

    go test -v ./... -test.run-tags +web,-slow

There are three types of tags:

Required Tags

A tag that starts with `+` is a required tag. Only tests which match all
required tags will run.

    -test.run-tags +db,+sql

Omit Tags

A tag that starts with `!` is an omit tag. Any test which matches any omit tag
will be skipped.

    -test.run-tags '!windows,!slow,!flaky'

Note: Make sure to single quote the '!' character so that it is not handled by
the shell.

Any Tags

A tag without any prefix is an any tag. Any test which matches any one of these
tags will be run.

	-test.run-tags app,web

Note: you must call `tags.Apply(t)` in all tests that you wish to skip. Any test
that does not call Apply() will not be skipped when required or any tags are
used.
*/
package tags

import (
	"encoding/csv"
	"flag"
	"strings"
)

type skipT interface {
	Skipf(format string, args ...interface{})
}

// requested stores tags specified by the user using the command line flag
var requested = []string{}

func init() {
	flag.Var(&sliceValue{}, "test.run-tags", "select tests to run based on their tags")
}

const (
	required = '+'
	skip     = '!'
)

// Apply tags to a test. When the test is run, these applied tags are matched
// against the tags specified by the `-test.run-tags` flag to determine if the test
// should be run or skipped.
func Apply(t skipT, tags ...string) {
	if len(requested) == 0 {
		return
	}

	appliedSet := newSet(tags)
	anyMatch := false

	for _, request := range requested {
		anyMatch = matchTag(t, request, appliedSet) || anyMatch
	}
	if !anyMatch {
		t.Skipf("no matching tag")
	}
}

func matchTag(t skipT, tag string, applied set) bool {
	if len(tag) < 2 {
		return applied.contains(tag)
	}

	first, value := tag[0], tag[1:]
	switch {
	case first == skip && applied.contains(value):
		t.Skipf("matched tag: %s", tag)
	case first == required && !applied.contains(value):
		t.Skipf("missing required tag: %s", value)
	case first != skip && first != required:
		return applied.contains(tag)
	}
	return true
}

type set struct {
	items map[string]bool
}

func (s *set) contains(item string) bool {
	return s.items[item]
}

func (s *set) add(item string) {
	s.items[item] = true
}

func newSet(items []string) set {
	s := set{items: make(map[string]bool)}
	for _, item := range items {
		s.add(item)
	}
	return s
}

type sliceValue struct{}

func (v *sliceValue) Set(value string) error {
	csvReader := csv.NewReader(strings.NewReader(value))
	values, err := csvReader.Read()
	requested = append(requested, values...)
	return err
}

func (v *sliceValue) String() string {
	return strings.Join(requested, ",")
}
