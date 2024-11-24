package analyze

import (
	"context"

	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/pkg/dbutil"
	"github.com/jlewi/foyle/app/pkg/logs"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/pkg/errors"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// NewLockingBlocksDB helper function to create a new LockingDB for BlockLog.
func NewLockingBlocksDB(db *pebble.DB) *dbutil.LockingDB[*logspb.BlockLog] {
	return dbutil.NewLockingDB[*logspb.BlockLog](db, newBlock, getBlockVersion, setBlockVersion)
}

func newBlock() *logspb.BlockLog {
	return &logspb.BlockLog{}
}

func getBlockVersion(block *logspb.BlockLog) string {
	return block.ResourceVersion
}

func setBlockVersion(block *logspb.BlockLog, version string) {
	block.ResourceVersion = version
}

// NewLockingEntriesDB helper function to create a new LockingDB for LogEntries.
func NewLockingEntriesDB(db *pebble.DB) *dbutil.LockingDB[*logspb.LogEntries] {
	return dbutil.NewLockingDB[*logspb.LogEntries](db, newLogEntries, getLogEntriesVersion, setLogEntriesVersion)
}

func newLogEntries() *logspb.LogEntries {
	return &logspb.LogEntries{}
}

func getLogEntriesVersion(m *logspb.LogEntries) string {
	return m.ResourceVersion
}

func setLogEntriesVersion(m *logspb.LogEntries, version string) {
	m.ResourceVersion = version
}

func logDBErrors(ctx context.Context, err error) {
	log := logs.FromContext(ctx)
	var sqlLiteErr *sqlite.Error
	if errors.As(err, &sqlLiteErr) {
		if sqlLiteErr.Code() == sqlite3.SQLITE_BUSY {
			sqlLiteBusyErrs.Inc()
			log.Error(err, "SQLITE_BUSY")
		}
	}
}
