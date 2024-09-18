package matchers

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/jlewi/foyle/app/pkg/oai"

	"github.com/jlewi/foyle/app/pkg/agent"
	"github.com/jlewi/foyle/app/pkg/logs"
)

func GetFunctionNameFromFunc(f interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()

	// GoLang adds a -fm suffix to the function name when dealing with a method that
	// has been converted to a function value. So we trim it to match what we get at runtime.
	name = strings.TrimSuffix(name, "-fm")
	return name
}

// TODO(jeremy): We should probably migrate this to matchers.
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

func Test_Matchers(t *testing.T) {
	type testCases struct {
		name     string
		Matcher  Matcher
		input    interface{}
		expected bool
	}

	cases := []testCases{
		{
			input:    (&oai.Completer{}).Complete,
			Matcher:  IsOAIComplete,
			name:     "IsOAIComplete",
			expected: true,
		},
		{
			input:    (&agent.Agent{}).StreamGenerate,
			Matcher:  IsStreamGenerate,
			name:     "IsStreamGenerate",
			expected: true,
		},
		{
			input:    logs.LogLLMUsage,
			Matcher:  IsLLMUsage,
			name:     "IsLLMUsage",
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.Matcher(GetFunctionNameFromFunc(c.input)); got != c.expected {
				t.Errorf("Expected %v, but got %v", c.expected, got)
			}
		})
	}
}
