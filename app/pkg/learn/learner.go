package learn

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/pebble"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/monogo/helpers"
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
	Config config.Config
	client *openai.Client
}

func NewLearner(cfg config.Config, client *openai.Client) (*Learner, error) {
	if client == nil {
		return nil, errors.New("OpenAI client is required")
	}
	return &Learner{
		Config: cfg,
		client: client,
	}, nil
}

func (l *Learner) Reconcile(ctx context.Context) error {
	// TODO(jeremy): Can we call Analyze to compute the latest logs?
	log := logs.FromContext(ctx)

	trainDir := l.Config.GetTrainingDir()
	if _, err := os.Stat(trainDir); err != nil {
		if os.IsNotExist(err) {
			log.V(logs.Debug).Info("Creating training directory", "dir", trainDir)
			if err := os.MkdirAll(trainDir, 0777); err != nil {
				return errors.Wrap(err, "Failed to create training directory")
			}
		} else {
			return errors.Wrap(err, "Failed to check if training directory exists")
		}
	}

	allErrors := &helpers.ListOfErrors{}

	// Load the blocksdb
	blocksDB, err := pebble.Open(l.Config.GetBlocksDBDir(), &pebble.Options{})
	if err != nil {
		return errors.Wrapf(err, "could not open blocks database %s", l.Config.GetBlocksDBDir())
	}

	if err != nil {
		log.Error(err, "There was a problem loading the latest block logs")
		allErrors.AddCause(errors.New("The latest block locks could not be loaded; check logs for more information"))
	}

	if err := l.reconcileExamples(ctx, blocksDB); err != nil {
		log.Error(err, "There were problems reconciling examples")
		allErrors.AddCause(errors.New("Not all example were reconciled successfully; check logs for more information"))
	}

	if err := l.reconcileEmbeddings(ctx); err != nil {
		log.Error(err, "There were problems reconciling embeddings")
		allErrors.AddCause(errors.New("Not all embedded were reconciled successfully; check logs for more information"))
	}
	return nil
}

// reconcileExamples ensures that an example file exists for mistakes
func (l *Learner) reconcileExamples(ctx context.Context, blocksDB *pebble.DB) error {
	log := logs.FromContext(ctx)

	allErrors := &helpers.ListOfErrors{}

	iter, err := blocksDB.NewIterWithContext(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if key == nil {
			break
		}

		value, err := iter.ValueAndErr()
		if err != nil {
			return errors.Wrapf(err, "Failed to read value for key %s", string(key))
		}

		b := &logspb.BlockLog{}
		if err := proto.Unmarshal(value, b); err != nil {
			allErrors.AddCause(errors.Wrapf(err, "Failed to read block %s", string(key)))
			continue
		}

		if b.ExecutedBlock == nil {
			// Skip unexecuted block
			continue
		}

		if b.GeneratedBlock == nil {
			// Block wasn't the result of AI generation
			continue
		}

		if b.EvalMode {
			log.V(logs.Debug).Info("Skipping block which was created as part of an eval", "id", b.GetId())
			continue
		}

		// TODO(jeremy): Should we use some sort of distance metric? e.g. edit distance? We could potentially
		// Use the metric used for eval.
		if strings.TrimSpace(b.ExecutedBlock.GetContents()) == strings.TrimSpace(b.GeneratedBlock.GetContents()) {
			log.V(logs.Debug).Info("Skipping executed block which matches generated block", "id", b.GetId())
			continue
		}

		expectedFile := l.getExampleFile(b.GetId())

		if _, err := os.Stat(expectedFile); err == nil {
			log.V(logs.Debug).Info("File for block exists", "id", b.GetId())
			continue
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

		encoded, err := proto.Marshal(example)
		if err != nil {
			log.Error(err, "Failed to serialize doc", "id", b.GetId())
			allErrors.AddCause(err)
			continue
		}

		if err := os.WriteFile(expectedFile, encoded, 0777); err != nil {
			log.Error(err, "Failed to serialize doc", "id", b.GetId())
			allErrors.AddCause(err)
			continue
		}
	}

	if len(allErrors.Causes) > 0 {
		return allErrors
	}
	return nil
}

func (l *Learner) getExampleFile(id string) string {
	return filepath.Join(l.Config.GetTrainingDir(), fmt.Sprintf("%s%s", id, fileSuffix))
}

func (l *Learner) reconcileEmbeddings(ctx context.Context) error {
	oLog := logs.FromContext(ctx)
	allErrors := &helpers.ListOfErrors{}

	glob := filepath.Join(l.Config.GetTrainingDir(), "*"+fileSuffix)
	matches, err := filepath.Glob(glob)
	if err != nil {
		return errors.Wrapf(err, "Failed to match glob %s", glob)
	}

	for _, eFile := range matches {
		log := oLog.WithValues("file", eFile)

		rawExample, err := os.ReadFile(eFile)
		if err != nil {
			log.Error(err, "Failed to read file containing doc for block")
			allErrors.AddCause(err)
			continue
		}

		example := &v1alpha1.Example{}
		if err := proto.Unmarshal(rawExample, example); err != nil {
			log.Error(err, "Failed to unmarshal example")
			allErrors.AddCause(err)
			continue
		}

		if example.Embedding != nil {
			log.V(logs.Debug).Info("Embedding already exists", "id", example.Id)
			// Skip if we already have an embedding
			continue
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
			allErrors.AddCause(err)
			continue
		}

		if len(resp.Data) != 1 {
			log.Error(err, "Expected exactly 1 embedding", "id", example.Id, "query", query, "got", len(resp.Data))
			allErrors.AddCause(errors.New("Expected exactly 1 embedding"))
			continue
		}

		if len(resp.Data[0].Embedding) != oai.SmallEmbeddingsDims {
			log.Error(err, "Embeddings have wrong dimension", "id", example.Id, "query", query, "got", len(resp.Data[0].Embedding), "want", oai.SmallEmbeddingsDims)
			allErrors.AddCause(errors.New("Embeddings have wrong dimension"))
			continue
		}

		example.Embedding = resp.Data[0].Embedding

		encoded, err := proto.Marshal(example)
		if err != nil {
			log.Error(err, "Failed to serialize example", "id", example.Id)
			allErrors.AddCause(err)
			continue
		}
		if err := os.WriteFile(eFile, encoded, 0777); err != nil {
			log.Error(err, "Failed to update example file", "id", example.Id)
			allErrors.AddCause(err)
			continue
		}
	}

	if len(allErrors.Causes) > 0 {
		return allErrors
	}
	return nil
}
