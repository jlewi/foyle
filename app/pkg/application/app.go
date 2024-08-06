package application

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jlewi/foyle/app/pkg/replicate"

	"github.com/jlewi/foyle/app/pkg/llms"

	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/pkg/agent"
	"github.com/jlewi/foyle/app/pkg/anthropic"
	"github.com/jlewi/foyle/app/pkg/dbutil"
	"github.com/jlewi/foyle/app/pkg/learn"
	"github.com/jlewi/foyle/app/pkg/oai"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"

	"github.com/jlewi/foyle/app/pkg/analyze"

	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/eval"
	"github.com/jlewi/hydros/pkg/util"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"go.uber.org/zap/zapcore"

	"github.com/honeycombio/honeycomb-opentelemetry-go"
	"github.com/honeycombio/otel-config-go/otelconfig"
	"github.com/jlewi/hydros/pkg/files"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/server"
	"github.com/jlewi/hydros/pkg/controllers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type logCloser func()

// App is a struct that takes care of wiring together all the different
// components of the application. It is designed in a modular way so different subcommands
// can initialize only the parts they need.
type App struct {
	Config              *config.Config
	Out                 io.Writer
	otelShutdownFn      func()
	logClosers          []logCloser
	Registry            *controllers.Registry
	logEntriesDB        *pebble.DB
	LockingLogEntriesDB *dbutil.LockingDB[*logspb.LogEntries]
	TracesDB            *pebble.DB
	blocksDB            *pebble.DB
	LockingBlocksDB     *dbutil.LockingDB[*logspb.BlockLog]

	analyzer           *analyze.Analyzer
	learner            *learn.Learner
	vectorizer         llms.Vectorizer
	completer          llms.Completer
	inMemoryExamplesDB *learn.InMemoryExampleDB
}

// NewApp creates a new application. You should call one more setup/Load functions to properly set it up.
func NewApp() *App {
	return &App{
		Out: os.Stdout,
	}
}

// LoadConfig loads the config. It takes an optional command. The command allows values to be overwritten from
// the CLI.
func (a *App) LoadConfig(cmd *cobra.Command) error {
	// N.B. at this point we haven't configured any logging so zap just returns the default logger.
	// TODO(jeremy): Should we just initialize the logger without cfg and then reinitialize it after we've read the config?
	if err := config.InitViper(cmd); err != nil {
		return err
	}
	cfg := config.GetConfig()

	if problems := cfg.IsValid(); len(problems) > 0 {
		fmt.Fprintf(os.Stdout, "Invalid configuration; %s\n", strings.Join(problems, "\n"))
		return fmt.Errorf("invalid configuration; fix the problems and then try again")
	}
	a.Config = cfg

	return nil
}

// SetupOTEL sets up OpenTelemetry. Call this function if you want to enable OpenTelemetry.
func (a *App) SetupOTEL() error {
	log := zapr.NewLogger(zap.L())
	if a.Config == nil {
		return errors.New("config shouldn't be nil; did you forget to call LoadConfig?")
	}

	if a.Config.UseHoneycomb() {
		if err := a.useHoneycomb(); err != nil {
			return errors.Wrap(err, "Could not configure Honeycomb")
		}
	} else {
		log.Info("Using default tracer provider")
		// We need to configure a tracer provider so that traces and spans get set even though we aren't actually
		// sending them anywhere.
		tracerProvider := trace.NewTracerProvider()
		otel.SetTracerProvider(tracerProvider)
		a.otelShutdownFn = func() {
			if err := tracerProvider.Shutdown(context.Background()); err != nil {
				log := zapr.NewLogger(zap.L())
				log.Error(err, "Error shutting down tracer provider")
			}
		}
	}

	// Set ottlhttp.DefaultClient to use a transport that will report metrics.
	// For other clients I think we need to use
	// otelhttp.NewTransport and configure the transport.
	otelhttp.DefaultClient = &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	return nil
}

// useHoneycomb configures OTEL to export metrics to Honeycomb
func (a *App) useHoneycomb() error {
	log := zapr.NewLogger(zap.L())
	log.Info("Configuring Honeycomb")

	key, err := files.Read(a.Config.Telemetry.Honeycomb.APIKeyFile)
	if err != nil {
		return errors.Wrapf(err, "Could not read secret: %v", a.Config.Telemetry.Honeycomb.APIKeyFile)
	}

	// Enable multi-span attributes
	bsp := honeycomb.NewBaggageSpanProcessor()

	opts := []otelconfig.Option{
		otelconfig.WithSpanProcessor(bsp),
		honeycomb.WithApiKey(string(key)),
	}

	serviceName := "foyle"
	// The environment variable OTEL_SERVICE_NAME is the default for the honeycomb dataset.
	// https://docs.honeycomb.io/getting-data-in/opentelemetry/go-distro/
	// This will default to unknown. We don't want to use "unknown" as the default value so we override it.
	if os.Getenv("OTEL_SERVICE_NAME") != "" {
		serviceName = os.Getenv("OTEL_SERVICE_NAME")
		log.Info("environment variable OTEL_SERVICE_NAME is set", "service", serviceName)
	}
	log.Info("Setting OTEL_SERVICE_NAME service name", "service", serviceName)

	opts = append(opts, otelconfig.WithServiceName(serviceName))

	// Configure Honeycomb
	otelShutdown, err := otelconfig.ConfigureOpenTelemetry(opts...)
	if err != nil {
		return errors.Wrapf(err, "error setting up open telemetry")
	}
	a.otelShutdownFn = otelShutdown
	return nil
}

func (a *App) SetupLogging(logToFile bool) error {
	if a.Config == nil {
		return errors.New("Config is nil; call LoadConfig first")
	}

	cores := make([]zapcore.Core, 0, 2)
	// `foyle assets download`
	// Configure encoder for JSON format
	if logToFile {
		jsonCore, err := a.createCoreLoggerForFiles()
		if err != nil {
			return errors.Wrap(err, "Could not create core logger for files")
		}
		cores = append(cores, jsonCore)
	}

	consoleCore, err := a.createCoreForConsole()
	if err != nil {
		return errors.Wrap(err, "Could not create core logger for console")
	}

	cores = append(cores, consoleCore)

	// Create a multi-core logger with different encodings
	core := zapcore.NewTee(cores...)

	// Create the logger
	newLogger := zap.New(core)
	// Record the caller of the log message
	newLogger = newLogger.WithOptions(zap.AddCaller())
	zap.ReplaceGlobals(newLogger)

	return nil
}

// SetupRegistry sets up the registry with a list of registered controllers
func (a *App) SetupRegistry() error {
	if a.Config == nil {
		return errors.New("Config is nil; call LoadConfig first")
	}
	a.Registry = &controllers.Registry{}

	evaluator, err := eval.NewEvaluator(*a.Config)
	if err != nil {
		return err
	}
	if err := a.Registry.Register(api.ExperimentGVK, evaluator); err != nil {
		return err
	}

	assertor, err := eval.NewAssertRunner(*a.Config)
	if err != nil {
		return err
	}
	if err := a.Registry.Register(api.AssertJobGVK, assertor); err != nil {
		return err
	}
	return nil
}

func (a *App) createCoreForConsole() (zapcore.Core, error) {
	// Configure encoder for non-JSON format (console-friendly)
	c := zap.NewDevelopmentEncoderConfig()

	// Use the keys used by cloud logging
	// https://cloud.google.com/logging/docs/structured-logging
	c.LevelKey = "severity"
	c.TimeKey = "time"
	c.MessageKey = "message"

	lvl := a.Config.GetLogLevel()
	zapLvl := zap.NewAtomicLevel()

	if err := zapLvl.UnmarshalText([]byte(lvl)); err != nil {
		return nil, errors.Wrapf(err, "Could not convert level %v to ZapLevel", lvl)
	}

	oFile, closer, err := zap.Open("stderr")
	if err != nil {
		return nil, errors.Wrapf(err, "could not create writer for stderr")
	}
	if a.logClosers == nil {
		a.logClosers = []logCloser{}
	}
	a.logClosers = append(a.logClosers, closer)

	encoder := zapcore.NewConsoleEncoder(c)
	core := zapcore.NewCore(encoder, zapcore.AddSync(oFile), zapLvl)
	return core, nil
}

// createCoreLoggerForFiles creates a core logger that writes logs to files. These logs are always written in JSON
// format. Their purpose is to capture AI traces that we use for retraining. Since these are supposed to be machine
// readable they are always written in JSON format.
func (a *App) createCoreLoggerForFiles() (zapcore.Core, error) {
	// Configure encoder for JSON format
	c := zap.NewProductionEncoderConfig()
	// Use the keys used by cloud logging
	// https://cloud.google.com/logging/docs/structured-logging
	c.LevelKey = "severity"
	c.TimeKey = "time"
	c.MessageKey = "message"
	// We attach the function key to the logs because that is useful for identifying the function that generated the log.
	c.FunctionKey = "function"

	jsonEncoder := zapcore.NewJSONEncoder(c)

	logDir := filepath.Join(a.Config.GetLogDir(), "raw")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		// Logger won't be setup yet so we can't use it.
		fmt.Fprintf(os.Stdout, "Creating log directory %s\n", logDir)
		err := os.MkdirAll(logDir, 0755)
		if err != nil {
			return nil, errors.Wrapf(err, "could not create log directory %s", logDir)
		}
	}

	// We need to set a unique file name for the logs as a way of dealing with log rotation.
	name := fmt.Sprintf("foyle.logs.%s.json", time.Now().Format("2006-01-02T15:04:05"))
	logFile := filepath.Join(logDir, name)

	fmt.Fprintf(os.Stdout, "Writing logs to %s\n", logFile)

	oFile, closer, err := zap.Open(logFile)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open log file %s", logFile)
	}
	if a.logClosers == nil {
		a.logClosers = []logCloser{}
	}
	a.logClosers = append(a.logClosers, closer)

	zapLvl := zap.NewAtomicLevel()

	if err := zapLvl.UnmarshalText([]byte(a.Config.GetLogLevel())); err != nil {
		return nil, errors.Wrapf(err, "Could not convert level %v to ZapLevel", a.Config.GetLogLevel())
	}

	// Force log level to be at least info. Because info is the level at which we capture the logs we need for
	// tracing.
	if zapLvl.Level() > zapcore.InfoLevel {
		zapLvl.SetLevel(zapcore.InfoLevel)
	}

	core := zapcore.NewCore(jsonEncoder, zapcore.AddSync(oFile), zapLvl)

	return core, nil
}

// SetupAnalyzer sets up the analyzer
func (a *App) SetupAnalyzer() (*analyze.Analyzer, error) {
	if a.Config == nil {
		return nil, errors.New("Config is nil; call LoadConfig first")
	}

	analyzer, err := analyze.NewAnalyzer(a.Config.GetLogOffsetsFile(), a.LockingLogEntriesDB, a.TracesDB, a.LockingBlocksDB)
	if err != nil {
		return nil, err
	}
	return analyzer, nil
}

// SetupLearner	sets up the learner
func (a *App) SetupLearner() (*learn.Learner, error) {
	if a.LockingBlocksDB == nil {
		return nil, errors.New("LockingBlocksDB is nil; call OpenDBs first")
	}
	client, err := oai.NewClient(*a.Config)
	if err != nil {
		return nil, err
	}
	return learn.NewLearner(*a.Config, client, a.LockingBlocksDB)
}

func (a *App) createComponents() error {
	analyzer, err := a.SetupAnalyzer()
	if err != nil {
		return err
	}
	a.analyzer = analyzer

	learner, err := a.SetupLearner()

	if err != nil {
		return err
	}
	a.learner = learner

	if err := a.setupLLM(); err != nil {
		return err
	}

	var inMemoryExampleDB *learn.InMemoryExampleDB
	if learner != nil {
		inMemoryExampleDB, err = learn.NewInMemoryExampleDB(*a.Config, a.vectorizer)
		if err != nil {
			return err
		}
	}
	a.inMemoryExamplesDB = inMemoryExampleDB
	return nil
}

func (a *App) setupLLM() error {
	client, err := oai.NewClient(*a.Config)
	if err != nil {
		return err
	}
	a.vectorizer = oai.NewVectorizer(client)

	switch a.Config.Agent.ModelProvider {
	case api.ModelProviderAnthropic:
		client, err := anthropic.NewClient(*a.Config)
		if err != nil {
			return err
		}
		completer, err := anthropic.NewCompleter(*a.Config, client)
		if err != nil {
			return err
		}
		a.completer = completer
	case api.ModelProviderReplicate:
		chatClient, err := replicate.NewChatClient(*a.Config)
		if err != nil {
			return err
		}
		completer, err := replicate.NewCompleter(*a.Config, chatClient)
		if err != nil {
			return err
		}

		a.completer = completer
	case api.ModelProviderOpenAI:
		fallthrough
	default:
		completer, err := oai.NewCompleter(*a.Config, client)
		if err != nil {
			return err
		}
		a.completer = completer
	}

	return nil
}

// Serve sets up and runs the server
// This is blocking
func (a *App) Serve() error {
	if a.Config == nil {
		return errors.New("Config is nil; call LoadConfig first")
	}

	// First we create the components.
	if err := a.createComponents(); err != nil {
		return err
	}

	// Start any asynchronous workers in the components
	logDirs := make([]string, 0, 2)
	logDirs = append(logDirs, a.Config.GetRawLogDir())

	if a.Config.Learner != nil {
		logDirs = append(logDirs, a.Config.Learner.LogDirs...)
	}

	if err := a.analyzer.Run(context.Background(), logDirs, a.learner.Enqueue); err != nil {
		return err
	}

	if a.learner != nil {
		if err := a.learner.Start(context.Background(), a.inMemoryExamplesDB.EnqueueExample); err != nil {
			return err
		}
	}

	if err := a.inMemoryExamplesDB.Start(context.Background()); err != nil {
		return err
	}

	agent, err := agent.NewAgent(*a.Config, a.completer, a.inMemoryExamplesDB)

	if err != nil {
		return err
	}

	s, err := server.NewServer(*a.Config, a.blocksDB, agent)

	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	return s.Run()

}

// ApplyPaths applies the resources in the specified paths.
// Paths can be files or directories.
func (a *App) ApplyPaths(ctx context.Context, paths []string) error {
	log := util.LogFromContext(ctx)

	yamlFiles := make([]string, 0, len(paths))
	for _, resourcePath := range paths {
		newPaths, err := util.FindYamlFiles(resourcePath)
		if err != nil {
			log.Error(err, "Failed to find YAML files", "path", resourcePath)
			return err
		}

		yamlFiles = append(yamlFiles, newPaths...)
	}
	sort.Strings(yamlFiles)
	for _, path := range yamlFiles {
		err := a.apply(ctx, path)
		if err != nil {
			log.Error(err, "Apply failed", "path", path)
		}
	}

	return nil
}

func (a *App) apply(ctx context.Context, path string) error {
	if a.Registry == nil {
		return errors.New("Registry is nil; call SetupRegistry first")
	}

	log := zapr.NewLogger(zap.L())
	log.Info("Reading file", "path", path)
	rNodes, err := util.ReadYaml(path)
	if err != nil {
		return err
	}

	allErrors := &util.ListOfErrors{
		Causes: []error{},
	}

	for _, n := range rNodes {
		m, err := n.GetMeta()
		if err != nil {
			log.Error(err, "Failed to get metadata", "n", n)
			continue
		}
		log.Info("Read resource", "meta", m)
		// Going forward we should be using the registry
		gvk := schema.FromAPIVersionAndKind(m.APIVersion, m.Kind)
		controller, err := a.Registry.GetController(gvk)
		if err != nil {
			log.Error(err, "Unsupported kind", "gvk", gvk)
			allErrors.AddCause(err)
			continue
		}

		if err := controller.ReconcileNode(ctx, n); err != nil {
			log.Error(err, "Failed to reconcile resource", "name", m.Name, "namespace", m.Namespace, "gvk", gvk)
			allErrors.AddCause(err)
		}
	}

	if len(allErrors.Causes) == 0 {
		return nil
	}
	allErrors.Final = fmt.Errorf("failed to apply one or more resources")
	return allErrors
}

func (a *App) OpenDBs() error {
	if a.Config == nil {
		return errors.New("Config is nil; call LoadConfig first")
	}

	log := zapr.NewLogger(zap.L())

	log.Info("Opening traces database", "database", a.Config.GetTracesDBDir())
	tracesDB, err := pebble.Open(a.Config.GetTracesDBDir(), &pebble.Options{})
	if err != nil {
		return errors.Wrapf(err, "could not open traces database %s", a.Config.GetTracesDBDir())
	}
	a.TracesDB = tracesDB

	log.Info("Opening blocks database", "database", a.Config.GetBlocksDBDir())
	blocksDB, err := pebble.Open(a.Config.GetBlocksDBDir(), &pebble.Options{})
	if err != nil {
		return errors.Wrapf(err, "could not open blocks database %s", a.Config.GetBlocksDBDir())
	}
	a.blocksDB = blocksDB

	a.LockingBlocksDB = analyze.NewLockingBlocksDB(blocksDB)

	log.Info("Opening loglines database", "database", a.Config.GetBlocksDBDir())
	logEntries, err := pebble.Open(a.Config.GetLogEntriesDBDir(), &pebble.Options{})
	if err != nil {
		return errors.Wrapf(err, "could not open log entries database %s", a.Config.GetLogEntriesDBDir())
	}
	a.logEntriesDB = logEntries

	a.LockingLogEntriesDB = analyze.NewLockingEntriesDB(a.logEntriesDB)

	return nil
}

// Shutdown the application.
func (a *App) Shutdown() error {
	// N.B. This is a placeholder for any operations that should be performed when shutting down the app
	// TODO(jeremy): Should we flush the logs? I think that becomes important if end up implementing a logger
	// that sends the logs somewhere explicitly.
	l := zap.L()
	log := zapr.NewLogger(l)

	if a.analyzer != nil {
		if err := a.analyzer.Shutdown(context.Background()); err != nil {
			log.Error(err, "Error shutting down analyzer")
		}
	}

	// Analyzer should be shutdown before the learner because analyzer tries to enqueue learner items
	if a.learner != nil {
		if err := a.learner.Shutdown(context.Background()); err != nil {
			log.Error(err, "Error shutting down learner")
		}
	}

	if a.inMemoryExamplesDB != nil {
		if err := a.inMemoryExamplesDB.Shutdown(context.Background()); err != nil {
			log.Error(err, "Error shutting down inMemoryExamplesDB")
		}
	}

	if a.otelShutdownFn != nil {
		log.Info("Shutting down open telemetry")
		a.otelShutdownFn()
	}

	if a.TracesDB != nil {
		log.Info("Closing trace database")
		if err := a.TracesDB.Close(); err != nil {
			log.Error(err, "Error closing trace database")
		}
	}

	if a.blocksDB != nil {
		log.Info("Closing blocks database")
		if err := a.blocksDB.Close(); err != nil {
			log.Error(err, "Error closing blocks database")
		}
	}

	log.Info("Shutting down the application")
	// Flush the logs
	for _, closer := range a.logClosers {
		closer()
	}
	return nil
}
