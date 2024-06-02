package dbutil

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

// SetProto reads a proto message from a Pebble DB.
func SetProto(db *pebble.DB, key string, value proto.Message) error {
	b, err := proto.Marshal(value)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal proto with key %s", key)
	}

	return db.Set([]byte(key), b, pebble.Sync)
}

// ReadModifyWrite reads a block from the database, modifies it and writes it back.
// If the block doesn't exist an empty BlockLog will be passed to the function.
//
// msg is used to deserialize the value into.
// T should be a pointer to a proto.Message.
// N.B. we pass in msg to use as a container because in generics it doesn't seem possible to create instances of t
func ReadModifyWrite[T proto.Message](db *pebble.DB, key string, msg T, modify func(T) error) error {
	b, closer, err := db.Get([]byte(key))
	if err != nil && !errors.Is(err, pebble.ErrNotFound) {
		return errors.Wrapf(err, "Failed to read record with key %s", key)
	}
	// Closer is nil on not found
	if closer != nil {
		defer closer.Close()
	}

	if err != pebble.ErrNotFound {
		if err := proto.Unmarshal(b, msg); err != nil {
			return errors.Wrapf(err, "Failed to unmarshal block with key %s", key)
		}
	}

	if err := modify(msg); err != nil {
		return errors.Wrapf(err, "Failed to modify block with key %s", key)
	}

	return SetProto(db, key, msg)
}
