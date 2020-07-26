package main

import (
	nativeerrors "errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/rpcclient"
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
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
	spawn(func() {
		wg := sync.WaitGroup{}
		for i := uint64(0); numberOfBlocks == 0 || i < numberOfBlocks; i++ {
			foundBlock := make(chan *util.Block)
			mineNextBlock(client, miningAddr, foundBlock, mineWhenNotSynced, templateStopChan, errChan)
			block := <-foundBlock
			templateStopChan <- struct{}{}
			wg.Add(1)
			spawn(func() {
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
	spawn(func() {
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

func mineNextBlock(client *minerClient, miningAddr util.Address, foundBlock chan *util.Block, mineWhenNotSynced bool,
	templateStopChan chan struct{}, errChan chan error) {

	newTemplateChan := make(chan *rpcmodel.GetBlockTemplateResult)
	spawn(func() {
		templatesLoop(client, miningAddr, newTemplateChan, errChan, templateStopChan)
	})
	spawn(func() {
		solveLoop(newTemplateChan, foundBlock, mineWhenNotSynced, errChan)
	})
}

func handleFoundBlock(client *minerClient, block *util.Block) error {
	log.Infof("Found block %s with parents %s. Submitting to %s", block.Hash(), block.MsgBlock().Header.ParentHashes, client.Host())

	err := client.SubmitBlock(block, &rpcmodel.SubmitBlockOptions{})
	if err != nil {
		return errors.Errorf("Error submitting block %s to %s: %s", block.Hash(), client.Host(), err)
	}
	return nil
}

func solveBlock(block *util.Block, stopChan chan struct{}, foundBlock chan *util.Block) {
	msgBlock := block.MsgBlock()
	targetDifficulty := util.CompactToBig(msgBlock.Header.Bits)
	initialNonce := random.Uint64()
	for i := initialNonce; i != initialNonce-1; i++ {
		select {
		case <-stopChan:
			return
		default:
			msgBlock.Header.Nonce = i
			hash := msgBlock.BlockHash()
			atomic.AddUint64(&hashesTried, 1)
			if daghash.HashToBig(hash).Cmp(targetDifficulty) <= 0 {
				foundBlock <- block
				return
			}
		}
	}

}

func templatesLoop(client *minerClient, miningAddr util.Address,
	newTemplateChan chan *rpcmodel.GetBlockTemplateResult, errChan chan error, stopChan chan struct{}) {

	longPollID := ""
	getBlockTemplateLongPoll := func() {
		if longPollID != "" {
			log.Infof("Requesting template with longPollID '%s' from %s", longPollID, client.Host())
		} else {
			log.Infof("Requesting template without longPollID from %s", client.Host())
		}
		template, err := getBlockTemplate(client, miningAddr, longPollID)
		if nativeerrors.Is(err, rpcclient.ErrResponseTimedOut) {
			log.Infof("Got timeout while requesting template '%s' from %s", longPollID, client.Host())
			return
		} else if err != nil {
			errChan <- errors.Errorf("Error getting block template from %s: %s", client.Host(), err)
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
		case <-client.onBlockAdded:
			getBlockTemplateLongPoll()
		case <-time.Tick(500 * time.Millisecond):
			getBlockTemplateLongPoll()
		}
	}
}

func getBlockTemplate(client *minerClient, miningAddr util.Address, longPollID string) (*rpcmodel.GetBlockTemplateResult, error) {
	return client.GetBlockTemplate(miningAddr.String(), longPollID)
}

func solveLoop(newTemplateChan chan *rpcmodel.GetBlockTemplateResult, foundBlock chan *util.Block,
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
		block, err := rpcclient.ConvertGetBlockTemplateResultToBlock(template)
		if err != nil {
			errChan <- errors.Errorf("Error parsing block: %s", err)
			return
		}

		spawn(func() {
			solveBlock(block, stopOldTemplateSolving, foundBlock)
		})
	}
	if stopOldTemplateSolving != nil {
		close(stopOldTemplateSolving)
	}
}
