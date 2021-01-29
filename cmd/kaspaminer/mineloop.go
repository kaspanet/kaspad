package main

import (
	nativeerrors "errors"
	"github.com/kaspanet/kaspad/util/difficulty"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model/pow"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/app/appmessage"
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

	templateStopChan := make(chan struct{})

	doneChan := make(chan struct{})
	spawn("mineLoop-internalLoop", func() {
		const windowSize = 10
		var expectedDurationForWindow time.Duration
		var windowExpectedEndTime time.Time
		hasBlockRateTarget := targetBlocksPerSecond != 0
		if hasBlockRateTarget {
			expectedDurationForWindow = time.Duration(float64(windowSize)/targetBlocksPerSecond) * time.Second
			windowExpectedEndTime = time.Now().Add(expectedDurationForWindow)
		}
		blockInWindowIndex := 0

		for i := uint64(0); numberOfBlocks == 0 || i < numberOfBlocks; i++ {

			foundBlock := make(chan *externalapi.DomainBlock)
			mineNextBlock(client, miningAddr, foundBlock, mineWhenNotSynced, templateStopChan, errChan)
			block := <-foundBlock
			templateStopChan <- struct{}{}
			err := handleFoundBlock(client, block)
			if err != nil {
				errChan <- err
			}

			if hasBlockRateTarget {
				blockInWindowIndex++
				if blockInWindowIndex == windowSize-1 {
					deviation := windowExpectedEndTime.Sub(time.Now())
					if deviation > 0 {
						log.Infof("Finished to mine %d blocks %s earlier than expected. Sleeping %s to compensate",
							windowSize, deviation, deviation)
						time.Sleep(deviation)
					}
					blockInWindowIndex = 0
					windowExpectedEndTime = time.Now().Add(expectedDurationForWindow)
				}
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

func mineNextBlock(client *minerClient, miningAddr util.Address, foundBlock chan *externalapi.DomainBlock, mineWhenNotSynced bool,
	templateStopChan chan struct{}, errChan chan error) {

	newTemplateChan := make(chan *appmessage.GetBlockTemplateResponseMessage)
	spawn("templatesLoop", func() {
		templatesLoop(client, miningAddr, newTemplateChan, errChan, templateStopChan)
	})
	spawn("solveLoop", func() {
		solveLoop(newTemplateChan, foundBlock, mineWhenNotSynced)
	})
}

func handleFoundBlock(client *minerClient, block *externalapi.DomainBlock) error {
	blockHash := consensushashing.BlockHash(block)
	log.Infof("Found block %s with parents %s. Submitting to %s", blockHash, block.Header.ParentHashes(), client.Address())

	rejectReason, err := client.SubmitBlock(block)
	if err != nil {
		if nativeerrors.Is(err, router.ErrTimeout) {
			log.Warnf("Got timeout while submitting block %s to %s: %s", blockHash, client.Address(), err)
			return nil
		}
		if rejectReason == appmessage.RejectReasonIsInIBD {
			const waitTime = 1 * time.Second
			log.Warnf("Block %s was rejected because the node is in IBD. Waiting for %s", blockHash, waitTime)
			time.Sleep(waitTime)
			return nil
		}
		return errors.Errorf("Error submitting block %s to %s: %s", blockHash, client.Address(), err)
	}
	return nil
}

func solveBlock(block *externalapi.DomainBlock, stopChan chan struct{}, foundBlock chan *externalapi.DomainBlock) {
	targetDifficulty := difficulty.CompactToBig(block.Header.Bits())
	headerForMining := block.Header.ToMutable()
	initialNonce := rand.Uint64() // Use the global concurrent-safe random source.
	for i := initialNonce; i != initialNonce-1; i++ {
		select {
		case <-stopChan:
			return
		default:
			headerForMining.SetNonce(i)
			atomic.AddUint64(&hashesTried, 1)
			if pow.CheckProofOfWorkWithTarget(headerForMining, targetDifficulty) {
				block.Header = headerForMining.ToImmutable()
				foundBlock <- block
				return
			}
		}
	}
}

func templatesLoop(client *minerClient, miningAddr util.Address,
	newTemplateChan chan *appmessage.GetBlockTemplateResponseMessage, errChan chan error, stopChan chan struct{}) {

	getBlockTemplate := func() {
		template, err := client.GetBlockTemplate(miningAddr.String())
		if nativeerrors.Is(err, router.ErrTimeout) {
			log.Warnf("Got timeout while requesting block template from %s: %s", client.Address(), err)
			return
		} else if err != nil {
			errChan <- errors.Errorf("Error getting block template from %s: %s", client.Address(), err)
			return
		}
		newTemplateChan <- template
	}
	getBlockTemplate()
	for {
		select {
		case <-stopChan:
			close(newTemplateChan)
			return
		case <-client.blockAddedNotificationChan:
			getBlockTemplate()
		case <-time.Tick(500 * time.Millisecond):
			getBlockTemplate()
		}
	}
}

func solveLoop(newTemplateChan chan *appmessage.GetBlockTemplateResponseMessage, foundBlock chan *externalapi.DomainBlock,
	mineWhenNotSynced bool) {

	var stopOldTemplateSolving chan struct{}
	for template := range newTemplateChan {
		if !template.IsSynced && !mineWhenNotSynced {
			log.Warnf("Kaspad is not synced. Skipping current block template")
			continue
		}

		if stopOldTemplateSolving != nil {
			close(stopOldTemplateSolving)
		}

		stopOldTemplateSolving = make(chan struct{})
		block := appmessage.MsgBlockToDomainBlock(template.MsgBlock)

		stopOldTemplateSolvingCopy := stopOldTemplateSolving
		spawn("solveBlock", func() {
			solveBlock(block, stopOldTemplateSolvingCopy, foundBlock)
		})
	}
	if stopOldTemplateSolving != nil {
		close(stopOldTemplateSolving)
	}
}
