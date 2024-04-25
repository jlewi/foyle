package learn

import (
	"context"
	"fmt"
	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/analyze"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
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
}

func NewLearner(cfg *config.Config) (*Learner, error) {
	return &Learner{
		Config: cfg,
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
