package dbutil

import (
	"os"
	"sync"
	"testing"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"

	"github.com/cockroachdb/pebble"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/jlewi/monogo/helpers"
)

func Test_Locking(t *testing.T) {
	oDir, err := os.MkdirTemp("", "lockingTest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	baseDB, err := pebble.Open(oDir, &pebble.Options{})
	if err != nil {
		t.Fatalf("could not open database %s", oDir)
	}
	defer helpers.DeferIgnoreError(baseDB.Close)

	db := NewLockingDB[*logspb.BlockLog](baseDB, func() *logspb.BlockLog {
		return &logspb.BlockLog{}
	}, func(log *logspb.BlockLog) string {
		return log.ResourceVersion
	}, func(log *logspb.BlockLog, version string) {
		log.ResourceVersion = version
	})

	// We want to test concurrent access to the same key
	// We'll start by writing an initial value to the db
	key := "key"
	if err := db.ReadModifyWrite(key, func(log *logspb.BlockLog) error {
		log.Id = "initialid"
		return nil
	}); err != nil {
		t.Fatalf("Error writing block: %v", err)
	}

	// Now to test concurrent access we will start the modify functions in two goroutines.
	// To verify concurrency we need to ensure that two functions that try to modify the same key both get called
	// at the same time. We want to ensure that ReadWriteModify will end up failing and retrying one of them so that
	// they don't wind up with a last win situation. To test this we need to ensure the modify operations are invoked
	// concurrently. Due to retries each modify operation could be called multiple times. So each modify operation
	// is given an "id". It is given an output channel which it uses to signal that it is ready to run. This
	// notifies the main thread that the modify function has been invoked. Each modify function also has a waitgroup
	// which blocks it from completing. The main is responsible for marking the waitgroup as done so that the modify
	// functions complete. The main thread will listen on the channel for the modify functions to signal they are
	// both running. It will then signal the waitgroup so that the modify functions can complete.
	type updateBlockFunc func(*logspb.BlockLog) error
	modifyFunc := func(block *logspb.BlockLog, id string, canRun *sync.WaitGroup, initRun chan<- string, update updateBlockFunc) error {
		// Signal the fact that this modify function has started
		initRun <- id
		// Block until we are ready to run because we want to wait for both goroutines to be ready
		canRun.Wait()

		result := update(block)
		return result
	}

	// This will create a modify function that will block until we are ready to run
	createModifyFunc := func(id string, waitForReady *sync.WaitGroup, initRun chan<- string, update updateBlockFunc) updateBlockFunc {
		newFunc := func(block *logspb.BlockLog) error {
			return modifyFunc(block, id, waitForReady, initRun, update)
		}
		return newFunc
	}

	controlChan := make(chan string, 5)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	setId := "set"
	setGenerate := createModifyFunc(setId, wg, controlChan, func(block *logspb.BlockLog) error {
		block.GenTraceId = "generateID"
		return nil
	})

	execId := "exec"
	setExec := createModifyFunc(execId, wg, controlChan, func(block *logspb.BlockLog) error {
		block.ExecutedBlock = &v1alpha1.Block{
			Id: execId,
		}
		return nil
	})

	waitForUpdates := sync.WaitGroup{}
	waitForUpdates.Add(2)
	// Start two goroutines that will modify the same key
	go func() {
		err := db.ReadModifyWrite(key, setGenerate)
		if err != nil {
			t.Errorf("Error modifying block: %v", err)
		}
		waitForUpdates.Done()
	}()
	go func() {
		err := db.ReadModifyWrite(key, setExec)
		if err != nil {
			t.Errorf("Error modifying block: %v", err)
		}
		waitForUpdates.Done()
	}()

	// Wait for both goroutines to be ready
	waitOn := map[string]bool{setId: true, execId: true}
	for len(waitOn) > 0 {
		id := <-controlChan
		delete(waitOn, id)
	}

	// Now that we know both goroutines are ready we can signal the waitgroup to allow them to run
	wg.Done()

	// Create a go routing to continue to read from the controlChan so we don't end up blocking
	go func() {
		<-controlChan
	}()

	// Now wait for both updates to complete
	waitForUpdates.Wait()
	// Now read the block and check that both changes were applied and didn't clobber each other
	block := &logspb.BlockLog{}
	if err := GetProto(baseDB, key, block); err != nil {
		t.Errorf("Error reading block: %v", err)
	}

	// Make sure both changes were applied
	if block.GenTraceId != "generateID" {
		t.Errorf("Expected GenTraceId to be generateID but got %s", block.GenTraceId)
	}

	if block.ExecutedBlock.Id != execId {
		t.Errorf("Expected ExecutedBlock.Id %s got %s", execId, block.ExecutedBlock.Id)
	}
}
