package eval

import (
	"context"
	"crypto/tls"
	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1/v1alpha1connect"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// AssertRunner runs assertions in batch mode
type AssertRunner struct {
	config config.Config

	assertions []Assertion
}

func NewAssertRunner(config config.Config) (*AssertRunner, error) {
	runner := &AssertRunner{config: config}

	// Load the assertions
	runner.assertions = make([]Assertion, 0, 10)
	runner.assertions = append(runner.assertions, &AssertCodeAfterMarkdown{})
	runner.assertions = append(runner.assertions, &AssertOneCodeCell{})
	runner.assertions = append(runner.assertions, &AssertEndsWithCodeCell{})
	return runner, nil
}

func newHTTPClient() *http.Client {
	// N.B. We need to use HTTP2 if we want to support bidirectional streaming
	// TODO(jeremy): Add the OTEL transport so we report OTEL metrics? See
	// https://github.com/connectrpc/otelconnect-go?tab=readme-ov-file#configuration-for-internal-services
	return &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				// Use the standard Dial function to create a plain TCP connection
				return net.Dial(network, addr)
			},
		},
	}
}

func newGenerateClient(baseURL string) v1alpha1connect.GenerateServiceClient {
	// Create a new client
	client := v1alpha1connect.NewGenerateServiceClient(
		newHTTPClient(),
		baseURL,
	)
	return client
}

func (r *AssertRunner) ReconcileNode(ctx context.Context, node *yaml.RNode) error {
	job := &api.AssertJob{}
	if err := node.YNode().Decode(job); err != nil {
		return errors.Wrapf(err, "Failed to decode AssertJob")
	}

	return r.Reconcile(ctx, *job)
}

func (r *AssertRunner) Reconcile(ctx context.Context, job api.AssertJob) error {
	log := logs.FromContext(ctx).WithValues("job", job.Metadata.Name)
	log.Info("Opening database", "database", job.Spec.DBDir)
	db, err := pebble.Open(job.Spec.DBDir, &pebble.Options{})
	if err != nil {
		return err
	}
	defer helpers.DeferIgnoreError(db.Close)

	if job.Spec.AgentAddress == "" {
		return errors.New("AgentAddress is required")
	}

	if len(job.Spec.Sources) == 0 {
		return errors.New("Sources must be specified")
	}

	// client := newGenerateClient(job.Spec.AgentAddress)

	// Process all the sources
	for _, source := range job.Spec.Sources {
		if source.MarkdownSource == nil {
			return errors.New("Only MarkdownSource is supported")
		}
		files, err := listEvalFiles(ctx, source.MarkdownSource.Path)
		if err != nil {
			return err
		}

		log.Info("Found eval files", "numFiles", len(files))

		// Now iterate over the DB and figure out which files haven't  been loaded into the db.

		unloadedFiles, err := findUnloadedFiles(ctx, db, files)
		if err != nil {
			return err
		}
		log.Info("Found unloaded files", "numFiles", len(unloadedFiles))

		// We need to load the evaluation data into the database.
		if err := loadMarkdownFiles(ctx, db, unloadedFiles); err != nil {
			return err
		}
	}

	// TODO(jeremy): Should we merge this with the evaluator? How should we update this code to work now that we are
	// doing simulations?
	return errors.New("Not implemented; code needs to be updated to work with the new protos and the new DB schema")
	// Now generate predictions for any results that are missing them.
	//if err := reconcilePredictions(ctx, db, client); err != nil {
	//	return err
	//}

	if err := reconcileAssertions(ctx, r.assertions, db); err != nil {
		return err
	}
	return nil
}

// reconcileAssertions reconciles the assertions with the results
func reconcileAssertions(ctx context.Context, assertions []Assertion, db *pebble.DB) error {
	return errors.New("This code needs to be updated to work with the new protos and the new DB schema")
	//olog := logs.FromContext(ctx)
	//iter, err := db.NewIterWithContext(ctx, nil)
	//if err != nil {
	//	return err
	//}
	//defer iter.Close()
	//
	//for iter.First(); iter.Valid(); iter.Next() {
	//	key := iter.Key()
	//	if key == nil {
	//		break
	//	}
	//
	//	log := olog.WithValues("id", string(key))
	//	value, err := iter.ValueAndErr()
	//	if err != nil {
	//		return errors.Wrapf(err, "Failed to read value for key %s", string(key))
	//	}
	//
	//	result := &v1alpha1.EvalResult{}
	//	if err := proto.Unmarshal(value, result); err != nil {
	//		return errors.Wrapf(err, "Failed to unmarshal value for key %s", string(key))
	//	}
	//
	//	actual := make(map[string]bool)
	//	for _, a := range result.GetAssertions() {
	//		actual[a.GetName()] = true
	//	}
	//
	//	if result.Assertions == nil {
	//		result.Assertions = make([]*v1alpha1.Assertion, 0, len(assertions))
	//	}
	//
	//	for _, a := range assertions {
	//		if _, ok := actual[a.Name()]; ok {
	//			continue
	//		}
	//
	//		// Run the assertion
	//		newA, err := a.Assert(ctx, result.Example.Query, nil, result.Actual)
	//
	//		if err != nil {
	//			log.Error(err, "Failed to run assertion", "name", a.Name())
	//		}
	//
	//		result.Assertions = append(result.Assertions, newA)
	//	}
	//
	//	if err := updateResult(ctx, string(key), result, db); err != nil {
	//		return err
	//	}
	//}
	//return nil
}

// loadMarkdownFiles loads a bunch of markdown files into example protos.
// Unlike loadMarkdownAnswerFiles this function doesn't load any answers.
func loadMarkdownFiles(ctx context.Context, db *pebble.DB, files []string) error {
	return errors.New("This function should no longer be needed with the new protos and use of sqllit")
	//oLog := logs.FromContext(ctx)
	//
	//allErrors := &helpers.ListOfErrors{}
	//for _, path := range files {
	//	log := oLog.WithValues("path", path)
	//	log.Info("Processing file")
	//
	//	contents, err := os.ReadFile(path)
	//	if err != nil {
	//		log.Error(err, "Failed to read file")
	//		allErrors.AddCause(err)
	//		// Keep going
	//		continue
	//	}
	//
	//	doc := &v1alpha1.Doc{}
	//
	//	blocks, err := docs.MarkdownToBlocks(string(contents))
	//	if err != nil {
	//		log.Error(err, "Failed to convert markdown to blocks")
	//		allErrors.AddCause(err)
	//		// Keep going
	//		continue
	//	}
	//
	//	doc.Blocks = blocks
	//
	//	if len(doc.GetBlocks()) < 2 {
	//		log.Info("Skipping doc; too few blocks; at least two are required")
	//		continue
	//	}
	//
	//	// We generate a stable ID for the example by hashing the contents of the document.
	//	example := &v1alpha1.Example{
	//		Query: doc,
	//	}
	//	example.Id = HashExample(example)
	//
	//	result := &v1alpha1.EvalResult{
	//		Example:     example,
	//		ExampleFile: path,
	//		// initialize distance to a negative value so we can tell when it hasn't been computed
	//		Distance: uninitializedDistance,
	//	}
	//
	//	if err := dbutil.SetProto(db, example.GetId(), result); err != nil {
	//		log.Error(err, "Failed to write result to DB")
	//		allErrors.AddCause(err)
	//		// Keep going
	//		continue
	//	}
	//}
	//
	//if len(allErrors.Causes) > 0 {
	//	return allErrors
	//}
	//
	//return nil
}
