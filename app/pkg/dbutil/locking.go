package dbutil

import (
	"github.com/cenkalti/backoff/v4"
	"github.com/cockroachdb/pebble"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

// LockingDB is a wrapper around a Pebble DB that provides row locking functionality.
type LockingDB[T proto.Message] struct {
	db         *pebble.DB
	newProto   func() T
	getVersion func(T) string
	setVersion func(T, string)
	locks      sync.Map
}

// NewLockingDB constructs a new LockingDB with the given Pebble DB and proto message constructor.
// We need to pass in a function to construct a new proto message because with generics it's not possible to create a
// new instance of the type parameter for the generic.
// We also need to pass in a function to get the version from the message.
func NewLockingDB[T proto.Message](db *pebble.DB, newFunc func() T, getVersionFunc func(T) string, setVersionFunc func(T, string)) *LockingDB[T] {
	return &LockingDB[T]{db: db, newProto: newFunc, getVersion: getVersionFunc, setVersion: setVersionFunc}
}

// ReadModifyWrite reads a block from the database, modifies it and writes it back.
// If the block doesn't exist an empty BlockLog will be passed to the function.
//
// msg is used to deserialize the value into.
// T should be a pointer to a proto.Message.
// N.B. we pass in msg to use as a container because in generics it doesn't seem possible to create instances of t
func (d *LockingDB[T]) ReadModifyWrite(key string, modify func(T) error) error {
	// Non-nil error means a non retryable error occurred
	op := func() (bool, error) {
		b, closer, err := d.db.Get([]byte(key))
		if err != nil && !errors.Is(err, pebble.ErrNotFound) {
			return false, errors.Wrapf(err, "Failed to read record with key %s", key)
		}
		// Closer is nil on not found
		if closer != nil {
			defer closer.Close()
		}

		msg := d.newProto()
		if !errors.Is(err, pebble.ErrNotFound) {
			if err := proto.Unmarshal(b, msg); err != nil {
				return false, errors.Wrapf(err, "Failed to unmarshal record with key %s", key)
			}
		}

		version := d.getVersion(msg)
		randomVersion := false
		if version == "" {
			version = uuid.NewString()
			randomVersion = true
		}

		actualVersion, loaded := d.locks.LoadOrStore(key, version)

		if !loaded && actualVersion != version {
			// The version we just read is already outdated so abort the transaction
			return false, nil
		}

		if loaded && randomVersion {
			// We need to ensure we delete the lock because no one else will be able to update it because
			// the version is a random UUID
			defer d.locks.Delete(key)
		}

		if err := modify(msg); err != nil {
			return false, errors.Wrapf(err, "Failed to modify block with key %s", key)
		}

		newVersion := uuid.NewString()
		d.setVersion(msg, newVersion)

		if !d.locks.CompareAndSwap(key, version, newVersion) {
			// Somebody else update the key so abort
			return false, nil
		}

		// If random version is set then we've already scheduled the delete. Otherwise we need to schedule it now
		if !randomVersion {
			defer d.locks.Delete(key)
		}

		if err := SetProto(d.db, key, msg); err != nil {
			return false, errors.Wrapf(err, "Failed to write record with key %s", key)
		}
		return true, nil
	}

	// Retry with backoff
	boff := backoff.NewExponentialBackOff(backoff.WithMaxElapsedTime(3 * time.Minute))

	for {
		success, err := op()
		if success {
			return nil
		}
		if err != nil {
			return err
		}
		next := boff.NextBackOff()
		if next == boff.Stop {
			return errors.New("Timeout trying to update record with key %s")
		}
		time.Sleep(next)
	}
}
