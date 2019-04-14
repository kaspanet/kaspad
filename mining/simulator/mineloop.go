package main

import (
	"encoding/hex"
	"fmt"
	"log"
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

func getBlockTemplate(client *rpcclient.Client, longPollID string) (*btcjson.GetBlockTemplateResult, error) {
	return client.GetBlockTemplate([]string{"coinbasetxn"}, longPollID)
}

func mineLoop(clients []*rpcclient.Client) error {
	clientsCount := int64(len(clients))

	foundBlock := make(chan *util.Block)
	templateChanged := make(chan struct{}, 1)
	errChan := make(chan error, 1)

	var template *btcjson.GetBlockTemplateResult

	setTemplate := func(newTemplate *btcjson.GetBlockTemplateResult) {
		template = newTemplate
		templateChanged <- struct{}{}
	}

	currentClient := clients[0]
	log.Printf("Next block will be mined by: %s", currentClient.Host())

	initTemplate, err := getBlockTemplate(currentClient, "")
	if err != nil {
		return fmt.Errorf("Error getting block template: %s", err)
	}
	setTemplate(initTemplate)

	go func() {
		for block := range foundBlock {
			log.Printf("Found block %s! Submitting to %s", block.Hash(), currentClient.Host())

			err := currentClient.SubmitBlock(block, &btcjson.SubmitBlockOptions{})
			if err != nil {
				errChan <- fmt.Errorf("Error submitting block: %s", err)
				return
			}

			if clientsCount == 1 {
				currentClient = clients[0]
			} else {
				currentClient = clients[random.Int63n(clientsCount)]
			}

			template, err := getBlockTemplate(currentClient, "")
			if err != nil {
				errChan <- fmt.Errorf("Error getting block template: %s", err)
				return
			}
			setTemplate(template)
		}
	}()

	go func() {
		for {
			longPollID := template.LongPollID
			client := currentClient
			longPolledTemplate, err := getBlockTemplate(currentClient, template.LongPollID)
			if err != nil {
				errChan <- fmt.Errorf("Error getting block template: %s", err)
				return
			}
			if longPollID == template.LongPollID && client == currentClient && longPolledTemplate.LongPollID != longPollID {
				log.Printf("Got new long poll template: %s", longPolledTemplate.LongPollID)
				setTemplate(longPolledTemplate)
			}
		}
	}()

	go func() {
		var stopOldTemplateSolving chan struct{}
		for range templateChanged {
			if stopOldTemplateSolving != nil {
				stopOldTemplateSolving <- struct{}{}
				close(stopOldTemplateSolving)
			}
			stopOldTemplateSolving = make(chan struct{}, 1)
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
	}()

	err = <-errChan

	return err
}
