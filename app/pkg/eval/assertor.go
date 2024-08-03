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
	client v1alpha1connect.AIServiceClient
}

func NewAssertRunner(config config.Config) *AssertRunner {
	return &AssertRunner{config: config, client: client}
}

func newAIServiceClient(baseURL string) v1alpha1connect.AIServiceClient {
	// Create a new client
	client := v1alpha1connect.NewAIServiceClient(
		// N.B. We need to use HTTP2 if we want to support bidirectional streaming
		//http.DefaultClient,
		&http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
					// Use the standard Dial function to create a plain TCP connection
					return net.Dial(network, addr)
				},
			},
		},
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

	agent, err := e.setupAgent(ctx, *experiment.Spec.Agent)
	if err != nil {
		return err
	}

	// List all the files
	files, err := e.listEvalFiles(ctx, experiment.Spec.EvalDir)
	if err != nil {
		return err
	}

	log.Info("Found eval files", "numFiles", len(files))

	// Now iterate over the DB and figure out which files haven't  been loaded into the db.

	unloadedFiles, err := e.findUnloadedFiles(ctx, db, files)
	if err != nil {
		return err
	}
	log.Info("Found unloaded files", "numFiles", len(unloadedFiles))

	// We need to load the evaluation data into the database.
	if err := e.loadMarkdownFiles(ctx, db, unloadedFiles); err != nil {
		return err
	}

	// Now generate predictions for any results that are missing them.
	if err := e.reconcilePredictions(ctx, db, agent); err != nil {
		return err
	}

	tracesDB, err := pebble.Open(e.config.GetTracesDBDir(), &pebble.Options{})
	if err != nil {
		return err
	}
	defer helpers.DeferIgnoreError(tracesDB.Close)

	if err := e.reconcileBestRAGResult(ctx, db, tracesDB); err != nil {
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
