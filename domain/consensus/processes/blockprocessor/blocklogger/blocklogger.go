// Copyright (c) 2015-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blocklogger

import (
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/mstime"
)

var (
	receivedLogBlocks       int64
	receivedLogHeaders      int64
	receivedLogTransactions int64
	lastBlockLogTime        time.Time
)

// LogBlock logs a new block blue score as an information message
// to show progress to the user. In order to prevent spam, it limits logging to
// one message every 10 seconds with duration and totals included.
func LogBlock(block *externalapi.DomainBlock) {
	if len(block.Transactions) == 0 {
		receivedLogHeaders++
	} else {
		receivedLogBlocks++
	}

	receivedLogTransactions += int64(len(block.Transactions))

	now := time.Now()
	duration := now.Sub(lastBlockLogTime)
	if duration < time.Second*10 {
		return
	}

	// Truncate the duration to 10s of milliseconds.
	truncatedDuration := duration.Round(10 * time.Millisecond)

	// Log information about new block blue score.
	blockStr := "blocks"
	if receivedLogBlocks == 1 {
		blockStr = "block"
	}

	txStr := "transactions"
	if receivedLogTransactions == 1 {
		txStr = "transaction"
	}

	headerStr := "headers"
	if receivedLogBlocks == 1 {
		headerStr = "header"
	}

	log.Infof("Processed %d %s and %d %s in the last %s (%d %s, %s)",
		receivedLogBlocks, blockStr, receivedLogHeaders, headerStr, truncatedDuration, receivedLogTransactions,
		txStr, mstime.UnixMilliseconds(block.Header.TimeInMilliseconds()))

	receivedLogBlocks = 0
	receivedLogHeaders = 0
	receivedLogTransactions = 0
	lastBlockLogTime = now
}
