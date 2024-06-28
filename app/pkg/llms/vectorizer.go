package llms

import (
	"context"

	"gonum.org/v1/gonum/mat"
)

// Vectorizer computes embedding representations of text.
type Vectorizer interface {
	// Embed computes the embedding of the text
	Embed(ctx context.Context, text string) (*mat.VecDense, error)
	// Length returns the length of the embeddings
	Length() int
}
