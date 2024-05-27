package analyze

import (
	"github.com/cockroachdb/pebble"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// GetProto reads a proto message from a Pebble DB.
func GetProto(db *pebble.DB, key string, value proto.Message) error {
	b, closer, err := db.Get([]byte(key))
	if err != nil {
		return errors.Wrapf(err, "Failed to read proto with key %s", key)
	}
	defer closer.Close()

	if err := proto.Unmarshal(b, value); err != nil {
		return errors.Wrapf(err, "Failed to unmarshal proto with key %s", key)
	}

	return nil
}
