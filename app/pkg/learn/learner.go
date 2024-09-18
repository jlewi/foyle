package learn

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/jlewi/monogo/files"
	"github.com/jlewi/monogo/helpers"

	"github.com/jlewi/foyle/app/pkg/dbutil"
	"k8s.io/client-go/util/workqueue"

	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/proto"
)

const (
	fileSuffix = ".example.binpb"
)

var (
	enqueuedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "learner_enqueued_total",
		Help: "Total number of enqueued blocks",
	})

	cellsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "learner_blocks_processed",
			Help: "Number of blocks processed by the learner",
		},
		[]string{"status"},
	)
)

// Learner handles the learn loop to learn from past mistakes.
//
// TODO(jeremy): Should we call this a trainer?
type Learner struct {
	Config          config.Config
	client          *openai.Client
	blocksDB        *dbutil.LockingDB[*logspb.BlockLog]
	queue           workqueue.DelayingInterface
	postFunc        PostLearnEvent
	eventLoopIsDone sync.WaitGroup
	factory         *files.Factory
}

func NewLearner(cfg config.Config, client *openai.Client, blocksDB *dbutil.LockingDB[*logspb.BlockLog]) (*Learner, error) {
	if client == nil {
		return nil, errors.New("OpenAI client is required")
	}
	return &Learner{
		Config:   cfg,
		client:   client,
		blocksDB: blocksDB,
		queue:    workqueue.NewDelayingQueue(),
		factory:  &files.Factory{},
	}, nil
}

// PostLearnEvent interface for functions to post events about new examples.
type PostLearnEvent func(exampleFile string) error

// Start starts a worker thread to asynchronously handle blocks enqueued via the Enqueue function
// Function is non-blocking
func (l *Learner) Start(ctx context.Context, postFunc PostLearnEvent) error {
	l.postFunc = postFunc
	l.eventLoopIsDone.Add(1)
	go l.eventLoop(ctx)
	return nil
}

// Enqueue adds an example id to be reconciled
func (l *Learner) Enqueue(id string) error {
	if l.queue.ShuttingDown() {
		return errors.New("Queue is shutting down; can't enqueue anymore items")
	}
	l.queue.Add(id)
	enqueuedCounter.Inc()
	return nil
}

func (l *Learner) eventLoop(ctx context.Context) {
	log := logs.FromContext(ctx)
	defer l.eventLoopIsDone.Done()
	for {
		item, shutdown := l.queue.Get()
		if shutdown {
			return
		}
		func() {
			defer l.queue.Done(item)
			exampleId, ok := item.(string)
			if !ok {
				log.Error(errors.New("Failed to cast item to string"), "Failed to cast item to string", "item", item)
				return
			}

			if err := l.Reconcile(ctx, exampleId); err != nil {
				log.Error(err, "Error learning from example", "example", exampleId)
				// Requeue the item so we will try again.
				// TODO(jeremy): should we use a rate limiting queue so we eventually give up?
				l.queue.AddAfter(exampleId, 30*time.Second)
				return
			}
		}()
	}
}

func (l *Learner) Shutdown(ctx context.Context) error {
	log := logs.FromContext(ctx)

	log.Info("Shutting down learner")

	// Shutdown the queues
	l.queue.ShutDown()

	// Wait for the eventloop to finish
	l.eventLoopIsDone.Wait()

	log.Info("Learner shutdown")
	return nil
}

// Reconcile learns from the block with the given id
func (l *Learner) Reconcile(ctx context.Context, id string) error {
	log := logs.FromContext(ctx)

	b, err := l.blocksDB.Get(id)
	if err != nil {
		return errors.Wrapf(err, "Failed to retrieve block %s", id)
	}

	if b.ExecutedBlock == nil {
		// Skip unexecuted block
		cellsProcessed.WithLabelValues("unexecuted").Inc()
		return nil
	}

	if b.GeneratedBlock == nil {
		// Block wasn't the result of AI generation
		cellsProcessed.WithLabelValues("notgenerated").Inc()
		return nil
	}

	if b.EvalMode {
		log.V(logs.Debug).Info("Skipping block which was created as part of an eval", "id", b.GetId())
		cellsProcessed.WithLabelValues("eval").Inc()
		return nil
	}

	// TODO(jeremy): Should we use some sort of distance metric? e.g. edit distance? We could potentially
	// Use the metric used for eval.
	if strings.TrimSpace(b.ExecutedBlock.GetContents()) == strings.TrimSpace(b.GeneratedBlock.GetContents()) {
		log.V(logs.Debug).Info("Skipping executed block which matches generated block", "id", b.GetId())
		cellsProcessed.WithLabelValues("nochange").Inc()
		return nil
	}

	cellsProcessed.WithLabelValues("learn").Inc()
	expectedFiles := l.getExampleFiles(b.GetId())

	log.Info("Found new training example", "blockId", b.GetId())

	if len(expectedFiles) == 0 {
		cellsProcessed.WithLabelValues("noExampleFiles").Inc()
		log.Error(err, "No training files found", "id", b.GetId())
		return errors.Wrapf(err, "No training files found for example %s", b.GetId())
	}

	// TODO(jeremy): Should we take into account execution status when looking for mistakes?

	// Deep copy the original message
	newDoc := proto.Clone(b.Doc).(*v1alpha1.Doc)
	newBlock := proto.Clone(b.ExecutedBlock).(*v1alpha1.Block)
	answer := []*v1alpha1.Block{newBlock}

	example := &v1alpha1.Example{
		Id:     b.GetId(),
		Query:  newDoc,
		Answer: answer,
	}

	if err := l.computeEmbeddings(ctx, example); err != nil {
		return errors.Wrapf(err, "Failed to compute embeddings for example %s", b.GetId())
	}

	encoded, err := proto.Marshal(example)
	if err != nil {
		log.Error(err, "Failed to serialize doc", "id", b.GetId())
		return errors.Wrapf(err, "Failed to serialize doc %s", b.GetId())
	}

	writeErrors := &helpers.ListOfErrors{}
	posted := false
	// An example can be saved in multiple locations.
	// This supports sharing by allowing examples to be written to a shared bucket.
	for _, expectedFile := range expectedFiles {
		writeErr := func() error {
			helper, err := l.factory.Get(expectedFile)
			if err != nil {
				return err
			}
			w, err := helper.NewWriter(expectedFile)
			if err != nil {
				return errors.Wrapf(err, "Failed to create writer for example %s; to file %s", b.GetId(), expectedFile)
			}
			if closer, ok := w.(io.Closer); ok {
				defer closer.Close()
			}

			if _, err := w.Write(encoded); err != nil {
				return errors.Wrapf(err, "Failed to write example %s; to file %s", b.GetId(), expectedFile)
			}
			return nil
		}()
		if writeErr != nil {
			// We need to log the individual error here so that its stack trace gets logged
			log.Error(err, "Failed to write example", "id", b.GetId(), "file", expectedFile)
			writeErrors.AddCause(writeErr)
			continue
		}
		// All post a single file because we don't need to read it multiple times
		if !posted && l.postFunc != nil {
			if err := l.postFunc(expectedFile); err != nil {
				return errors.Wrapf(err, "Failed to post learn event for example %s", b.GetId())
			}
			posted = true
		}
	}

	if len(writeErrors.Causes) > 0 {
		writeErrors.Final = errors.New("Not all examples could be successfully reconciled")
		return writeErrors
	}
	return nil
}

func (l *Learner) getExampleFiles(id string) []string {
	log := logs.FromContext(context.Background())
	paths := make([]string, 0)
	for _, d := range l.Config.GetTrainingDirs() {
		h, err := l.factory.GetDirHelper(d)
		if err != nil {
			log.Error(err, "Unable to DirHelper", "dir", d)
			continue
		}
		paths = append(paths, h.Join(d, fmt.Sprintf("%s%s", id, fileSuffix)))
	}

	return paths
}

func (l *Learner) computeEmbeddings(ctx context.Context, example *v1alpha1.Example) error {
	log := logs.FromContext(ctx)
	if example.Embedding != nil {
		log.V(logs.Debug).Info("Embedding already exists", "id", example.Id)
		// Skip if we already have an embedding
		return nil
	}

	query := docs.DocToMarkdown(example.Query)

	request := openai.EmbeddingRequestStrings{
		Input:          []string{query},
		Model:          openai.SmallEmbedding3,
		User:           "",
		EncodingFormat: "float",
	}
	resp, err := l.client.CreateEmbeddings(ctx, request)
	if err != nil {
		log.Error(err, "Failed to create embeddings", "id", example.Id, "query", query)
		return errors.Wrapf(err, "Failed to create embeddings")
	}

	if len(resp.Data) != 1 {
		log.Error(err, "Expected exactly 1 embedding", "id", example.Id, "query", query, "got", len(resp.Data))
		return errors.Errorf("Expected exactly 1 embedding but got %d", len(resp.Data))
	}

	if len(resp.Data[0].Embedding) != oai.SmallEmbeddingsDims {
		log.Error(err, "Embeddings have wrong dimension", "id", example.Id, "query", query, "got", len(resp.Data[0].Embedding), "want", oai.SmallEmbeddingsDims)
		return errors.Wrapf(err, "Embeddings have wrong dimension; got %v, want %v", len(resp.Data[0].Embedding), oai.SmallEmbeddingsDims)
	}

	example.Embedding = resp.Data[0].Embedding
	return nil
}
