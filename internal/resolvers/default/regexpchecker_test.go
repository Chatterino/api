package defaultresolver

import (
	"errors"
	"fmt"
	"regexp"

	qt "github.com/frankban/quicktest"
)

type argNames []string

func (a argNames) ArgNames() []string {
	return a
}

var MatchesRegexp qt.Checker = &regexpChecker{
	argNames: []string{"got value", "regexp"},
}

type regexpChecker struct {
	argNames
}

// match checks that the given error message matches the given pattern.
func match(got string, pattern *regexp.Regexp, msg string, note func(key string, value interface{})) error {
	if pattern.MatchString(got) {
		return nil
	}

	return errors.New(msg)
}

func (c *regexpChecker) Check(got interface{}, args []interface{}, note func(key string, value interface{})) error {
	switch pattern := args[0].(type) {
	case *regexp.Regexp:
		switch v := got.(type) {
		case string:
			return match(v, pattern, "value does not match regexp", note)
		case fmt.Stringer:
			return match(v.String(), pattern, "value.String() does not match regexp", note)
		}
		return qt.BadCheckf("value is not a string or a fmt.Stringer")
	}
	return qt.BadCheckf("pattern is not a *regexp.Regexp")
}
