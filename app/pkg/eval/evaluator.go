package eval

import (
	"context"
	"fmt"
	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/agent"
	"github.com/jlewi/foyle/app/pkg/oai"
	"os"
	"path/filepath"

	"github.com/cockroachdb/pebble"
	"github.com/google/uuid"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/executor"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Evaluator struct {
	config config.Config
	parser *executor.BashishParser
}

func NewEvaluator(cfg config.Config) (*Evaluator, error) {
	parser, err := executor.NewBashishParser()

	if err != nil {
		return nil, err
	}

	return &Evaluator{
		config: cfg,
		parser: parser,
	}, nil
}

func (e *Evaluator) Reconcile(ctx context.Context, experiment api.Experiment) error {
	db, err := pebble.Open(experiment.Spec.DBDir, &pebble.Options{})
	if err != nil {
		return err
	}
	defer helpers.DeferIgnoreError(db.Close)

	agent, err := e.setupAgent(ctx, *experiment.Spec.Agent)
	if err != nil {
		return err
	}

	// List all the files
	files, err := e.listEvalFiles(ctx, experiment.Spec.EvalDir)
	if err != nil {
		return err
	}

	log := logs.FromContext(ctx)
	log.Info("Found eval files", "numFiles", len(files))

	// Now iterate over the DB and figure out which files haven't  been loaded into the db.

	unloadedFiles, err := e.findUnloadedFiles(ctx, db, files)
	if err != nil {
		return err
	}
	log.Info("Found unloaded files", "numFiles", len(unloadedFiles))

	// We need to load the evaluation data into the database.
	if err := e.loadFoyleFiles(ctx, db, unloadedFiles); err != nil {
		return err
	}

	// Now generate predictions for any results that are missing them.
	if err := e.reconcilePredictions(ctx, db, agent); err != nil {
		return err
	}

	// Compute the distance
	if err := e.reconcileDistance(ctx, db); err != nil {
		return err
	}

	// Update the Google Sheet
	if err := e.updateGoogleSheet(ctx, experiment, db); err != nil {
		return err
	}
	return nil
}

func (e *Evaluator) setupAgent(ctx context.Context, agentConfig api.AgentConfig) (*agent.Agent, error) {
	cfg := e.config.DeepCopy()

	// Swap out the AgentConfig
	cfg.Agent = &agentConfig

	// Ensure we are in evaluation mode.
	cfg.Agent.EvalMode = true

	client, err := oai.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	agent, err := agent.NewAgent(cfg, client)

	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (e *Evaluator) reconcilePredictions(ctx context.Context, db *pebble.DB, agent *agent.Agent) error {
	olog := logs.FromContext(ctx)
	iter, err := db.NewIterWithContext(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if key == nil {
			break
		}

		log := olog.WithValues("id", string(key))
		value, err := iter.ValueAndErr()
		if err != nil {
			return errors.Wrapf(err, "Failed to read value for key %s", string(key))
		}

		result := &v1alpha1.EvalResult{}
		if err := proto.Unmarshal(value, result); err != nil {
			return errors.Wrapf(err, "Failed to unmarshal value for key %s", string(key))
		}

		if len(result.GetActual()) > 0 {
			log.Info("Skipping; already have answer", "path", result.ExampleFile)
			// We have the answer so we don't need to generate it.
			continue
		}

		if len(result.Actual) == 0 {
			// We need to generate the answer.
			resp, err := agent.Generate(ctx, &v1alpha1.GenerateRequest{
				Doc: result.Example.Query,
			})
			if err != nil {
				result.Error = err.Error()
				result.Status = v1alpha1.EvalResultStatus_ERROR
				continue
			}

			result.Actual = resp.GetBlocks()
			b, err := proto.Marshal(result)
			if err != nil {
				return errors.Wrapf(err, "Failed to marshal result")
			}
			if err := db.Set(key, b, nil); err != nil {
				return errors.Wrapf(err, "Failed to write result to DB")
			}
		}
	}
	return nil
}

func (e *Evaluator) updateResult(ctx context.Context, id string, result *v1alpha1.EvalResult, db *pebble.DB) error {
	b, err := proto.Marshal(result)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal result")
	}
	if err := db.Set([]byte(id), b, nil); err != nil {
		return errors.Wrapf(err, "Failed to write result to DB")
	}
	return nil
}

func (e *Evaluator) reconcileDistance(ctx context.Context, db *pebble.DB) error {
	olog := logs.FromContext(ctx)
	iter, err := db.NewIterWithContext(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if key == nil {
			break
		}

		log := olog.WithValues("id", string(key))
		value, err := iter.ValueAndErr()
		if err != nil {
			return errors.Wrapf(err, "Failed to read value for key %s", string(key))
		}

		result := &v1alpha1.EvalResult{}
		if err := proto.Unmarshal(value, result); err != nil {
			return errors.Wrapf(err, "Failed to unmarshal value for key %s", string(key))
		}

		if result.Distance >= 0 && result.Status != v1alpha1.EvalResultStatus_UNKNOWN_EVAL_RESULT_STATUS {
			log.Info("Skipping; distance already computed")
			continue
		}

		var actualBlock *v1alpha1.Block

		for _, b := range result.Actual {
			if b.Kind == v1alpha1.BlockKind_CODE {
				actualBlock = b
				break
			}
		}
		if actualBlock == nil {
			log.Info("Skipping; no code blocks found in the answer")
			continue
		}

		if len(result.Example.GetAnswer()) > 1 {
			log.Info("Warning; expected answer more than one answer block. Only the first is used")
		}

		expected, err := e.parser.Parse(result.Example.Answer[0].GetContents())
		if err != nil {
			log.Error(err, "Failed to parse expected answer to command")
			result.Error = err.Error()
			result.Status = v1alpha1.EvalResultStatus_ERROR
			if err := e.updateResult(ctx, string(key), result, db); err != nil {
				log.Error(err, "Failed to update result")
			}
			continue
		}

		actual, err := e.parser.Parse(actualBlock.GetContents())
		if err != nil {
			log.Error(err, "Failed to parse actual answer to command")
			result.Error = err.Error()
			result.Status = v1alpha1.EvalResultStatus_ERROR
			if err := e.updateResult(ctx, string(key), result, db); err != nil {
				log.Error(err, "Failed to update result")
			}
			continue
		}

		distance, err := Distance(expected[0], actual[0])

		if err != nil {
			log.Error(err, "Failed to compute distance")
			result.Error = err.Error()
			result.Status = v1alpha1.EvalResultStatus_ERROR
			if err := e.updateResult(ctx, string(key), result, db); err != nil {
				log.Error(err, "Failed to update result")
			}
			continue
		}

		result.Distance = int32(distance)
		result.Status = v1alpha1.EvalResultStatus_DONE
		if err := e.updateResult(ctx, string(key), result, db); err != nil {
			log.Error(err, "Failed to update result")
		}
	}
	return nil
}

func (e *Evaluator) updateGoogleSheet(ctx context.Context, experiment api.Experiment, db *pebble.DB) error {
	log := logs.FromContext(ctx)
	if e.config.Eval == nil || e.config.Eval.GCPServiceAccount == "" {
		return errors.New("GCPServiceAccount is required to update Google Sheet")
	}

	sheetName := experiment.Spec.SheetName
	sheetID := experiment.Spec.GoogleSheetID
	log.WithValues("spreadsheetID", sheetID, "sheetName", sheetName)
	log.Info("Updating Google Sheet")
	credentialsConfig := &impersonate.CredentialsConfig{
		TargetPrincipal: e.config.Eval.GCPServiceAccount,
		Scopes:          []string{"https://www.googleapis.com/auth/spreadsheets", "https://www.googleapis.com/auth/drive"},
	}

	credentials, err := impersonate.CredentialsTokenSource(ctx, *credentialsConfig)
	if err != nil {
		log.Error(err, "Unable to create impersonated credentials")
		return err
	}

	srv, err := sheets.NewService(ctx, option.WithTokenSource(credentials))
	if err != nil {
		log.Error(err, "Unable to retrieve Sheets client")
		return err
	}

	// Create the sheet if it doesn't exist
	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title: experiment.Spec.SheetName,
					},
				},
			},
		},
	}

	_, err = srv.Spreadsheets.BatchUpdate(experiment.Spec.GoogleSheetID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		apiErr, ok := err.(*googleapi.Error)
		if ok {
			if apiErr.Code == 400 {
				log.V(1).Info("Sheet already exists")
			} else {
				log.Error(err, "Unable to create new sheet ")
				return errors.Wrapf(err, "Unable to create new sheet named: %s", sheetName)
			}
		} else {
			return errors.Wrapf(err, "Unable to create new sheet named: %s", sheetName)
		}
	}

	// Prepare the value range to write
	writeRange := fmt.Sprintf("%s", sheetName)
	values := [][]interface{}{{"id", "prompt", "actual", "expected", "distance"}}

	iter, err := db.NewIterWithContext(ctx, nil)
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

		result := &v1alpha1.EvalResult{}
		if err := proto.Unmarshal(value, result); err != nil {
			return errors.Wrapf(err, "Failed to unmarshal value for key %s", string(key))
		}

		prompt := docs.DocToMarkdown(result.Example.Query)
		row := []interface{}{result.Example.Id, prompt, docs.BlocksToMarkdown(result.Actual), docs.BlocksToMarkdown(result.Example.Answer), result.Distance}
		values = append(values, row)
	}
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	// Write the value range to the sheet
	_, err = srv.Spreadsheets.Values.Update(sheetID, writeRange, valueRange).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	if err != nil {
		log.Error(err, "Unable to write data to sheet")
		return errors.Wrapf(err, "Unable to write data to sheet")
	}

	return nil
}

func (e *Evaluator) findUnloadedFiles(ctx context.Context, db *pebble.DB, files []string) ([]string, error) {
	unprocessed := map[string]bool{}

	iter, err := db.NewIterWithContext(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for _, file := range files {
		unprocessed[file] = true
	}

	// Iterate over the files in the DB and remove them from the list of files to load.
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if key == nil {
			break
		}

		value, err := iter.ValueAndErr()
		if err != nil {
			// Should we ignore the error?
			return nil, errors.Wrapf(err, "Failed to read value for key %s", string(key))
		}

		result := &v1alpha1.EvalResult{}
		if err := proto.Unmarshal(value, result); err != nil {
			return nil, errors.Wrapf(err, "Failed to unmarshal value for key %s", string(key))
		}

		delete(unprocessed, result.ExampleFile)

	}

	toProcess := make([]string, 0, len(unprocessed))
	for file := range unprocessed {
		toProcess = append(toProcess, file)
	}

	return toProcess, nil
}

// listEvalFiles returns a list of the all the Foyle files in the eval directory.
func (e *Evaluator) listEvalFiles(ctx context.Context, evalDir string) ([]string, error) {
	examples := make([]string, 0, 100)
	filepath.Walk(evalDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".foyle" {
			return nil
		}

		examples = append(examples, path)
		return nil
	})

	return examples, nil
}

// loadFoyleFiles loads a bunch of Foyle files representing evaluation data and converts them into example
// protos.
func (e *Evaluator) loadFoyleFiles(ctx context.Context, db *pebble.DB, files []string) error {
	oLog := logs.FromContext(ctx)

	allErrors := &helpers.ListOfErrors{}
	for _, path := range files {
		log := oLog.WithValues("path", path)
		log.Info("Processing file")

		contents, err := os.ReadFile(path)
		if err != nil {
			log.Error(err, "Failed to read file")
			allErrors.AddCause(err)
			// Keep going
			continue
		}

		doc := &v1alpha1.Doc{}
		if err := protojson.Unmarshal(contents, doc); err != nil {
			log.Error(err, "Failed to unmarshal example")
			allErrors.AddCause(err)
			// Keep going
			continue
		}

		if len(doc.GetBlocks()) < 2 {
			log.Info("Skipping doc; too few blocks; at least two are required")
			continue
		}

		answer := doc.GetBlocks()[len(doc.GetBlocks())-1]
		doc.Blocks = doc.Blocks[:len(doc.GetBlocks())-1]
		if answer.Kind != v1alpha1.BlockKind_CODE {
			log.Info("Skipping doc; last block must be code")
			continue
		}

		id := uuid.NewString()
		example := &v1alpha1.Example{
			Id:     id,
			Query:  doc,
			Answer: []*v1alpha1.Block{answer},
		}

		result := &v1alpha1.EvalResult{
			Example:     example,
			ExampleFile: path,
			// initialize distance to a negative value so we can tell when it hasn't been computed
			Distance: -1,
		}

		b, err := proto.Marshal(result)
		if err != nil {
			log.Error(err, "Failed to marshal result")
			allErrors.AddCause(err)
			// Keep going
			continue
		}
		if err := db.Set([]byte(id), b, nil); err != nil {
			log.Error(err, "Failed to write result to DB")
			allErrors.AddCause(err)
			// Keep going
			continue
		}
	}

	if len(allErrors.Causes) > 0 {
		return allErrors
	}

	return nil
}
