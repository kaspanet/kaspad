package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/rpcclient"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func parseBlock(template *btcjson.GetBlockTemplateResult) (*util.Block, error) {
	// parse parent hashes
	parentHashes := make([]*daghash.Hash, len(template.ParentHashes))
	for i, parentHash := range template.ParentHashes {
		hash, err := daghash.NewHashFromStr(parentHash)
		if err != nil {
			return nil, fmt.Errorf("Error decoding hash %s: %s", parentHash, err)
		}
		parentHashes[i] = hash
	}

	// parse Bits
	bitsInt64, err := strconv.ParseInt(template.Bits, 16, 32)
	if err != nil {
		return nil, fmt.Errorf("Error decoding bits %s: %s", template.Bits, err)
	}
	bits := uint32(bitsInt64)

	// parse rest of block
	msgBlock := wire.NewMsgBlock(wire.NewBlockHeader(template.Version, parentHashes, &daghash.Hash{}, &daghash.Hash{}, uint32(bits), 0))

	for i, txResult := range append([]btcjson.GetBlockTemplateResultTx{*template.CoinbaseTxn}, template.Transactions...) {
		reader := hex.NewDecoder(strings.NewReader(txResult.Data))
		tx := &wire.MsgTx{}
		if err := tx.BtcDecode(reader, 0); err != nil {
			return nil, fmt.Errorf("Error decoding tx #%d: %s", i, err)
		}
		msgBlock.AddTransaction(tx)
	}

	return util.NewBlock(msgBlock), nil
}

func solveBlock(block *util.Block, stopChan chan struct{}, foundBlock chan *util.Block) {
	msgBlock := block.MsgBlock()
	maxNonce := ^uint64(0) // 2^64 - 1
	targetDifficulty := util.CompactToBig(msgBlock.Header.Bits)
	for i := uint64(0); i < maxNonce; i++ {
		select {
		case <-stopChan:
			return
		default:
			msgBlock.Header.Nonce = i
			hash := msgBlock.BlockHash()
			if daghash.HashToBig(hash).Cmp(targetDifficulty) <= 0 {
				foundBlock <- block
				return
			}
		}
	}

}

func getBlockTemplate(client *simulatorClient, longPollID string) (*btcjson.GetBlockTemplateResult, error) {
	return client.GetBlockTemplate([]string{"coinbasetxn"}, longPollID)
}

func templatesLoop(client *simulatorClient, newTemplateChan chan *btcjson.GetBlockTemplateResult, errChan chan error, stopChan chan struct{}) {
	longPollID := ""
	getBlockTemplateLongPoll := func() {
		if longPollID != "" {
			log.Infof("Requesting template with longPollID '%s' from %s", longPollID, client.Host())
		} else {
			log.Infof("Requesting template without longPollID from %s", client.Host())
		}
		template, err := getBlockTemplate(client, longPollID)
		if err == rpcclient.ErrResponseTimedOut {
			log.Infof("Got timeout while requesting template '%s' from %s", longPollID, client.Host())
			return
		} else if err != nil {
			errChan <- fmt.Errorf("Error getting block template: %s", err)
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

func solveLoop(newTemplateChan chan *btcjson.GetBlockTemplateResult, foundBlock chan *util.Block, errChan chan error) {
	var stopOldTemplateSolving chan struct{}
	for template := range newTemplateChan {
		if stopOldTemplateSolving != nil {
			close(stopOldTemplateSolving)
		}
		stopOldTemplateSolving = make(chan struct{})
		block, err := parseBlock(template)
		if err != nil {
			errChan <- fmt.Errorf("Error parsing block: %s", err)
			return
		}

		msgBlock := block.MsgBlock()

		msgBlock.Header.HashMerkleRoot = blockdag.BuildHashMerkleTreeStore(block.Transactions()).Root()
		msgBlock.Header.IDMerkleRoot = blockdag.BuildIDMerkleTreeStore(block.Transactions()).Root()

		go solveBlock(block, stopOldTemplateSolving, foundBlock)
	}
}

func mineNextBlock(client *simulatorClient, foundBlock chan *util.Block, templateStopChan chan struct{}, errChan chan error) {
	newTemplateChan := make(chan *btcjson.GetBlockTemplateResult)
	go templatesLoop(client, newTemplateChan, errChan, templateStopChan)
	go solveLoop(newTemplateChan, foundBlock, errChan)
}

func handleFoundBlock(client *simulatorClient, block *util.Block, templateStopChan chan struct{}) error {
	templateStopChan <- struct{}{}
	log.Infof("Found block %s! Submitting to %s", block.Hash(), client.Host())

	err := client.SubmitBlock(block, &btcjson.SubmitBlockOptions{})
	if err != nil {
		return fmt.Errorf("Error submitting block: %s", err)
	}
	return nil
}

func getRandomClient(clients []*simulatorClient) *simulatorClient {
	clientsCount := int64(len(clients))
	if clientsCount == 1 {
		return clients[0]
	}
	return clients[random.Int63n(clientsCount)]
}

func mineLoop(clients []*simulatorClient) error {
	foundBlock := make(chan *util.Block)
	errChan := make(chan error)

	templateStopChan := make(chan struct{})

	spawn(func() {
		for {
			currentClient := getRandomClient(clients)
			currentClient.notifyForNewBlocks = true
			log.Infof("Next block will be mined by: %s", currentClient.Host())
			mineNextBlock(currentClient, foundBlock, templateStopChan, errChan)
			block, ok := <-foundBlock
			if !ok {
				errChan <- nil
				return
			}
			currentClient.notifyForNewBlocks = false
			err := handleFoundBlock(currentClient, block, templateStopChan)
			if err != nil {
				errChan <- err
				return
			}
		}
	})

	err := <-errChan

	return err
}
