// Copyright (c) 2015-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blocklogger

import (
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/mstime"
)

type blockType uint8

const (
	blockTypeHeader blockType = iota
	blockTypeBody
)

var statsPerBlockType = map[blockType]*struct {
	receivedLogBlocks int64
	receivedLogTx     int64
	lastBlockLogTime  time.Time
}{
	blockTypeHeader: {
		lastBlockLogTime: time.Now(),
	},
	blockTypeBody: {
		lastBlockLogTime: time.Now(),
	},
}

// LogBlock logs a new block blue score as an information message
// to show progress to the user. In order to prevent spam, it limits logging to
// one message every 10 seconds with duration and totals included.
func LogBlock(block *externalapi.DomainBlock) {
	currentBlockType := blockTypeBody
	if len(block.Transactions) == 0 {
		currentBlockType = blockTypeHeader
	}

	stats := statsPerBlockType[currentBlockType]
	stats.receivedLogBlocks++
	stats.receivedLogTx += int64(len(block.Transactions))

	now := time.Now()
	duration := now.Sub(stats.lastBlockLogTime)
	if duration < time.Second*10 {
		return
	}

	// Truncate the duration to 10s of milliseconds.
	tDuration := duration.Round(10 * time.Millisecond)

	// Log information about new block blue score.
	blockStr := ""
	txStr := ""
	if currentBlockType == blockTypeBody {
		blockStr = "blocks"
		if stats.receivedLogBlocks == 1 {
			blockStr = "block"
		}

		txStr = "transactions"
		if stats.receivedLogTx == 1 {
			txStr = "transaction"
		}
	} else {
		blockStr = "headers"
		if stats.receivedLogBlocks == 1 {
			blockStr = "header"
		}
	}

	if currentBlockType == blockTypeBody {
		log.Infof("Processed %d %s in the last %s (%d %s, %s)",
			stats.receivedLogBlocks, blockStr, tDuration, stats.receivedLogTx,
			txStr, mstime.UnixMilliseconds(block.Header.TimeInMilliseconds()))
	} else {
		log.Infof("Processed %d %s in the last %s (%s)",
			stats.receivedLogBlocks, blockStr, tDuration, mstime.UnixMilliseconds(block.Header.TimeInMilliseconds()))
	}

	stats.receivedLogBlocks = 0
	stats.receivedLogTx = 0
	stats.lastBlockLogTime = now
}
