package eval

import (
	"github.com/pkg/errors"
	"math"
	"sort"
)

type IntegerPercentile struct {
	Percentile float64
	Value      int64
}

// computePercentilesOfInts computes the percentiles of a slice of integers.
// p is a slice of percentiles to compute. Values should be between > 0 and <1.
func computePercentilesOfInts(data []int, p []float64) ([]IntegerPercentile, error) {
	sort.Ints(data)

	indexes := map[int]bool{}
	for _, p := range p {
		if p < 0 || p >= 1 {
			return nil, errors.Errorf("Percentile %v is not between 0 and 1", p)
		}

		actual := p*float64(len(data)) - 1
		index := int(math.Floor(actual))
		indexes[index] = true
		if actual != math.Floor(actual) && index+1 < len(data) {
			indexes[index+1] = true
		}
	}

	// Extract keys from the map
	var keys []int
	for k := range indexes {
		keys = append(keys, k)
	}

	// Sort the keys
	sort.Ints(keys)

	percentiles := make([]IntegerPercentile, 0, len(p))
	for _, k := range keys {
		// Compute the Value at the kth index
		percentiles = append(percentiles, IntegerPercentile{
			Percentile: float64(k+1) / float64(len(data)),
			Value:      int64(data[k]),
		})
	}

	return percentiles, nil
}
