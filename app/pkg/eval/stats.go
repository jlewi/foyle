package eval

import (
	"math"
	"sort"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
)

// computePercentilesOfInts computes the percentiles of a slice of integers.
// p is a slice of percentiles to compute. Values should be between > 0 and <1.
func computePercentilesOfInts(data []int, p []float64) ([]*v1alpha1.PercentileStat, error) {
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

	percentiles := make([]*v1alpha1.PercentileStat, 0, len(p))
	for _, k := range keys {
		// Compute the Value at the kth index
		percentiles = append(percentiles, &v1alpha1.PercentileStat{
			Percentile: float64(k+1) / float64(len(data)),
			Value:      float64(data[k]),
		})
	}

	return percentiles, nil
}
