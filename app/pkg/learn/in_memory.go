package learn

import (
	"context"
	"io"
	"sort"
	"sync"

	"github.com/jlewi/foyle/app/pkg/docs"

	"github.com/jlewi/foyle/app/pkg/llms"

	"github.com/jlewi/monogo/files"

	"k8s.io/client-go/util/workqueue"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gonum.org/v1/gonum/mat"
	"google.golang.org/protobuf/proto"
)

// InMemoryExampleDB is an in-memory example database.
// It uses brute force to retrieve examples.
type InMemoryExampleDB struct {
	config     config.Config
	vectorizer llms.Vectorizer
	// examples stores the examples with the embeddings zeroed out.
	examples []*v1alpha1.Example

	// idToRow maps the example ID to the row in the embeddings matrix.
	idToRow map[string]int

	// embeddings stores the embeddings for the examples.
	// this is num_examples x num_features.
	// embeddings[i, :] is the embedding for examples[i]
	// Each row is 1 vector because data is stored in row major order in the backing array.
	// So this way all the elements of a vector are next to each other
	embeddings *mat.Dense

	// A queue of examples to process
	q workqueue.DelayingInterface

	// loaderDone is used to signal when the loader is done
	eventLoopDone sync.WaitGroup

	// mu protects the examples and embeddings fields so that we can update it safely.
	lock sync.RWMutex

	factory *files.Factory
}

func NewInMemoryExampleDB(cfg config.Config, vectorizer llms.Vectorizer) (*InMemoryExampleDB, error) {
	if vectorizer == nil {
		return nil, errors.New("Vectorizer client is required")
	}
	db := &InMemoryExampleDB{
		config:     cfg,
		vectorizer: vectorizer,
		idToRow:    make(map[string]int),
		q:          workqueue.NewDelayingQueue(),
		factory:    &files.Factory{},
	}

	if err := db.loadExamples(context.Background()); err != nil {
		return nil, errors.Wrap(err, "Failed to load examples")
	}
	return db, nil
}

func (db *InMemoryExampleDB) GetExamples(ctx context.Context, req *v1alpha1.GenerateRequest, maxResults int) ([]*v1alpha1.Example, error) {
	log := logs.FromContext(ctx)

	if len(db.examples) == 0 {
		// Since there are no examples just return an empty list
		return []*v1alpha1.Example{}, nil
	}

	blocks, err := docs.CreateQuery(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create query")
	}

	// Compute the embedding for the query.
	qVecData, err := db.vectorizer.Embed(ctx, blocks)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to compute embedding for query")
	}

	qVec := llms.VectorToVecDense(qVecData)
	// Acquire a lock on the data so we can safely read it.
	db.lock.RLock()
	defer db.lock.RUnlock()

	// Create a new vector to store the result
	numExamples := len(db.examples)
	numRows, _ := db.embeddings.Dims()
	result := mat.NewVecDense(numRows, nil)

	// Multiply the matrix by the vector
	result.MulVec(db.embeddings, qVec)

	// only the 0:len(db.examples) row of embeddings are valid so we need to trim the indexes
	sorted := sortIndexes(result, numExamples)

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

func (db *InMemoryExampleDB) GetExample(ctx context.Context, id string) (*v1alpha1.Example, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	row, ok := db.idToRow[id]
	if !ok {
		return nil, errors.Errorf("Example with id %s not found", id)
	}

	return db.examples[row], nil
}

// Start starts the event loop to process enqueued examples
func (db *InMemoryExampleDB) Start(ctx context.Context) error {
	db.eventLoopDone.Add(1)
	go db.eventLoop(ctx)
	return nil
}

// EnqueueExample enqueues the exampleFile to be loaded. This could be a new example or an existing example
func (db *InMemoryExampleDB) EnqueueExample(exampleFile string) error {
	if db.q.ShuttingDown() {
		return errors.New("Queue is shutting down; can't enqueue any more items")
	}
	db.q.Add(exampleFile)
	return nil
}

func (db *InMemoryExampleDB) eventLoop(ctx context.Context) {
	log := logs.FromContext(ctx)
	defer db.eventLoopDone.Done()

	for {
		obj, shutdown := db.q.Get()
		if shutdown {
			return
		}
		func() {
			defer db.q.Done(obj)

			exampleFile := obj.(string)
			if err := db.loadRow(ctx, exampleFile); err != nil {
				log.Error(err, "Failed to load example", "uri", exampleFile)
			}
		}()
	}
}

// sortIndexes returns the indexes of the vector[0:dim] sorted in ascending order
func sortIndexes(v mat.Vector, dim int) (indexes []int) {
	log := logs.FromContext(context.Background())
	if v.Len() < dim {
		log.Error(errors.New("Vector is too small"), "Vector is too small; will truncate results to vector size", "len", v.Len(), "dim", dim)
		dim = v.Len()
	}
	indexes = make([]int, dim)
	for i := 0; i < dim; i++ {
		indexes[i] = i
	}

	sort.Slice(indexes, func(i, j int) bool {
		return v.AtVec(indexes[i]) < v.AtVec(indexes[j])
	})
	return indexes
}

func (db *InMemoryExampleDB) loadExamples(ctx context.Context) error {
	log := logs.FromContext(ctx)

	stores := db.config.GetTrainingDirs()

	for _, s := range stores {
		filehelper, err := db.factory.GetDirHelper(s)
		if err != nil {
			return errors.Wrapf(err, "Failed to get file helper for %s", s)
		}
		glob := filehelper.Join(s, "*"+fileSuffix)
		matches, err := filehelper.Glob(glob)
		if err != nil {
			return errors.Wrapf(err, "Failed to match glob %s", glob)
		}

		if db.examples == nil {
			db.examples = make([]*v1alpha1.Example, 0, len(matches))
		}

		if db.embeddings == nil {
			db.embeddings = mat.NewDense(initialNumberOfRows(len(matches)), oai.SmallEmbeddingsDims, nil)
		}

		// Load the examples.
		for _, match := range matches {
			if err := db.loadRow(ctx, match); err != nil {
				// Just keep going
				log.Error(err, "Failed to load example", "file", match)
			}
		}
	}

	return nil
}

func initialNumberOfRows(numExamples int) int {
	// We intentionally initialize an initial matrix which is too small so that during the initial load
	// grow will be triggered. Since we grow by a factor of two we should end up with an overallocated matrix
	// This means that by default the matrix should contain extra rows that haven't been populated with examples
	// yet. This way we can verify that doesn't trip up rag
	size := int(float32(numExamples) / 1.5)
	// If size is < 1 we end up initializing an empty matrix which will cause a panic
	if size < 1 {
		size = 1
	}
	return size
}

func (db *InMemoryExampleDB) Shutdown(ctx context.Context) error {
	log := logs.FromContext(ctx)

	log.Info("Shutting down InMemoryExampleDB")

	// Shutdown the queues
	db.q.ShutDown()

	// Wait for the eventloop to finish
	db.eventLoopDone.Wait()

	log.Info("InMemoryExampleDB shutdown")
	return nil
}

// loadRow loads the example from the specified file into embeddings matrix.
func (db *InMemoryExampleDB) loadRow(ctx context.Context, exampleFile string) error {
	log := logs.FromContext(ctx)
	log.V(logs.Debug).Info("Loading example", "file", exampleFile)

	fileHelper, err := db.factory.Get(exampleFile)
	if err != nil {
		return errors.Wrapf(err, "Failed to get file helper for %s", exampleFile)
	}

	reader, err := fileHelper.NewReader(exampleFile)
	if err != nil {
		return errors.Wrapf(err, "Failed to open file %s", exampleFile)
	}
	raw, err := io.ReadAll(reader)
	if err != nil {
		return errors.Wrapf(err, "Failed to read file %s", exampleFile)
	}
	example := &v1alpha1.Example{}
	if err := proto.Unmarshal(raw, example); err != nil {
		return errors.Wrapf(err, "Failed to unmarshal example from %s", exampleFile)
	}

	if len(example.Embedding) != oai.SmallEmbeddingsDims {
		return errors.Errorf("Expected embedding to have %d elements but got %d", oai.SmallEmbeddingsDims, len(example.Embedding))
	}

	return db.updateExample(example)
}

// updateExample adds or updates the example in the database.
func (db *InMemoryExampleDB) updateExample(example *v1alpha1.Example) error {
	// Acquire an exclusive lock
	db.lock.Lock()
	defer db.lock.Unlock()

	// Check if this example is already in the matrix and if it is we just overwrite it
	row, ok := db.idToRow[example.Id]
	if !ok {
		// Since the example isn't in the matrix we just use the next row which is the length of
		// the examples array.
		row = len(db.examples)
	}

	numRows, _ := db.embeddings.Dims()
	if row >= numRows {
		newMat := db.embeddings.Grow(numRows*2, 0)
		newDense, ok := newMat.(*mat.Dense)
		if !ok {
			return errors.New("Failed to grow matrix; the returned value was not a dense matrix")
		}
		db.embeddings = newDense
	}
	for col := 0; col < len(example.Embedding); col++ {
		db.embeddings.Set(row, col, float64(example.Embedding[col]))
	}

	// Zero out the embedding because we don't want to store it in two places
	example.Embedding = nil
	if row < len(db.examples) {
		db.examples[row] = example
	} else {
		db.examples = append(db.examples, example)
	}
	db.idToRow[example.Id] = row
	return nil
}
