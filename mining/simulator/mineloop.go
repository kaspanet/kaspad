package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/rpcclient"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func parseBlock(template *btcjson.GetBlockTemplateResult) (*util.Block, error) {
	// parse parent hashes
	parentHashes := make([]daghash.Hash, len(template.ParentHashes))
	for i, parentHash := range template.ParentHashes {
		hash, err := daghash.NewHashFromStr(parentHash)
		if err != nil {
			return nil, fmt.Errorf("Error decoding hash %s: %s", parentHash, err)
		}
		parentHashes[i] = *hash
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

func solveBlock(msgBlock *wire.MsgBlock) {
	maxNonce := ^uint64(0) // 2^64 - 1
	targetDifficulty := util.CompactToBig(msgBlock.Header.Bits)
	for i := uint64(0); i < maxNonce; i++ {
		msgBlock.Header.Nonce = i
		hash := msgBlock.BlockHash()
		if daghash.HashToBig(&hash).Cmp(targetDifficulty) <= 0 {
			break
		}
	}

}

func mineLoop(clients []*rpcclient.Client) error {
	clientsCount := int64(len(clients))

	for atomic.LoadInt32(&isRunning) == 1 {
		var currentClient *rpcclient.Client
		if clientsCount == 1 {
			currentClient = clients[0]
		} else {
			currentClient = clients[r.Int63n(clientsCount)]
		}
		log.Printf("Next block will be mined by: %s", currentClient.Host())

		template, err := currentClient.GetBlockTemplate([]string{"coinbasetxn"})
		if err != nil {
			return fmt.Errorf("Error getting block template: %s", err)
		}

		block, err := parseBlock(template)
		if err != nil {
			return fmt.Errorf("Error parsing block: %s", err)
		}

		msgBlock := block.MsgBlock()

		msgBlock.Header.HashMerkleRoot = *blockdag.BuildHashMerkleTreeStore(block.Transactions()).Root()
		msgBlock.Header.IDMerkleRoot = *blockdag.BuildIDMerkleTreeStore(block.Transactions()).Root()

		solveBlock(msgBlock)

		log.Printf("Found block %s! Submitting", block.Hash())

		err = currentClient.SubmitBlock(block, &btcjson.SubmitBlockOptions{})
		if err != nil {
			return fmt.Errorf("Error submitting block: %s", err)
		}
	}

	return nil
}
