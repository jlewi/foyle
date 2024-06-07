package analyze

import (
	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/pkg/dbutil"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
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
