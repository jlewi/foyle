package learn

import (
	"context"
	"fmt"
	"github.com/jlewi/foyle/app/pkg/dbutil"
	"k8s.io/client-go/util/workqueue"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

// Learner handles the learn loop to learn from past mistakes.
//
// TODO(jeremy): Should we call this a trainer?
type Learner struct {
	Config   config.Config
	client   *openai.Client
	blocksDB *dbutil.LockingDB[*logspb.BlockLog]
	queue    workqueue.DelayingInterface

	eventLoopIsDone sync.WaitGroup
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
	}, nil
}

// Start starts a worker thread to asynchronously handle blocks enqueued via the Enqueue function
// Function is non-blocking
func (l *Learner) Start(ctx context.Context) error {
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
			// TODO(jeremy): How do we send a notification to InMemoryDB that the example has been learned and we need
			// to do an update?
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
		return nil
	}

	if b.GeneratedBlock == nil {
		// Block wasn't the result of AI generation
		return nil
	}

	if b.EvalMode {
		log.V(logs.Debug).Info("Skipping block which was created as part of an eval", "id", b.GetId())
		return nil
	}

	// TODO(jeremy): Should we use some sort of distance metric? e.g. edit distance? We could potentially
	// Use the metric used for eval.
	if strings.TrimSpace(b.ExecutedBlock.GetContents()) == strings.TrimSpace(b.GeneratedBlock.GetContents()) {
		log.V(logs.Debug).Info("Skipping executed block which matches generated block", "id", b.GetId())
		return nil
	}

	expectedFile := l.getExampleFile(b.GetId())

	if _, err := os.Stat(expectedFile); err == nil {
		log.V(logs.Debug).Info("File for block exists", "id", b.GetId())
		return nil
	}

	log.Info("Found new training example", "blockId", b.GetId())

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

	if err := os.WriteFile(expectedFile, encoded, 0777); err != nil {
		log.Error(err, "Failed to serialize doc", "id", b.GetId())
		return errors.Wrapf(err, "Failed to write example file doc %s for id %s", expectedFile, b.GetId())
	}

	return nil
}

// Backfill reconciles all blocks in the database
//func (l *Learner) Backfill(ctx context.Context, blocksDB *pebble.DB) error {
//	log := logs.FromContext(ctx)
//
//	allErrors := &helpers.ListOfErrors{}
//
//	iter, err := blocksDB.NewIterWithContext(ctx, nil)
//	if err != nil {
//		return err
//	}
//	defer iter.Close()
//
//	for iter.First(); iter.Valid(); iter.Next() {
//		key := iter.Key()
//		if key == nil {
//			break
//		}
//
//		value, err := iter.ValueAndErr()
//		if err != nil {
//			return errors.Wrapf(err, "Failed to read value for key %s", string(key))
//		}
//
//		b := &logspb.BlockLog{}
//		if err := proto.Unmarshal(value, b); err != nil {
//			allErrors.AddCause(errors.Wrapf(err, "Failed to read block %s", string(key)))
//			continue
//		}
//
//		if b.ExecutedBlock == nil {
//			// Skip unexecuted block
//			continue
//		}
//
//		if b.GeneratedBlock == nil {
//			// Block wasn't the result of AI generation
//			continue
//		}
//
//		if b.EvalMode {
//			log.V(logs.Debug).Info("Skipping block which was created as part of an eval", "id", b.GetId())
//			continue
//		}
//
//		// TODO(jeremy): Should we use some sort of distance metric? e.g. edit distance? We could potentially
//		// Use the metric used for eval.
//		if strings.TrimSpace(b.ExecutedBlock.GetContents()) == strings.TrimSpace(b.GeneratedBlock.GetContents()) {
//			log.V(logs.Debug).Info("Skipping executed block which matches generated block", "id", b.GetId())
//			continue
//		}
//
//		expectedFile := l.getExampleFile(b.GetId())
//
//		if _, err := os.Stat(expectedFile); err == nil {
//			log.V(logs.Debug).Info("File for block exists", "id", b.GetId())
//			continue
//		}
//
//		log.Info("Found new training example", "blockId", b.GetId())
//
//		// TODO(jeremy): Should we take into account execution status when looking for mistakes?
//
//		// Deep copy the original message
//		newDoc := proto.Clone(b.Doc).(*v1alpha1.Doc)
//		newBlock := proto.Clone(b.ExecutedBlock).(*v1alpha1.Block)
//		answer := []*v1alpha1.Block{newBlock}
//
//		example := &v1alpha1.Example{
//			Id:     b.GetId(),
//			Query:  newDoc,
//			Answer: answer,
//		}
//
//		encoded, err := proto.Marshal(example)
//		if err != nil {
//			log.Error(err, "Failed to serialize doc", "id", b.GetId())
//			allErrors.AddCause(err)
//			continue
//		}
//
//		if err := os.WriteFile(expectedFile, encoded, 0777); err != nil {
//			log.Error(err, "Failed to serialize doc", "id", b.GetId())
//			allErrors.AddCause(err)
//			continue
//		}
//	}
//
//	if len(allErrors.Causes) > 0 {
//		return allErrors
//	}
//	return nil
//}

func (l *Learner) getExampleFile(id string) string {
	return filepath.Join(l.Config.GetTrainingDir(), fmt.Sprintf("%s%s", id, fileSuffix))
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
