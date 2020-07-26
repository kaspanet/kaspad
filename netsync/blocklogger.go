// Copyright (c) 2015-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package netsync

import (
	"github.com/kaspanet/kaspad/util/mstime"
	"sync"
	"time"

	"github.com/kaspanet/kaspad/logs"
	"github.com/kaspanet/kaspad/util"
)

// blockProgressLogger provides periodic logging for other services in order
// to show users progress of certain "actions" involving some or all current
// blocks. Ex: syncing, indexing all blocks, etc.
type blockProgressLogger struct {
	receivedLogBlocks int64
	receivedLogTx     int64
	lastBlockLogTime  mstime.Time

	subsystemLogger *logs.Logger
	progressAction  string
	sync.Mutex
}

// newBlockProgressLogger returns a new block progress logger.
// The progress message is templated as follows:
//  {progressAction} {numProcessed} {blocks|block} in the last {timePeriod}
//  ({numTxs}, height {lastBlockHeight}, {lastBlockTimeStamp})
func newBlockProgressLogger(progressMessage string, logger *logs.Logger) *blockProgressLogger {
	return &blockProgressLogger{
		lastBlockLogTime: mstime.Now(),
		progressAction:   progressMessage,
		subsystemLogger:  logger,
	}
}

// LogBlockBlueScore logs a new block blue score as an information message
// to show progress to the user. In order to prevent spam, it limits logging to
// one message every 10 seconds with duration and totals included.
func (b *blockProgressLogger) LogBlockBlueScore(block *util.Block, blueScore uint64) {
	b.Lock()
	defer b.Unlock()

	b.receivedLogBlocks++
	b.receivedLogTx += int64(len(block.MsgBlock().Transactions))

	now := mstime.Now()
	duration := now.Sub(b.lastBlockLogTime)
	if duration < time.Second*10 {
		return
	}

	// Truncate the duration to 10s of milliseconds.
	durationMillis := int64(duration / time.Millisecond)
	tDuration := 10 * time.Millisecond * time.Duration(durationMillis/10)

	// Log information about new block height.
	blockStr := "blocks"
	if b.receivedLogBlocks == 1 {
		blockStr = "block"
	}
	txStr := "transactions"
	if b.receivedLogTx == 1 {
		txStr = "transaction"
	}
	b.subsystemLogger.Infof("%s %d %s in the last %s (%d %s, blue score %d, %s)",
		b.progressAction, b.receivedLogBlocks, blockStr, tDuration, b.receivedLogTx,
		txStr, blueScore, block.MsgBlock().Header.Timestamp)

	b.receivedLogBlocks = 0
	b.receivedLogTx = 0
	b.lastBlockLogTime = now
}
