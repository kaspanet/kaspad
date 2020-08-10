// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/kaspanet/kaspad/rpc/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

// FutureSendRawTransactionResult is a future promise to deliver the result
// of a SendRawTransactionAsync RPC invocation (or an applicable error).
type FutureSendRawTransactionResult chan *response

// Receive waits for the response promised by the future and returns the result
// of submitting the encoded transaction to the server which then relays it to
// the network.
func (r FutureSendRawTransactionResult) Receive() (*daghash.TxID, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a string.
	var txIDStr string
	err = json.Unmarshal(res, &txIDStr)
	if err != nil {
		return nil, err
	}

	return daghash.NewTxIDFromStr(txIDStr)
}

// SendRawTransactionAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See SendRawTransaction for the blocking version and more details.
func (c *Client) SendRawTransactionAsync(tx *domainmessage.MsgTx, allowHighFees bool) FutureSendRawTransactionResult {
	txHex := ""
	if tx != nil {
		// Serialize the transaction and convert to hex string.
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(buf); err != nil {
			return newFutureError(err)
		}
		txHex = hex.EncodeToString(buf.Bytes())
	}

	cmd := model.NewSendRawTransactionCmd(txHex, &allowHighFees)
	return c.sendCmd(cmd)
}

// SendRawTransaction submits the encoded transaction to the server which will
// then relay it to the network.
func (c *Client) SendRawTransaction(tx *domainmessage.MsgTx, allowHighFees bool) (*daghash.TxID, error) {
	return c.SendRawTransactionAsync(tx, allowHighFees).Receive()
}
