package main

import (
	"encoding/hex"
	nativeerrors "errors"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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
	spawn("mineLoop-internalLoop", func() {
		wg := sync.WaitGroup{}
		for i := uint64(0); numberOfBlocks == 0 || i < numberOfBlocks; i++ {
			foundBlock := make(chan *util.Block)
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

func mineNextBlock(client *minerClient, miningAddr util.Address, foundBlock chan *util.Block, mineWhenNotSynced bool,
	templateStopChan chan struct{}, errChan chan error) {

	newTemplateChan := make(chan *appmessage.GetBlockTemplateResponseMessage)
	spawn("templatesLoop", func() {
		templatesLoop(client, miningAddr, newTemplateChan, errChan, templateStopChan)
	})
	spawn("solveLoop", func() {
		solveLoop(newTemplateChan, foundBlock, mineWhenNotSynced, errChan)
	})
}

func handleFoundBlock(client *minerClient, block *util.Block) error {
	log.Infof("Found block %s with parents %s. Submitting to %s", block.Hash(), block.MsgBlock().Header.ParentHashes, client.address())

	err := client.submitBlock(block)
	if err != nil {
		return errors.Errorf("Error submitting block %s to %s: %s", block.Hash(), client.address(), err)
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
	newTemplateChan chan *appmessage.GetBlockTemplateResponseMessage, errChan chan error, stopChan chan struct{}) {

	longPollID := ""
	getBlockTemplateLongPoll := func() {
		if longPollID != "" {
			log.Infof("Requesting template with longPollID '%s' from %s", longPollID, client.address())
		} else {
			log.Infof("Requesting template without longPollID from %s", client.address())
		}
		template, err := client.getBlockTemplate(miningAddr.String(), longPollID)
		if nativeerrors.Is(err, router.ErrTimeout) {
			log.Infof("Got timeout while requesting template '%s' from %s", longPollID, client.address())
			return
		} else if err != nil {
			errChan <- errors.Errorf("Error getting block template from %s: %s", client.address(), err)
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

func solveLoop(newTemplateChan chan *appmessage.GetBlockTemplateResponseMessage, foundBlock chan *util.Block,
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
		block, err := convertGetBlockTemplateResultToBlock(template)
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

func convertGetBlockTemplateResultToBlock(template *appmessage.GetBlockTemplateResponseMessage) (*util.Block, error) {
	// parse parent hashes
	parentHashes := make([]*daghash.Hash, len(template.ParentHashes))
	for i, parentHash := range template.ParentHashes {
		hash, err := daghash.NewHashFromStr(parentHash)
		if err != nil {
			return nil, errors.Wrapf(err, "error decoding hash: '%s'", parentHash)
		}
		parentHashes[i] = hash
	}

	// parse Bits
	bitsUint64, err := strconv.ParseUint(template.Bits, 16, 32)
	if err != nil {
		return nil, errors.Wrapf(err, "error decoding bits: '%s'", template.Bits)
	}
	bits := uint32(bitsUint64)

	// parse hashMerkleRoot
	hashMerkleRoot, err := daghash.NewHashFromStr(template.HashMerkleRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing HashMerkleRoot: '%s'", template.HashMerkleRoot)
	}

	// parse AcceptedIDMerkleRoot
	acceptedIDMerkleRoot, err := daghash.NewHashFromStr(template.AcceptedIDMerkleRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing acceptedIDMerkleRoot: '%s'", template.AcceptedIDMerkleRoot)
	}
	utxoCommitment, err := daghash.NewHashFromStr(template.UTXOCommitment)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing utxoCommitment '%s'", template.UTXOCommitment)
	}
	// parse rest of block
	msgBlock := appmessage.NewMsgBlock(
		appmessage.NewBlockHeader(template.Version, parentHashes, hashMerkleRoot,
			acceptedIDMerkleRoot, utxoCommitment, bits, 0))

	for i, txResult := range template.Transactions {
		reader := hex.NewDecoder(strings.NewReader(txResult.Data))
		tx := &appmessage.MsgTx{}
		if err := tx.KaspaDecode(reader, 0); err != nil {
			return nil, errors.Wrapf(err, "error decoding tx #%d", i)
		}
		msgBlock.AddTransaction(tx)
	}

	block := util.NewBlock(msgBlock)
	return block, nil
}
