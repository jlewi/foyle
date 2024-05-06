package eval

import (
	"errors"
	"math"
	"strings"

	"github.com/agnivade/levenshtein"
	"github.com/jlewi/foyle/app/pkg/executor"
)

const (
	alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type command struct {
	unnamed []string
	named   map[string]string
}

type DistanceResult struct {
	Distance   int
	Max        int
	Normalized float32
}

// Distance computes the distance between two instructions
//
// For details refer to tn003_learning_eval.md.
func Distance(left executor.Instruction, right executor.Instruction) (DistanceResult, error) {
	// Split each instruction into named and unnamed arguments
	leftArgs := splitInstruction(left)
	rightArgs := splitInstruction(right)

	result := DistanceResult{
		Distance:   -1,
		Max:        -1,
		Normalized: -1,
	}

	// Compute the distance of the unnamed arguments
	unamedDistance, err := editDistance(leftArgs.unnamed, rightArgs.unnamed)
	if err != nil {
		return result, err
	}

	// Compute the distance of the named arguments
	namedDistance := dictDistance(leftArgs.named, rightArgs.named)

	totalDistance := unamedDistance + namedDistance

	result.Distance = totalDistance

	// Compute the max distance.
	// For the unnamed arguments the maximum distance is the length of which ever command is longer
	// For the named arguments the maximum distance is the number of unique keys in the dictionaries.
	max := int(math.Max(float64(len(leftArgs.unnamed)), float64(len(rightArgs.unnamed))))

	// Need to count the number of unique keys in the dictionaries.
	unique := map[string]string{}

	for k := range leftArgs.named {
		unique[k] = ""
	}
	for k := range rightArgs.named {
		unique[k] = ""
	}

	max += len(unique)
	normalizedDistance := float32(totalDistance) / float32(max)

	result.Max = max
	result.Normalized = normalizedDistance
	return result, nil
}

// editDistance computes the edit distance between two slices of strings.
// Each string in the slice is considered a single token.
func editDistance(left []string, right []string) (int, error) {
	// Our levenstein distance function operates on strings.
	// So we need to map our tokens to single character strings.
	// We do this by building up a dictionary mapping the tokens to single character strings.
	// We currently use a fixed alphabet of 62 characters. We should be able to easily extend this to 100s or thousands
	// of characters because our levenstein library works with UTF8. Just using 1 byte we should be able to represent
	// 128 characters. I wanted to keep it to printable characters. Seems unlikely we will have commands
	// of more than 62 tokens.

	t := tokenizer{
		index: 0,
		dict:  map[string]string{},
	}
	leftVal, err := t.tokenize(left)
	if err != nil {
		return 0, err
	}

	rightVal, err := t.tokenize(right)
	if err != nil {
		return 0, err
	}
	// I picked this particular library because
	// Its code was readable and pretty compact.
	// https://github.com/ka-weihe/fast-levenshtein claims to be 15 times faster but its code is unreadable because
	// its so heavily optimized. Its also not thread safe although it would be trivial to make it so.
	return levenshtein.ComputeDistance(leftVal, rightVal), nil
}

type tokenizer struct {
	index int
	dict  map[string]string
}

func (t *tokenizer) tokenize(vals []string) (string, error) {
	result := ""
	for _, l := range vals {
		if _, ok := t.dict[l]; !ok {
			t.dict[l] = string(alphabet[t.index])
			t.index++
			if t.index >= len(alphabet) {
				return "", errors.New("Too many tokens")
			}
		}
		result += t.dict[l]
	}
	return result, nil
}

// dictDistance computes the distance between two dictionaries.
func dictDistance(left map[string]string, right map[string]string) int {
	// Each key in one dictionary but not the other contributes 1 to the distance.
	distance := 0
	distance += countKeysNotInRight(left, right)
	distance += countKeysNotInRight(right, left)

	// Now we need to check the values of the keys that are in both dictionaries.
	// If the values don't match then we need to add 1 to the distance.
	for k := range left {
		if _, ok := right[k]; !ok {
			continue
		}

		if left[k] != right[k] {
			distance += 1
		}
	}

	return distance
}

func countKeysNotInRight(left map[string]string, right map[string]string) int {
	d := 0
	for k := range left {
		if _, ok := right[k]; !ok {
			d += 1
		}
	}
	return d
}

func splitInstruction(instruction executor.Instruction) command {
	c := command{
		unnamed: []string{instruction.Command.Name},
		named:   map[string]string{},
	}

	for _, arg := range instruction.Command.Args {
		if strings.HasPrefix(arg, "--") {
			pieces := strings.Split(arg, "=")
			if len(pieces) >= 2 {
				c.named[pieces[0]] = strings.Join(pieces[1:], "=")
				continue
			}
		}
		c.unnamed = append(c.unnamed, arg)
	}

	return c
}
