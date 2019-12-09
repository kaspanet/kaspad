package main

import (
	"encoding/hex"
	nativeerrors "errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/kaspanet/kaspad/rpcclient"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/kaspajson"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func parseBlock(template *kaspajson.GetBlockTemplateResult) (*util.Block, error) {
	// parse parent hashes
	parentHashes := make([]*daghash.Hash, len(template.ParentHashes))
	for i, parentHash := range template.ParentHashes {
		hash, err := daghash.NewHashFromStr(parentHash)
		if err != nil {
			return nil, errors.Errorf("Error decoding hash %s: %s", parentHash, err)
		}
		parentHashes[i] = hash
	}

	// parse Bits
	bitsInt64, err := strconv.ParseInt(template.Bits, 16, 32)
	if err != nil {
		return nil, errors.Errorf("Error decoding bits %s: %s", template.Bits, err)
	}
	bits := uint32(bitsInt64)

	// parseAcceptedIDMerkleRoot
	acceptedIDMerkleRoot, err := daghash.NewHashFromStr(template.AcceptedIDMerkleRoot)
	if err != nil {
		return nil, errors.Errorf("Error parsing acceptedIDMerkleRoot: %s", err)
	}
	utxoCommitment, err := daghash.NewHashFromStr(template.UTXOCommitment)
	if err != nil {
		return nil, errors.Errorf("Error parsing utxoCommitment: %s", err)
	}
	// parse rest of block
	msgBlock := wire.NewMsgBlock(
		wire.NewBlockHeader(template.Version, parentHashes, &daghash.Hash{},
			acceptedIDMerkleRoot, utxoCommitment, bits, 0))

	for i, txResult := range append([]kaspajson.GetBlockTemplateResultTx{*template.CoinbaseTxn}, template.Transactions...) {
		reader := hex.NewDecoder(strings.NewReader(txResult.Data))
		tx := &wire.MsgTx{}
		if err := tx.KaspaDecode(reader, 0); err != nil {
			return nil, errors.Errorf("Error decoding tx #%d: %s", i, err)
		}
		msgBlock.AddTransaction(tx)
	}

	block := util.NewBlock(msgBlock)
	msgBlock.Header.HashMerkleRoot = blockdag.BuildHashMerkleTreeStore(block.Transactions()).Root()
	return block, nil
}

func solveBlock(block *util.Block, stopChan chan struct{}, foundBlock chan *util.Block) {
	msgBlock := block.MsgBlock()
	targetDifficulty := util.CompactToBig(msgBlock.Header.Bits)
	initialNonce := random.Uint64()
	for i := random.Uint64(); i != initialNonce-1; i++ {
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

func getBlockTemplate(client *simulatorClient, longPollID string) (*kaspajson.GetBlockTemplateResult, error) {
	return client.GetBlockTemplate([]string{"coinbasetxn"}, longPollID)
}

func templatesLoop(client *simulatorClient, newTemplateChan chan *kaspajson.GetBlockTemplateResult, errChan chan error, stopChan chan struct{}) {
	longPollID := ""
	getBlockTemplateLongPoll := func() {
		if longPollID != "" {
			log.Infof("Requesting template with longPollID '%s' from %s", longPollID, client.Host())
		} else {
			log.Infof("Requesting template without longPollID from %s", client.Host())
		}
		template, err := getBlockTemplate(client, longPollID)
		if nativeerrors.Is(err, rpcclient.ErrResponseTimedOut) {
			log.Infof("Got timeout while requesting template '%s' from %s", longPollID, client.Host())
			return
		} else if err != nil {
			errChan <- errors.Errorf("Error getting block template: %s", err)
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

func solveLoop(newTemplateChan chan *kaspajson.GetBlockTemplateResult, foundBlock chan *util.Block, errChan chan error) {
	var stopOldTemplateSolving chan struct{}
	for template := range newTemplateChan {
		if stopOldTemplateSolving != nil {
			close(stopOldTemplateSolving)
		}
		stopOldTemplateSolving = make(chan struct{})
		block, err := parseBlock(template)
		if err != nil {
			errChan <- errors.Errorf("Error parsing block: %s", err)
			return
		}

		go solveBlock(block, stopOldTemplateSolving, foundBlock)
	}
	if stopOldTemplateSolving != nil {
		close(stopOldTemplateSolving)
	}
}

func mineNextBlock(client *simulatorClient, foundBlock chan *util.Block, templateStopChan chan struct{}, errChan chan error) {
	newTemplateChan := make(chan *kaspajson.GetBlockTemplateResult)
	go templatesLoop(client, newTemplateChan, errChan, templateStopChan)
	go solveLoop(newTemplateChan, foundBlock, errChan)
}

func handleFoundBlock(client *simulatorClient, block *util.Block) error {
	log.Infof("Found block %s with parents %s! Submitting to %s", block.Hash(), block.MsgBlock().Header.ParentHashes, client.Host())

	err := client.SubmitBlock(block, &kaspajson.SubmitBlockOptions{})
	if err != nil {
		return errors.Errorf("Error submitting block %s to %s: %s", block.Hash(), client.Host(), err)
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

func mineLoop(connManager *connectionManager, blockDelay uint64) error {
	errChan := make(chan error)

	templateStopChan := make(chan struct{})

	spawn(func() {
		for {
			foundBlock := make(chan *util.Block)
			currentClient := getRandomClient(connManager.clients)
			currentClient.notifyForNewBlocks = true
			log.Infof("Next block will be mined by: %s", currentClient.Host())
			mineNextBlock(currentClient, foundBlock, templateStopChan, errChan)
			block, ok := <-foundBlock
			if !ok {
				errChan <- nil
				return
			}
			currentClient.notifyForNewBlocks = false
			templateStopChan <- struct{}{}
			spawn(func() {
				if blockDelay != 0 {
					time.Sleep(time.Duration(blockDelay) * time.Millisecond)
				}
				err := handleFoundBlock(currentClient, block)
				if err != nil {
					errChan <- err
				}
			})
		}
	})

	err := <-errChan

	return err
}
