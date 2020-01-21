// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcclient

import (
	"encoding/hex"
	"encoding/json"
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util"
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
func (c *Client) SubmitBlockAsync(block *util.Block, options *rpcmodel.SubmitBlockOptions) FutureSubmitBlockResult {
	blockHex := ""
	if block != nil {
		blockBytes, err := block.Bytes()
		if err != nil {
			return newFutureError(err)
		}

		blockHex = hex.EncodeToString(blockBytes)
	}

	cmd := rpcmodel.NewSubmitBlockCmd(blockHex, options)
	return c.sendCmd(cmd)
}

// SubmitBlock attempts to submit a new block into the kaspa network.
func (c *Client) SubmitBlock(block *util.Block, options *rpcmodel.SubmitBlockOptions) error {
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
func (c *Client) GetBlockTemplateAsync(capabilities []string, longPollID string) FutureGetBlockTemplateResult {
	request := &rpcmodel.TemplateRequest{
		Mode:         "template",
		Capabilities: capabilities,
		LongPollID:   longPollID,
	}
	cmd := rpcmodel.NewGetBlockTemplateCmd(request)
	return c.sendCmd(cmd)
}

// Receive waits for the response promised by the future and returns an error if
// any occurred when submitting the block.
func (r FutureGetBlockTemplateResult) Receive() (*rpcmodel.GetBlockTemplateResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	var result rpcmodel.GetBlockTemplateResult
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBlockTemplate request a block template from the server, to mine upon
func (c *Client) GetBlockTemplate(capabilities []string, longPollID string) (*rpcmodel.GetBlockTemplateResult, error) {
	return c.GetBlockTemplateAsync(capabilities, longPollID).Receive()
}
