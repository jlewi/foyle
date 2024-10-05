package oai

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
)

func Test_HTTPStatusCode(t *testing.T) {
	type testCase struct {
		name     string
		err      error
		expected int
	}

	cases := []testCase{
		{
			name: "basic",
			err: &openai.APIError{
				HTTPStatusCode: 404,
			},
			expected: 404,
		},
		{
			name: "wrapped",
			err: errors.Wrapf(&openai.APIError{
				HTTPStatusCode: 509,
			}, "wrapped"),
			expected: 509,
		},
		{
			name:     "not api error",
			err:      errors.New("not an api error"),
			expected: -1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := HTTPStatusCode(tc.err)
			if actual != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}
