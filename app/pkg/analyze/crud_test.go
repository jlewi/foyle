package analyze

import (
	"context"
	"os"
	"testing"

	"connectrpc.com/connect"
	"github.com/jlewi/foyle/app/pkg/dbutil"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"

	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/pkg/config"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/pkg/errors"
)

func populateDB(db *pebble.DB) error {
	block := &logspb.BlockLog{
		Id:         "test-id",
		GenTraceId: "sometrace",
	}

	return dbutil.SetProto(db, block.GetId(), block)
}

func populateTraceDB(db *pebble.DB) error {
	trace := &logspb.Trace{
		Id: "test-trace",
		Data: &logspb.Trace_Generate{
			Generate: &logspb.GenerateTrace{
				Request: &v1alpha1.GenerateRequest{
					Doc: &v1alpha1.Doc{
						Blocks: []*v1alpha1.Block{
							{
								Id:       "test-block",
								Contents: "echo hello",
							},
						},
					},
				},
			},
		},
	}
	return dbutil.SetProto(db, trace.GetId(), trace)
}

func createCrudHandler() (*CrudHandler, error) {
	logsDir, err := os.MkdirTemp("", "crudTest")
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create temp dir")
	}
	cfg := config.Config{
		Logging: config.Logging{
			LogDir: logsDir,
		},
	}

	db, err := pebble.Open(cfg.GetBlocksDBDir(), &pebble.Options{})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open DB")
	}

	if err := populateDB(db); err != nil {
		return nil, errors.Wrapf(err, "Failed to populate DB")
	}

	tracesDB, err := pebble.Open(cfg.GetTracesDBDir(), &pebble.Options{})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open DB")
	}

	if err := populateTraceDB(tracesDB); err != nil {
		return nil, errors.Wrapf(err, "Failed to populate DB")
	}

	// Create a new CrudHandler with a mock configuration
	return NewCrudHandler(cfg, db, tracesDB, nil)
}

func tearDown(handler *CrudHandler) {
	handler.blocksDB.Close()
	handler.tracesDB.Close()
}

func TestGetTrace(t *testing.T) {
	handler, err := createCrudHandler()
	defer tearDown(handler)

	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	type testCase struct {
		name         string
		id           string
		expectedCode *connect.Code
	}

	notFound := connect.CodeNotFound
	cases := []testCase{
		{
			name: "basic",
			id:   "test-trace",
		},
		{
			name:         "notFound",
			id:           "someunknownid",
			expectedCode: &notFound,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := &connect.Request[logspb.GetTraceRequest]{
				Msg: &logspb.GetTraceRequest{
					Id: c.id,
				},
			}
			resp, err := handler.GetTrace(context.Background(), req)

			if c.expectedCode == nil {
				if resp == nil {
					t.Fatalf("Expected response but got nil")
				}
			} else {
				if connect.CodeOf(err) != *c.expectedCode {
					t.Fatalf("Expected code %v, got %v", *c.expectedCode, connect.CodeOf(err))
				}
			}
		})
	}
}
