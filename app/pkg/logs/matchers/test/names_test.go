// package test is a hacky way to avoid circular imports in the test.
// The test imports some packages (e.g. anthropic/oai) that also import matchers
// so if we don't use a separate package we end up with a circular import.
package test

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/jlewi/foyle/app/pkg/logs/matchers"

	"github.com/jlewi/foyle/app/pkg/anthropic"

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
			expected: matchers.LogEvents,
			input:    (&agent.Agent{}).LogEvents,
		},

		{
			expected: matchers.StreamGenerate,
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
		Matcher  matchers.Matcher
		input    interface{}
		expected bool
	}

	cases := []testCases{
		{
			input:    (&oai.Completer{}).Complete,
			Matcher:  matchers.IsOAIComplete,
			name:     "IsOAIComplete",
			expected: true,
		},
		{
			input:    (&anthropic.Completer{}).Complete,
			Matcher:  matchers.IsAnthropicComplete,
			name:     "IsAnthropicComplete",
			expected: true,
		},
		{
			input:    (&agent.Agent{}).Generate,
			Matcher:  matchers.IsGenerate,
			name:     "IsGenerate",
			expected: true,
		},
		{
			input:    (&agent.Agent{}).StreamGenerate,
			Matcher:  matchers.IsStreamGenerate,
			name:     "IsStreamGenerate",
			expected: true,
		},
		{
			input:    logs.LogLLMUsage,
			Matcher:  matchers.IsLLMUsage,
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
