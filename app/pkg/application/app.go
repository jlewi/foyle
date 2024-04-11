package application

import (
	"context"
	"fmt"
	"github.com/honeycombio/honeycomb-opentelemetry-go"
	"github.com/honeycombio/otel-config-go/otelconfig"
	"github.com/jlewi/hydros/pkg/files"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/server"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// App is a struct to hold values needed across all commands.
// Intent is to simplify initialization across commands.
type App struct {
	Config         *config.Config
	Out            io.Writer
	otelShutdownFn func()
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
			tracerProvider.Shutdown(context.Background())
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

func (a *App) SetupLogging() error {
	if a.Config == nil {
		return errors.New("Config is nil; call LoadConfig first")
	}
	cfg := a.Config
	// Use a non-json configuration configuration
	c := zap.NewDevelopmentConfig()

	// Use the keys used by cloud logging
	// https://cloud.google.com/logging/docs/structured-logging
	c.EncoderConfig.LevelKey = "severity"
	c.EncoderConfig.TimeKey = "time"
	c.EncoderConfig.MessageKey = "message"

	lvl := cfg.GetLogLevel()
	zapLvl := zap.NewAtomicLevel()

	if err := zapLvl.UnmarshalText([]byte(lvl)); err != nil {
		return errors.Wrapf(err, "Could not convert level %v to ZapLevel", lvl)
	}

	c.Level = zapLvl
	newLogger, err := c.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize zap logger; error %v", err))
	}

	zap.ReplaceGlobals(newLogger)

	return nil
}

// SetupServer sets up the server
func (a *App) SetupServer() (*server.Server, error) {
	if a.Config == nil {
		return nil, errors.New("Config is nil; call LoadConfig first")
	}

	s, err := server.NewServer(*a.Config)

	if err != nil {
		return nil, err
	}
	return s, nil
}

// Shutdown the application.
func (a *App) Shutdown() error {
	// N.B. This is a placeholder for any operations that should be performed when shutting down the app
	// TODO(jeremy): Should we flush the logs? I think that becomes important if end up implementing a logger
	// that sends the logs somewhere explicitly.
	l := zap.L()
	log := zapr.NewLogger(l)

	if a.otelShutdownFn != nil {
		log.Info("Shutting down open telemetry")
		a.otelShutdownFn()
	}

	log.Info("Shutting down the application")
	return nil
}
