package oai

import (
	"context"

	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"gonum.org/v1/gonum/mat"
)

func NewVectorizer(client *openai.Client) *Vectorizer {
	return &Vectorizer{
		client: client,
	}
}

type Vectorizer struct {
	client *openai.Client
}

func (v *Vectorizer) Embed(ctx context.Context, text string) (*mat.VecDense, error) {
	log := logs.FromContext(ctx)
	log.Info("RAG Query", "query", text)
	request := openai.EmbeddingRequestStrings{
		Input:          []string{text},
		Model:          openai.SmallEmbedding3,
		User:           "",
		EncodingFormat: "float",
	}

	resp, err := v.client.CreateEmbeddings(ctx, request)
	if err != nil {
		return nil, errors.Errorf("Failed to create embeddings")
	}

	if len(resp.Data) != 1 {
		return nil, errors.Errorf("Expected exactly 1 embedding but got %d", len(resp.Data))
	}

	if len(resp.Data[0].Embedding) != SmallEmbeddingsDims {
		return nil, errors.Errorf("Embeddings have wrong dimension; got %v, want %v", len(resp.Data[0].Embedding), SmallEmbeddingsDims)
	}

	// Compute the cosine similarity between the query and each example.
	qVec := mat.NewVecDense(SmallEmbeddingsDims, nil)
	for i := 0; i < SmallEmbeddingsDims; i++ {
		qVec.SetVec(i, float64(resp.Data[0].Embedding[i]))
	}
	return qVec, nil
}

func (v *Vectorizer) Length() int {
	return SmallEmbeddingsDims
}
