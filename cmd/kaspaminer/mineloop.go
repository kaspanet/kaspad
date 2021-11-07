package main

import (
	nativeerrors "errors"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/cmd/kaspaminer/templatemanager"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/pow"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

var hashesTried uint64

const logHashRateInterval = 10 * time.Second

func mineLoop(client *minerClient, numberOfBlocks uint64, targetBlocksPerSecond float64, mineWhenNotSynced bool,
	miningAddr util.Address) error {
	rand.Seed(time.Now().UnixNano()) // Seed the global concurrent-safe random source.

	errChan := make(chan error)
	doneChan := make(chan struct{})

	// We don't want to send router.DefaultMaxMessages blocks at once because there's
	// a high chance we'll get disconnected from the node, so we make the channel
	// capacity router.DefaultMaxMessages/2 (we give some slack for getBlockTemplate
	// requests)
	foundBlockChan := make(chan *externalapi.DomainBlock, router.DefaultMaxMessages/2)

	spawn("templatesLoop", func() {
		templatesLoop(client, miningAddr, errChan)
	})

	spawn("blocksLoop", func() {
		const windowSize = 10
		hasBlockRateTarget := targetBlocksPerSecond != 0
		var windowTicker, blockTicker *time.Ticker
		// We use tickers to limit the block rate:
		// 1. windowTicker -> makes sure that the last windowSize blocks take at least windowSize*targetBlocksPerSecond.
		// 2. blockTicker -> makes sure that each block takes at least targetBlocksPerSecond/windowSize.
		// that way we both allow for fluctuation in block rate but also make sure they're not too big (by an order of magnitude)
		if hasBlockRateTarget {
			windowRate := time.Duration(float64(time.Second) / (targetBlocksPerSecond / windowSize))
			blockRate := time.Duration(float64(time.Second) / (targetBlocksPerSecond * windowSize))
			log.Infof("Minimum average time per %d blocks: %s, smaller minimum time per block: %s", windowSize, windowRate, blockRate)
			windowTicker = time.NewTicker(windowRate)
			blockTicker = time.NewTicker(blockRate)
			defer windowTicker.Stop()
			defer blockTicker.Stop()
		}
		windowStart := time.Now()
		for blockIndex := 1; ; blockIndex++ {
			foundBlockChan <- mineNextBlock(mineWhenNotSynced)
			if hasBlockRateTarget {
				<-blockTicker.C
				if (blockIndex % windowSize) == 0 {
					tickerStart := time.Now()
					<-windowTicker.C
					log.Infof("Finished mining %d blocks in: %s. slept for: %s", windowSize, time.Since(windowStart), time.Since(tickerStart))
					windowStart = time.Now()
				}
			}
		}
	})

	spawn("handleFoundBlock", func() {
		for i := uint64(0); numberOfBlocks == 0 || i < numberOfBlocks; i++ {
			block := <-foundBlockChan
			err := handleFoundBlock(client, block)
			if err != nil {
				errChan <- err
				return
			}
		}
		doneChan <- struct{}{}
	})

	logHashRate()

	select {
	case err := <-errChan:
		return err
	case <-doneChan:
		return nil
	}
}

func logHashRate() {
	spawn("logHashRate", func() {
		lastCheck := time.Now()
		for range time.Tick(logHashRateInterval) {
			currentHashesTried := atomic.LoadUint64(&hashesTried)
			currentTime := time.Now()
			kiloHashesTried := float64(currentHashesTried) / 1000.0
			hashRate := kiloHashesTried / currentTime.Sub(lastCheck).Seconds()
			log.Infof("Current hash rate is %.2f Khash/s", hashRate)
			lastCheck = currentTime
			// subtract from hashesTried the hashes we already sampled
			atomic.AddUint64(&hashesTried, -currentHashesTried)
		}
	})
}

func handleFoundBlock(client *minerClient, block *externalapi.DomainBlock) error {
	blockHash := consensushashing.BlockHash(block)
	log.Infof("Submitting block %s to %s", blockHash, client.Address())

	rejectReason, err := client.SubmitBlock(block)
	if err != nil {
		if nativeerrors.Is(err, router.ErrTimeout) {
			log.Warnf("Got timeout while submitting block %s to %s: %s", blockHash, client.Address(), err)
			return client.Reconnect()
		}
		if nativeerrors.Is(err, router.ErrRouteClosed) {
			log.Debugf("Got route is closed while requesting block template from %s. "+
				"The client is most likely reconnecting", client.Address())
			return nil
		}
		if rejectReason == appmessage.RejectReasonIsInIBD {
			const waitTime = 1 * time.Second
			log.Warnf("Block %s was rejected because the node is in IBD. Waiting for %s", blockHash, waitTime)
			time.Sleep(waitTime)
			return nil
		}
		return errors.Wrapf(err, "Error submitting block %s to %s", blockHash, client.Address())
	}
	return nil
}

func mineNextBlock(mineWhenNotSynced bool) *externalapi.DomainBlock {
	nonce := rand.Uint64() // Use the global concurrent-safe random source.
	for {
		nonce++
		// For each nonce we try to build a block from the most up to date
		// block template.
		// In the rare case where the nonce space is exhausted for a specific
		// block, it'll keep looping the nonce until a new block template
		// is discovered.
		block, state := getBlockForMining(mineWhenNotSynced)
		state.Nonce = nonce
		atomic.AddUint64(&hashesTried, 1)
		if state.CheckProofOfWork() {
			mutHeader := block.Header.ToMutable()
			mutHeader.SetNonce(nonce)
			block.Header = mutHeader.ToImmutable()
			log.Infof("Found block %s with parents %s", consensushashing.BlockHash(block), block.Header.DirectParents())
			return block
		}
	}
}

func getBlockForMining(mineWhenNotSynced bool) (*externalapi.DomainBlock, *pow.MinerState) {
	tryCount := 0

	const sleepTime = 500 * time.Millisecond
	const sleepTimeWhenNotSynced = 5 * time.Second

	for {
		tryCount++

		shouldLog := (tryCount-1)%10 == 0
		template, state, isSynced := templatemanager.Get()
		if template == nil {
			if shouldLog {
				log.Info("Waiting for the initial template")
			}
			time.Sleep(sleepTime)
			continue
		}
		if !isSynced && !mineWhenNotSynced {
			if shouldLog {
				log.Warnf("Kaspad is not synced. Skipping current block template")
			}
			time.Sleep(sleepTimeWhenNotSynced)
			continue
		}

		return template, state
	}
}

func templatesLoop(client *minerClient, miningAddr util.Address, errChan chan error) {
	getBlockTemplate := func() {
		template, err := client.GetBlockTemplate(miningAddr.String())
		if nativeerrors.Is(err, router.ErrTimeout) {
			log.Warnf("Got timeout while requesting block template from %s: %s", client.Address(), err)
			reconnectErr := client.Reconnect()
			if reconnectErr != nil {
				errChan <- reconnectErr
			}
			return
		}
		if nativeerrors.Is(err, router.ErrRouteClosed) {
			log.Debugf("Got route is closed while requesting block template from %s. "+
				"The client is most likely reconnecting", client.Address())
			return
		}
		if err != nil {
			errChan <- errors.Wrapf(err, "Error getting block template from %s", client.Address())
			return
		}
		err = templatemanager.Set(template)
		if err != nil {
			errChan <- errors.Wrapf(err, "Error setting block template from %s", client.Address())
			return
		}
	}

	getBlockTemplate()
	const tickerTime = 500 * time.Millisecond
	ticker := time.NewTicker(tickerTime)
	for {
		select {
		case <-client.blockAddedNotificationChan:
			getBlockTemplate()
			ticker.Reset(tickerTime)
		case <-ticker.C:
			getBlockTemplate()
		}
	}
}
