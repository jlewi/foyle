package eval

import (
	"errors"
	"github.com/agnivade/levenshtein"
	"github.com/jlewi/foyle/app/pkg/executor"
	"strings"
)

const (
	alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type command struct {
	unnamed []string
	named   map[string]string
}

// Distance computes the distance between two instructions
//
// For details refer to tn003_learning_eval.md.
func Distance(left executor.Instruction, right executor.Instruction) (int, error) {
	// Split each instruction into named and unnamed arguments
	leftArgs := splitInstruction(left)
	rightArgs := splitInstruction(right)

	// Compute the distance of the unnamed arguments
	unamedDistance, err := editDistance(leftArgs.unnamed, rightArgs.unnamed)
	return unamedDistance, err
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
