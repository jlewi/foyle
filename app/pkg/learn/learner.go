package learn

import (
	"context"
	"fmt"
	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/analyze"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"strings"
)

// Learner handles the learn loop to learn from past mistakes.
//
// TODO(jeremy): Should we call this a trainer?
type Learner struct {
	Config *config.Config
	client *openai.Client
}

func NewLearner(cfg *config.Config, client *openai.Client) (*Learner, error) {
	if cfg == nil {
		return nil, errors.New("Config is required")
	}
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

	// Load the blocklogs
	blocks, err := analyze.LoadLatestBlockLogs(ctx, l.Config.GetProcessedLogDir())
	if err != nil {

	}

	if err := l.reconcileExamples(ctx, blocks); err != nil {
		log.Error(err, "There were problems reconciling examples")
		allErrors.AddCause(errors.New("Not all example were reconciled successfully; check logs for more information"))
	}

	return nil
}

// reconcileExamples ensures that an example file exists for mistakes
func (l *Learner) reconcileExamples(ctx context.Context, blocks map[string]api.BlockLog) error {
	log := logs.FromContext(ctx)

	allErrors := &helpers.ListOfErrors{}

	for _, b := range blocks {
		if b.ExecutedBlock == nil {
			// Skip unexecuted block
			continue
		}

		if b.GeneratedBlock == nil {
			// Block wasn't the result of AI generation
			continue
		}
		// TODO(jeremy): Should we use some sort of distance metric? e.g. edit distance?
		if strings.TrimSpace(b.ExecutedBlock.GetContents()) == strings.TrimSpace(b.GeneratedBlock.GetContents()) {
			log.V(logs.Debug).Info("Skipping executed block which matches generated block", "id", b.ID)
			continue
		}

		expectedFile := filepath.Join(l.Config.GetTrainingDir(), fmt.Sprintf("%s.foyle", b.ID))

		_, err := os.Stat(expectedFile)
		if err == nil {
			log.V(logs.Debug).Info("File for block exists", "id", b.ID)
			continue
		}

		// TODO(jeremy): Should we take into account execution status?

		// Deep copy the original message
		newDoc := proto.Clone(b.Doc).(*v1alpha1.Doc)
		newBlock := proto.Clone(b.ExecutedBlock).(*v1alpha1.Block)
		newDoc.Blocks = append(newDoc.Blocks, newBlock)

		// Using jsonpb.Marshaler
		marshaler := &protojson.MarshalOptions{
			Indent: "  ",
		}
		encoded, err := marshaler.Marshal(newDoc)
		if err != nil {
			log.Error(err, "Failed to serialize doc", "id", b.ID)
			allErrors.AddCause(err)
			continue
		}

		if err := os.WriteFile(expectedFile, encoded, 0777); err != nil {
			log.Error(err, "Failed to serialize doc", "id", b.ID)
			allErrors.AddCause(err)
			continue
		}
	}

	if len(allErrors.Causes) > 0 {
		return allErrors
	}
	return nil
}

func (l *Learner) reconcileEmbeddings(ctx context.Context, bids []string) error {
	log := logs.FromContext(ctx)

	allErrors := &helpers.ListOfErrors{}

	for id := range bids {

		expectedFile := filepath.Join(l.Config.GetTrainingDir(), fmt.Sprintf("%s.embeddings.binpb", id))

		_, err := os.Stat(expectedFile)
		if err == nil {
			log.V(logs.Debug).Info("File for block exists", "id", id, "file", expectedFile)
			continue
		}

		inFile := filepath.Join(l.Config.GetTrainingDir(), fmt.Sprintf("%s.foyle", id))
		rawDoc, err := os.ReadFile(inFile)
		if err != nil {
			log.Error(err, "Failed to read file containing doc for block", "file", inFile, "id", id)
			allErrors.AddCause(err)
			continue
		}

		doc := &v1alpha1.Doc{}
		if err := protojson.Unmarshal(rawDoc, doc); err != nil {
			log.Error(err, "Failed to unmarshal doc", "id", id)
			allErrors.AddCause(err)
			continue
		}

		// Get the query which is everything except the last block
		query := docToQuery(doc)

		request := openai.EmbeddingRequestStrings{
			Input:          []string{query},
			Model:          openai.SmallEmbedding3,
			User:           "",
			EncodingFormat: "float",
		}
		resp, err := l.client.CreateEmbeddings(ctx, request)
		if err != nil {
			log.Error(err, "Failed to create embeddings", "id", id, "query", query)
			allErrors.AddCause(err)
			continue
		}

		if len(resp.Data) != 1 {
			log.Error(err, "Expected exactly 1 embedding", "id", id, "query", query, "got", len(resp.Data))
			allErrors.AddCause(errors.New("Expected exactly 1 embedding"))
			continue
		}

		if len(resp.Data[0].Embedding) != oai.SmallEmbeddingsDims {
			log.Error(err, "Embeddings have wrong dimension", "id", id, "query", query, "got", len(resp.Data[0].Embedding), "want", oai.SmallEmbeddingsDims)
			allErrors.AddCause(errors.New("Embeddings have wrong dimension"))
			continue
		}

		e := v1alpha1.Example{}
		if err := os.WriteFile(expectedFile, encoded, 0777); err != nil {
			log.Error(err, "Failed to serialize doc", "id", b.ID)
			allErrors.AddCause(err)
			continue
		}
	}

	if len(allErrors.Causes) > 0 {
		return allErrors
	}
	return nil
}

// docToExample converts a doc into an example which is a query and an answer
func docToQuery(doc *v1alpha1.Doc) string {
	var sb strings.Builder
	for _, b := range doc.Blocks[:len(doc.Blocks)-1] {
		sb.WriteString(docs.BlockToMarkdown())
	}
	return sb.String()
}
