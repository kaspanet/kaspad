// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"testing"

	"github.com/daglabs/btcd/dagconfig"
)

// TestNotifications ensures that notification callbacks are fired on events.
func TestNotifications(t *testing.T) {
	blocks, err := loadBlocks("blk_0_to_4.dat")
	if err != nil {
		t.Fatalf("Error loading file: %v\n", err)
	}

	// Create a new database and dag instance to run tests against.
	dag, teardownFunc, err := DAGSetup("notifications", Config{
		DAGParams: &dagconfig.SimNetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup dag instance: %v", err)
	}
	defer teardownFunc()

	notificationCount := 0
	callback := func(notification *Notification) {
		if notification.Type == NTBlockAdded {
			notificationCount++
		}
	}

	// Register callback multiple times then assert it is called that many
	// times.
	const numSubscribers = 3
	for i := 0; i < numSubscribers; i++ {
		dag.Subscribe(callback)
	}

	isOrphan, err := dag.ProcessBlock(blocks[1], BFNone)
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned block " +
			"is an orphan\n")
	}
	if err != nil {
		t.Fatalf("ProcessBlock fail on block 1: %v\n", err)
	}

	if notificationCount != numSubscribers {
		t.Fatalf("Expected notification callback to be executed %d "+
			"times, found %d", numSubscribers, notificationCount)
	}
}
