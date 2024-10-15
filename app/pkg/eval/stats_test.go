package eval

import (
	"github.com/google/go-cmp/cmp"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"testing"
)

func Test_computePercentilesInts(t *testing.T) {
	type testCase struct {
		name        string
		data        []int
		percentiles []float64
		expected    []v1alpha1.PercentileStat
	}

	cases := []testCase{
		{
			name:        "Basic",
			data:        []int{2, 1, 4, 3, 5},
			percentiles: []float64{0.5, .8, .99},
			expected: []v1alpha1.PercentileStat{
				{
					Percentile: 0.4,
					Value:      2,
				},
				{
					Percentile: 0.6,
					Value:      3,
				},
				{
					Percentile: 0.8,
					Value:      4,
				},
				{
					Percentile: 1,
					Value:      5,
				},
			},
		},
		{
			name:        "up and down",
			data:        []int{2, 1, 4, 3},
			percentiles: []float64{0.60},
			expected: []v1alpha1.PercentileStat{
				{
					Percentile: 0.5,
					Value:      2,
				},
				{
					Percentile: 0.75,
					Value:      3,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := computePercentilesOfInts(c.data, c.percentiles)
			if err != nil {
				t.Fatalf("Error computing percentiles: %v", err)
			}

			if d := cmp.Diff(c.expected, actual); d != "" {
				t.Fatalf("Unexpected diff between expected and actual percentiles:\n%+v", d)
			}
		})
	}
}
