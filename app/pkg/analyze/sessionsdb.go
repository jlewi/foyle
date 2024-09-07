package analyze

import (
	"context"
	"database/sql"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/analyze/fsql"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	_ "embed"
	_ "modernc.org/sqlite"
)

//go:embed fsql/schema.sql
var ddl string

// SessionsDB manages the database containing sessions.
type SessionsDB struct {
	queries *fsql.Queries
}

func NewSessionsDB(cfg config.Config) (*SessionsDB, error) {
	db, err := sql.Open("sqlite", cfg.GetSessionsDB())

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open database: %v", cfg.GetSessionsDB())
	}

	// create tables
	if _, err := db.ExecContext(context.TODO(), ddl); err != nil {
		return nil, err
	}

	// Create the dbtx from the actual database
	queries := fsql.New(db)

	sessions, err := queries.ListSessions(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list sessions")
	}

	log := zapr.NewLogger(zap.L())
	log.Info("Got sessions", "number", len(sessions))
	return &SessionsDB{
		queries: queries,
	}, nil
}
