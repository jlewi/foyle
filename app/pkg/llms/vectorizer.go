package llms

import (
	"context"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"

	"gonum.org/v1/gonum/mat"
)

type Vector []float32

// Vectorizer computes embedding representations of text.
type Vectorizer interface {
	// Embed computes the embedding of the text
	Embed(ctx context.Context, req *v1alpha1.GenerateRequest) (Vector, error)
	// Length returns the length of the embeddings
	Length() int
}

// VectorToVecDense converts a Vector to a *mat.VecDense
func VectorToVecDense(v Vector) *mat.VecDense {
	// We need to cast from float32 to float64
	qVec := mat.NewVecDense(len(v), nil)
	for i := 0; i < len(v); i++ {
		qVec.SetVec(i, float64(v[i]))
	}
	return qVec
}
