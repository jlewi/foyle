package fnames

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/jlewi/foyle/app/pkg/agent"
)

func GetFunctionNameFromFunc(f interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()

	// GoLang adds a -fm suffix to the function name when dealing with a method that
	// has been converted to a function value. So we trim it to match what we get at runtime.
	name = strings.TrimSuffix(name, "-fm")
	return name
}

func Test_Names(t *testing.T) {
	type testCases struct {
		expected string
		input    interface{}
	}

	cases := []testCases{
		{
			expected: LogEvents,
			input:    (&agent.Agent{}).LogEvents,
		},
		{
			expected: StreamGenerate,
			input:    (&agent.Agent{}).StreamGenerate,
		},
	}

	for _, c := range cases {
		if got := GetFunctionNameFromFunc(c.input); got != c.expected {
			t.Errorf("Expected %s, but got %s", c.expected, got)
		}
	}
}
