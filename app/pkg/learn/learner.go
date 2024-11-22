package learn

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/jlewi/foyle/app/pkg/analyze"
	"github.com/jlewi/foyle/app/pkg/runme/converters"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"

	"github.com/jlewi/foyle/app/pkg/docs"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/jlewi/monogo/files"
	"github.com/jlewi/monogo/helpers"

	"k8s.io/client-go/util/workqueue"

	"github.com/jlewi/foyle/app/pkg/config"
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

	sessProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "learner_sessions_processed",
			Help: "Number of sessions processed by the learner",
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
	sessions        *analyze.SessionsManager
	queue           workqueue.DelayingInterface
	postFunc        PostLearnEvent
	eventLoopIsDone sync.WaitGroup
	factory         *files.Factory
	vectorizer      *oai.Vectorizer
}

func NewLearner(cfg config.Config, client *openai.Client, sessions *analyze.SessionsManager) (*Learner, error) {
	if client == nil {
		return nil, errors.New("OpenAI client is required")
	}

	if sessions == nil {
		return nil, errors.New("SessionsManager is required")
	}

	vectorizer := oai.NewVectorizer(client)
	return &Learner{
		Config:     cfg,
		client:     client,
		sessions:   sessions,
		queue:      workqueue.NewDelayingQueue(),
		factory:    &files.Factory{},
		vectorizer: vectorizer,
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
				// N.B. Right now we treat learning errors as permanent and don't retry.
				// The most likely source of retryable errors the vectorizer endpoint should already be handled
				// by using a retryable HTTP client.
				log.Error(err, "Error learning from example", "example", exampleId)
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

// Reconcile learns from the session with the given id
func (l *Learner) Reconcile(ctx context.Context, id string) error {
	log := logs.FromContext(ctx)

	session, err := l.sessions.Get(ctx, id)
	if err != nil {
		sessProcessed.WithLabelValues("nosession").Inc()
		return errors.Wrapf(err, "Unable to learn from session %s; failed to retrieve session", id)
	}

	// Make sure there is a cell execution log event.
	// Right now we rely on the cell execution log event to get the actual cell content.
	// So we can't learn from sessions without cell execution events. TN013 has some ideas for how we could
	// learn in the event of non cell execution events.
	var execEvent *v1alpha1.LogEvent
	for _, event := range session.LogEvents {
		if event.GetType() != v1alpha1.LogEventType_EXECUTE {
			continue
		}

		// We don't want to learn from failed events.
		if event.ExecuteStatus != v1alpha1.LogEvent_SUCCEEDED {
			continue
		}

		// We want to learn from the final successful event.
		execEvent = event
	}

	if execEvent == nil {
		// Since the cell wasn't successfully executed we don't learn from it
		sessProcessed.WithLabelValues("noexec").Inc()
		return nil
	}

	if session.GetFullContext() == nil {
		sessProcessed.WithLabelValues("nocontext").Inc()
		return errors.Errorf("Unable to learn from session %s; session has no context", session.GetContextId())
	}

	if session.GetFullContext().GetNotebook() == nil {
		sessProcessed.WithLabelValues("nonotebook").Inc()
		return errors.Errorf("Unable to learn from session %s; session has no notebook", session.GetContextId())
	}

	if session.GetFullContext().GetSelected() == 0 {
		// If its the first cell we can't learn from it because what would we use as context to predict it?
		sessProcessed.WithLabelValues("firstcell").Inc()
		return nil
	}

	sessProcessed.WithLabelValues("learn").Inc()
	expectedFiles := l.getExampleFiles(session.GetContextId())

	log.Info("Found new training example", "sessionId", session.GetContextId())

	if len(expectedFiles) == 0 {
		sessProcessed.WithLabelValues("noExampleFiles").Inc()
		log.Error(err, "No training files found", "sessionId", session.GetContextId())
		return errors.Wrapf(err, "No training files found for example %s", session.GetContextId())
	}

	var executedCell *parserv1.Cell
	var execID string
	for _, c := range execEvent.Cells {
		execID = converters.GetCellID(c)
		if execID == "" {
			// I don't think this should happen
			sessProcessed.WithLabelValues("cellnoid").Inc()
			continue
		}
		if execID == execEvent.GetSelectedId() {
			executedCell = c
		}
	}

	if executedCell == nil {
		sessProcessed.WithLabelValues("noexeccell").Inc()
		return errors.Errorf("Could not learn from session %s; the executed cell couldn't be found in the session", session.GetContextId())
	}

	executedBlock, err := converters.CellToBlock(executedCell)
	if err != nil {
		log.Error(err, "Failed to convert cell to block", "sessionId", session.GetContextId(), "cellId", execID)
		return errors.Wrapf(err, "Could not learn from session: %s; Could not convert cell to block", session.GetContextId())
	}

	// Make sure the executed block is not the empty string
	executedBlock.Contents = strings.TrimSpace(executedBlock.Contents)

	if executedBlock.Contents == "" {
		sessProcessed.WithLabelValues("emptyblock").Inc()
		return errors.Errorf("Could not learn from session %s; the executed block is empty", session.GetContextId())
	}

	req, err := sessionToQuery(session)
	if err != nil {
		return errors.Wrapf(err, "Could not learn from session %s; Could not convert session to query", session.GetContextId())
	}

	queryBlocks, err := docs.CreateQuery(ctx, req)

	if err != nil {
		log.Error(err, "Failed to create query", "exampleId", session.GetContextId())
		return errors.Wrapf(err, "Failed to create query for example %s", session.GetContextId())
	}

	newDoc := &v1alpha1.Doc{
		Blocks: queryBlocks,
	}

	example := &v1alpha1.Example{
		Id:     session.GetContextId(),
		Query:  newDoc,
		Answer: []*v1alpha1.Block{executedBlock},
	}

	exampleId := session.GetContextId()

	if err := l.computeEmbeddings(ctx, example); err != nil {
		return errors.Wrapf(err, "Failed to compute embeddings for example %s", exampleId)
	}

	encoded, err := proto.Marshal(example)
	if err != nil {
		log.Error(err, "Failed to serialize doc", "id", exampleId)
		return errors.Wrapf(err, "Failed to serialize doc %s", exampleId)
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
				return errors.Wrapf(err, "Failed to create writer for example %s; to file %s", exampleId, expectedFile)
			}
			if closer, ok := w.(io.Closer); ok {
				defer closer.Close()
			}

			if _, err := w.Write(encoded); err != nil {
				return errors.Wrapf(err, "Failed to write example %s; to file %s", exampleId, expectedFile)
			}
			return nil
		}()
		if writeErr != nil {
			// We need to log the individual error here so that its stack trace gets logged
			log.Error(err, "Failed to write example", "id", exampleId, "file", expectedFile)
			writeErrors.AddCause(writeErr)
			continue
		}
		// All post a single file because we don't need to read it multiple times
		if !posted && l.postFunc != nil {
			if err := l.postFunc(expectedFile); err != nil {
				return errors.Wrapf(err, "Failed to post learn event for example %s", exampleId)
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
			log.Error(err, "Unable to get DirHelper", "dir", d)
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

	qVec, err := l.vectorizer.Embed(ctx, example.Query.GetBlocks())

	if err != nil {
		return err
	}

	if len(qVec) != oai.SmallEmbeddingsDims {
		log.Error(err, "Embeddings have wrong dimension", "id", example.Id, "query", example.Query, "got", len(qVec), "want", oai.SmallEmbeddingsDims)
		return errors.Wrapf(err, "Embeddings have wrong dimension; got %v, want %v", len(qVec), oai.SmallEmbeddingsDims)
	}

	example.Embedding = qVec
	return nil
}

func sessionToQuery(session *logspb.Session) (*v1alpha1.GenerateRequest, error) {
	if session.GetFullContext() == nil {
		return nil, errors.Errorf("Unable to learn from session %s; session has no context", session.GetContextId())
	}
	if session.GetFullContext().GetNotebook() == nil {
		return nil, errors.Errorf("Unable to learn from session %s; session has no notebook", session.GetContextId())
	}

	doc, err := converters.NotebookToDoc(session.GetFullContext().GetNotebook())
	if err != nil {
		return nil, errors.Wrapf(err, "Could not learn from session %s; Could not convert notebook to doc", session.GetContextId())
	}

	// We need to truncate the doc to only include the blocks up to the selected index
	if session.GetFullContext().GetSelected() == 0 {
		return nil, errors.Errorf("Unable to learn from session %s; because the selected cell is the first in the doc", session.GetContextId())
	}

	// We need to remove the last block from the doc. Because the last block in the doc i.e. the selected block
	// is actually what we want to predict
	doc.Blocks = doc.Blocks[:session.GetFullContext().GetSelected()]
	selectedIndex := len(doc.Blocks) - 1
	req := &v1alpha1.GenerateRequest{
		Doc:           doc,
		SelectedIndex: int32(selectedIndex),
	}

	return req, nil
}
