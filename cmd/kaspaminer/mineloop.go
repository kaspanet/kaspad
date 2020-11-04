package main

import (
	nativeerrors "errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"

	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"

	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))
var hashesTried uint64

const logHashRateInterval = 10 * time.Second

func mineLoop(client *minerClient, numberOfBlocks uint64, blockDelay uint64, mineWhenNotSynced bool,
	miningAddr util.Address) error {

	errChan := make(chan error)

	templateStopChan := make(chan struct{})

	doneChan := make(chan struct{})
	spawn("mineLoop-internalLoop", func() {
		wg := sync.WaitGroup{}
		for i := uint64(0); numberOfBlocks == 0 || i < numberOfBlocks; i++ {
			foundBlock := make(chan *externalapi.DomainBlock)
			mineNextBlock(client, miningAddr, foundBlock, mineWhenNotSynced, templateStopChan, errChan)
			block := <-foundBlock
			templateStopChan <- struct{}{}
			wg.Add(1)
			spawn("mineLoop-handleFoundBlock", func() {
				if blockDelay != 0 {
					time.Sleep(time.Duration(blockDelay) * time.Millisecond)
				}
				err := handleFoundBlock(client, block)
				if err != nil {
					errChan <- err
				}
				wg.Done()
			})
		}
		wg.Wait()
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
			currentHashesTried := hashesTried
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
		solveLoop(newTemplateChan, foundBlock, mineWhenNotSynced, errChan)
	})
}

func handleFoundBlock(client *minerClient, block *externalapi.DomainBlock) error {
	blockHash := hashserialization.BlockHash(block)
	log.Infof("Found block %s with parents %s. Submitting to %s", blockHash, block.Header.ParentHashes, client.Address())

	err := client.SubmitBlock(block)
	if err != nil {
		return errors.Errorf("Error submitting block %s to %s: %s", blockHash, client.Address(), err)
	}
	return nil
}

func solveBlock(block *externalapi.DomainBlock, stopChan chan struct{}, foundBlock chan *externalapi.DomainBlock) {
	targetDifficulty := util.CompactToBig(block.Header.Bits)
	initialNonce := random.Uint64()
	for i := initialNonce; i != initialNonce-1; i++ {
		select {
		case <-stopChan:
			return
		default:
			block.Header.Nonce = i
			hash := hashserialization.BlockHash(block)
			atomic.AddUint64(&hashesTried, 1)
			if hashes.ToBig(hash).Cmp(targetDifficulty) <= 0 {
				foundBlock <- block
				return
			}
		}
	}
}

func templatesLoop(client *minerClient, miningAddr util.Address,
	newTemplateChan chan *appmessage.GetBlockTemplateResponseMessage, errChan chan error, stopChan chan struct{}) {

	longPollID := ""
	getBlockTemplateLongPoll := func() {
		if longPollID != "" {
			log.Infof("Requesting template with longPollID '%s' from %s", longPollID, client.Address())
		} else {
			log.Infof("Requesting template without longPollID from %s", client.Address())
		}
		template, err := client.GetBlockTemplate(miningAddr.String(), longPollID)
		if nativeerrors.Is(err, router.ErrTimeout) {
			log.Infof("Got timeout while requesting template '%s' from %s", longPollID, client.Address())
			return
		} else if err != nil {
			errChan <- errors.Errorf("Error getting block template from %s: %s", client.Address(), err)
			return
		}
		if !template.IsConnected {
			errChan <- errors.Errorf("Kaspad is not connected for %s", client.Address())
			return
		}
		if template.LongPollID != longPollID {
			log.Infof("Got new long poll template: %s", template.LongPollID)
			longPollID = template.LongPollID
			newTemplateChan <- template
		}
	}
	getBlockTemplateLongPoll()
	for {
		select {
		case <-stopChan:
			close(newTemplateChan)
			return
		case <-client.blockAddedNotificationChan:
			getBlockTemplateLongPoll()
		case <-time.Tick(500 * time.Millisecond):
			getBlockTemplateLongPoll()
		}
	}
}

func solveLoop(newTemplateChan chan *appmessage.GetBlockTemplateResponseMessage, foundBlock chan *externalapi.DomainBlock,
	mineWhenNotSynced bool, errChan chan error) {

	var stopOldTemplateSolving chan struct{}
	for template := range newTemplateChan {
		if stopOldTemplateSolving != nil {
			close(stopOldTemplateSolving)
		}

		if !template.IsSynced {
			if !mineWhenNotSynced {
				errChan <- errors.Errorf("got template with isSynced=false")
				return
			}
			log.Warnf("Got template with isSynced=false")
		}

		stopOldTemplateSolving = make(chan struct{})
		block, err := mining.ConvertGetBlockTemplateResultToBlock(template)
		if err != nil {
			errChan <- errors.Errorf("Error parsing block: %s", err)
			return
		}

		spawn("solveBlock", func() {
			solveBlock(block, stopOldTemplateSolving, foundBlock)
		})
	}
	if stopOldTemplateSolving != nil {
		close(stopOldTemplateSolving)
	}
}
