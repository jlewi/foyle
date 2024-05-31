package learn

import (
	"context"
	"os"
	"path/filepath"
	"sort"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"gonum.org/v1/gonum/mat"
	"google.golang.org/protobuf/proto"
)

// InMemoryExampleDB is an in-memory example database.
// It uses brute force to retrieve examples.
type InMemoryExampleDB struct {
	config config.Config
	client *openai.Client
	// examples stores the examples with the embeddings zeroed out.
	examples []*v1alpha1.Example

	// embeddings stores the embeddings for the examples.
	// this is num_examples x num_features.
	// embeddings[i, :] is the embedding for examples[i]
	// Each row is 1 vector because data is stored in row major order in the backing array.
	// So this way all the elements of a vector are next to each other
	embeddings *mat.Dense
}

func NewInMemoryExampleDB(cfg config.Config, client *openai.Client) (*InMemoryExampleDB, error) {
	if client == nil {
		return nil, errors.New("OpenAI client is required")
	}
	db := &InMemoryExampleDB{
		config: cfg,
		client: client,
	}

	if err := db.loadExamples(context.Background()); err != nil {
		return nil, errors.Wrap(err, "Failed to load examples")
	}
	return db, nil
}

func (db *InMemoryExampleDB) GetExamples(ctx context.Context, doc *v1alpha1.Doc, maxResults int) ([]*v1alpha1.Example, error) {
	log := logs.FromContext(ctx)
	query := docs.DocToMarkdown(doc)

	if len(db.examples) == 0 {
		// TODO(jeremy): What should we do in this case?
		return nil, errors.New("No examples available")
	}

	log.Info("RAG Query", "query", query)
	request := openai.EmbeddingRequestStrings{
		Input:          []string{query},
		Model:          openai.SmallEmbedding3,
		User:           "",
		EncodingFormat: "float",
	}
	resp, err := db.client.CreateEmbeddings(ctx, request)
	if err != nil {
		return nil, errors.Errorf("Failed to create embeddings")
	}

	if len(resp.Data) != 1 {
		return nil, errors.Errorf("Expected exactly 1 embedding but got %d", len(resp.Data))
	}

	if len(resp.Data[0].Embedding) != oai.SmallEmbeddingsDims {
		return nil, errors.Errorf("Embeddings have wrong dimension; got %v, want %v", len(resp.Data[0].Embedding), oai.SmallEmbeddingsDims)
	}

	// Compute the cosine similarity between the query and each example.
	qVec := mat.NewVecDense(oai.SmallEmbeddingsDims, nil)
	for i := 0; i < oai.SmallEmbeddingsDims; i++ {
		qVec.SetVec(i, float64(resp.Data[0].Embedding[i]))
	}

	// Create a new vector to store the result
	result := mat.NewVecDense(len(db.examples), nil)

	// Multiply the matrix by the vector
	result.MulVec(db.embeddings, qVec)

	sorted := sortIndexes(result)

	numResults := maxResults
	if len(db.examples) < maxResults {
		numResults = len(db.examples)
	}

	results := make([]*v1alpha1.Example, 0, numResults)

	for i := len(sorted) - numResults; i < len(sorted); i++ {
		example := db.examples[sorted[i]]
		score := result.AtVec(sorted[i])
		log.Info("RAG result", zap.Object("example", example), "score", score)
		results = append(results, example)
	}

	return results, nil
}

// sortIndexes returns the indexes of the vector sorted in ascending order
func sortIndexes(v mat.Vector) (indexes []int) {
	indexes = make([]int, v.Len())
	for i := 0; i < v.Len(); i++ {
		indexes[i] = i
	}

	sort.Slice(indexes, func(i, j int) bool {
		return v.AtVec(indexes[i]) < v.AtVec(indexes[j])
	})
	return indexes
}

func (db *InMemoryExampleDB) loadExamples(ctx context.Context) error {
	log := logs.FromContext(ctx)
	glob := filepath.Join(db.config.GetTrainingDir(), "*"+fileSuffix)
	matches, err := filepath.Glob(glob)
	if err != nil {
		return errors.Wrapf(err, "Failed to match glob %s", glob)
	}

	// TODO(jeremy): How should we handle there being no examples?
	if len(matches) == 0 {
		return nil
	}

	db.examples = make([]*v1alpha1.Example, 0, len(matches))
	embeddings := mat.NewDense(len(matches), oai.SmallEmbeddingsDims, nil)

	// Load the examples.
	for row, match := range matches {
		log.V(logs.Debug).Info("Loading example", "file", match)
		// TODO(jeremy): We should make this resilient to some files failing to load. We might need to think
		// about how we deal with allocating the array when we don't know the number of examples to load.
		raw, err := os.ReadFile(match)
		if err != nil {
			return errors.Wrapf(err, "Failed to read file %s", match)
		}
		example := &v1alpha1.Example{}
		if err := proto.Unmarshal(raw, example); err != nil {
			return errors.Wrapf(err, "Failed to unmarshal example from %s", match)
		}

		if len(example.Embedding) != oai.SmallEmbeddingsDims {
			return errors.Errorf("Expected embedding to have %d elements but got %d", oai.SmallEmbeddingsDims, len(example.Embedding))
		}
		for col := 0; col < len(example.Embedding); col++ {
			embeddings.Set(row, col, float64(example.Embedding[col]))
		}

		// Zero out the embedding because we don't want to store it in two places
		example.Embedding = nil
		db.examples = append(db.examples, example)
	}

	db.embeddings = embeddings
	return nil
}
