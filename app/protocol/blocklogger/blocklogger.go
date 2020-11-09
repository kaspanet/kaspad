// Copyright (c) 2015-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blocklogger

import (
	"sync"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/mstime"
)

var (
	receivedLogBlocks int64
	receivedLogTx     int64
	lastBlockLogTime  = mstime.Now()
	mtx               sync.Mutex
)

// LogBlock logs a new block blue score as an information message
// to show progress to the user. In order to prevent spam, it limits logging to
// one message every 10 seconds with duration and totals included.
func LogBlock(block *externalapi.DomainBlock) error {
	mtx.Lock()
	defer mtx.Unlock()

	receivedLogBlocks++
	receivedLogTx += int64(len(block.Transactions))

	now := mstime.Now()
	duration := now.Sub(lastBlockLogTime)
	if duration < time.Second*10 {
		return nil
	}

	// Truncate the duration to 10s of milliseconds.
	tDuration := duration.Round(10 * time.Millisecond)

	// Log information about new block blue score.
	blockStr := "blocks"
	if receivedLogBlocks == 1 {
		blockStr = "block"
	}
	txStr := "transactions"
	if receivedLogTx == 1 {
		txStr = "transaction"
	}

	log.Infof("Processed %d %s in the last %s (%d %s, %s)",
		receivedLogBlocks, blockStr, tDuration, receivedLogTx,
		txStr, mstime.UnixMilliseconds(block.Header.TimeInMilliseconds))

	receivedLogBlocks = 0
	receivedLogTx = 0
	lastBlockLogTime = now
	return nil
}
