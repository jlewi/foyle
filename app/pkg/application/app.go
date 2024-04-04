package application

import (
	"fmt"
	"io"
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
	Config *config.Config
	Out    io.Writer
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
	log.Info("Shutting down the application")
	return nil
}
