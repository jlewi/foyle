package eval

import (
	"github.com/go-cmd/cmd"
	"github.com/google/go-cmp/cmp"
	"github.com/jlewi/foyle/app/pkg/executor"
	"testing"
)

func Test_SplitInstruction(t *testing.T) {
	type testCase struct {
		name     string
		args     []string
		expected command
	}

	cases := []testCase{
		{
			name: "simple",
			args: []string{"gcloud", "-p", "acme", "--foo=bar", "baz"},
			expected: command{
				unnamed: []string{"gcloud", "-p", "acme", "baz"},
				named: map[string]string{
					"--foo": "bar",
				},
			},
		},
		{
			name: "equals-in-value",
			args: []string{"foyle", "config", "--foo=bar=baz"},
			expected: command{
				unnamed: []string{"foyle", "config"},
				named: map[string]string{
					"--foo": "bar=baz",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			instruction := executor.Instruction{
				Command: cmd.NewCmd(c.args[0], c.args[1:]...),
			}
			actual := splitInstruction(instruction)
			if d := cmp.Diff(c.expected, actual, cmp.AllowUnexported(command{})); d != "" {
				t.Errorf("Unexpected result (-want +got):\n%s", d)
			}
		})
	}
}

func Test_editDistance(t *testing.T) {
	type testCase struct {
		name     string
		left     []string
		right    []string
		expected int
	}

	cases := []testCase{
		{
			name:     "simple",
			left:     []string{"gcloud", "-p", "acme", "baz"},
			right:    []string{"gcloud", "-p", "acme", "baz"},
			expected: 0,
		},
		{
			name:     "simple",
			left:     []string{"-p", "acme", "baz", "extra"},
			right:    []string{"gcloud", "-p", "acme", "baz"},
			expected: 2,
		},
		{
			name:     "substitution",
			left:     []string{"acme", "foo", "baz"},
			right:    []string{"acme", "bar", "baz"},
			expected: 1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := editDistance(c.left, c.right)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if actual != c.expected {
				t.Errorf("Expected %d but got %d", c.expected, actual)
			}
		})
	}
}
