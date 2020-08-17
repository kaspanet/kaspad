// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package client

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/kaspanet/kaspad/network/appmessage"
	"github.com/kaspanet/kaspad/network/rpc/model"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

// FutureSubmitBlockResult is a future promise to deliver the result of a
// SubmitBlockAsync RPC invocation (or an applicable error).
type FutureSubmitBlockResult chan *response

// Receive waits for the response promised by the future and returns an error if
// any occurred when submitting the block.
func (r FutureSubmitBlockResult) Receive() error {
	res, err := receiveFuture(r)
	if err != nil {
		return err
	}

	if string(res) != "null" {
		var result string
		err = json.Unmarshal(res, &result)
		if err != nil {
			return err
		}

		return errors.New(result)
	}

	return nil
}

// SubmitBlockAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See SubmitBlock for the blocking version and more details.
func (c *Client) SubmitBlockAsync(block *util.Block, options *model.SubmitBlockOptions) FutureSubmitBlockResult {
	blockHex := ""
	if block != nil {
		blockBytes, err := block.Bytes()
		if err != nil {
			return newFutureError(err)
		}

		blockHex = hex.EncodeToString(blockBytes)
	}

	cmd := model.NewSubmitBlockCmd(blockHex, options)
	return c.sendCmd(cmd)
}

// SubmitBlock attempts to submit a new block into the kaspa network.
func (c *Client) SubmitBlock(block *util.Block, options *model.SubmitBlockOptions) error {
	return c.SubmitBlockAsync(block, options).Receive()
}

// FutureGetBlockTemplateResult is a future promise to deliver the result of a
// GetBlockTemplate RPC invocation (or an applicable error).
type FutureGetBlockTemplateResult chan *response

// GetBlockTemplateAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See GetBlockTemplate for the blocking version and more details
func (c *Client) GetBlockTemplateAsync(payAddress string, longPollID string) FutureGetBlockTemplateResult {
	request := &model.TemplateRequest{
		Mode:       "template",
		LongPollID: longPollID,
		PayAddress: payAddress,
	}
	cmd := model.NewGetBlockTemplateCmd(request)
	return c.sendCmd(cmd)
}

// Receive waits for the response promised by the future and returns an error if
// any occurred when submitting the block.
func (r FutureGetBlockTemplateResult) Receive() (*model.GetBlockTemplateResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	var result model.GetBlockTemplateResult
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBlockTemplate request a block template from the server, to mine upon
func (c *Client) GetBlockTemplate(payAddress string, longPollID string) (*model.GetBlockTemplateResult, error) {
	return c.GetBlockTemplateAsync(payAddress, longPollID).Receive()
}

// ConvertGetBlockTemplateResultToBlock Accepts a GetBlockTemplateResult and parses it into a Block
func ConvertGetBlockTemplateResultToBlock(template *model.GetBlockTemplateResult) (*util.Block, error) {
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
