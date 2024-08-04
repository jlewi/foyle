package eval

import (
	"context"
	"crypto/tls"
	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1/v1alpha1connect"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/proto"
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
	return &AssertRunner{config: config}, nil
}

func newHTTPClient() *http.Client {
	// N.B. We need to use HTTP2 if we want to support bidirectional streaming
	//http.DefaultClient,
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

	client := newGenerateClient(job.Spec.AgentAddress)

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

	// Now generate predictions for any results that are missing them.
	if err := reconcilePredictions(ctx, db, client); err != nil {
		return err
	}

	if err := reconcileAssertions(ctx, r.assertions, db); err != nil {

	}
	// TODO(jeremy): Run the assertions
	return nil
}

// reconcileAssertions reconciles the assertions with the results
func reconcileAssertions(ctx context.Context, assertions []Assertion, db *pebble.DB) error {
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

		if len(result.GetAssertions()) == len(assertions) {
			log.Info("Skipping; already have assertions", "path", result.ExampleFile)
			// We have the answer so we don't need to generate it.
			continue
		}

		actual := make(map[string]bool)
		for _, a := range result.GetAssertions() {
			actual[a.GetName()] = true
		}

		if result.Assertions == nil {
			result.Assertions = make([]*v1alpha1.Assertion, 0, len(assertions))
		}

		for _, a := range assertions {
			if _, ok := actual[a.Name()]; ok {
				continue
			}

			// Run the assertion
			newA, err := a.Assert(ctx, result.Example.Query, nil, result.Actual)

			if err != nil {
				log.Error(err, "Failed to run assertion", "name", a.Name())
			}

			result.Assertions = append(result.Assertions, newA)
		}

		if err := updateResult(ctx, string(key), result, db); err != nil {
			return err
		}
	}
	return nil
}
