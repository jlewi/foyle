package oai

import (
	"context"

	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/llms"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"

	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
)

func NewVectorizer(client *openai.Client) *Vectorizer {
	return &Vectorizer{
		client: client,
	}
}

type Vectorizer struct {
	client *openai.Client
}

func (v *Vectorizer) Embed(ctx context.Context, blocks []*v1alpha1.Block) (llms.Vector, error) {
	text := docs.BlocksToMarkdown(blocks)

	// Compute the embedding for the query.
	log := logs.FromContext(ctx)
	log.Info("RAG Query", "query", text)
	request := openai.EmbeddingRequestStrings{
		Input:          []string{text},
		Model:          openai.SmallEmbedding3,
		User:           "",
		EncodingFormat: "float",
	}

	// N.B. regarding retries. We should already be doing retries in the HTTP client.
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

	return resp.Data[0].Embedding, nil
}

func (v *Vectorizer) Length() int {
	return SmallEmbeddingsDims
}
