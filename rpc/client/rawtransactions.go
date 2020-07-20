// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/kaspanet/kaspad/rpc/model"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// FutureDecodeRawTransactionResult is a future promise to deliver the result
// of a DecodeRawTransactionAsync RPC invocation (or an applicable error).
type FutureDecodeRawTransactionResult chan *response

// Receive waits for the response promised by the future and returns information
// about a transaction given its serialized bytes.
func (r FutureDecodeRawTransactionResult) Receive() (*model.TxRawResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a decoderawtransaction result object.
	var rawTxResult model.TxRawResult
	err = json.Unmarshal(res, &rawTxResult)
	if err != nil {
		return nil, err
	}

	return &rawTxResult, nil
}

// DecodeRawTransactionAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See DecodeRawTransaction for the blocking version and more details.
func (c *Client) DecodeRawTransactionAsync(serializedTx []byte) FutureDecodeRawTransactionResult {
	txHex := hex.EncodeToString(serializedTx)
	cmd := model.NewDecodeRawTransactionCmd(txHex)
	return c.sendCmd(cmd)
}

// DecodeRawTransaction returns information about a transaction given its
// serialized bytes.
func (c *Client) DecodeRawTransaction(serializedTx []byte) (*model.TxRawResult, error) {
	return c.DecodeRawTransactionAsync(serializedTx).Receive()
}

// FutureCreateRawTransactionResult is a future promise to deliver the result
// of a CreateRawTransactionAsync RPC invocation (or an applicable error).
type FutureCreateRawTransactionResult chan *response

// Receive waits for the response promised by the future and returns a new
// transaction spending the provided inputs and sending to the provided
// addresses.
func (r FutureCreateRawTransactionResult) Receive() (*wire.MsgTx, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a string.
	var txHex string
	err = json.Unmarshal(res, &txHex)
	if err != nil {
		return nil, err
	}

	// Decode the serialized transaction hex to raw bytes.
	serializedTx, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, err
	}

	// Deserialize the transaction and return it.
	var msgTx wire.MsgTx
	if err := msgTx.Deserialize(bytes.NewReader(serializedTx)); err != nil {
		return nil, err
	}
	return &msgTx, nil
}

// CreateRawTransactionAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See CreateRawTransaction for the blocking version and more details.
func (c *Client) CreateRawTransactionAsync(inputs []model.TransactionInput,
	amounts map[util.Address]util.Amount, lockTime *uint64) FutureCreateRawTransactionResult {

	convertedAmts := make(map[string]float64, len(amounts))
	for addr, amount := range amounts {
		convertedAmts[addr.String()] = amount.ToKAS()
	}
	cmd := model.NewCreateRawTransactionCmd(inputs, convertedAmts, lockTime)
	return c.sendCmd(cmd)
}

// CreateRawTransaction returns a new transaction spending the provided inputs
// and sending to the provided addresses.
func (c *Client) CreateRawTransaction(inputs []model.TransactionInput,
	amounts map[util.Address]util.Amount, lockTime *uint64) (*wire.MsgTx, error) {

	return c.CreateRawTransactionAsync(inputs, amounts, lockTime).Receive()
}

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
func (c *Client) SendRawTransactionAsync(tx *wire.MsgTx, allowHighFees bool) FutureSendRawTransactionResult {
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
func (c *Client) SendRawTransaction(tx *wire.MsgTx, allowHighFees bool) (*daghash.TxID, error) {
	return c.SendRawTransactionAsync(tx, allowHighFees).Receive()
}

// FutureDecodeScriptResult is a future promise to deliver the result
// of a DecodeScriptAsync RPC invocation (or an applicable error).
type FutureDecodeScriptResult chan *response

// Receive waits for the response promised by the future and returns information
// about a script given its serialized bytes.
func (r FutureDecodeScriptResult) Receive() (*model.DecodeScriptResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a decodescript result object.
	var decodeScriptResult model.DecodeScriptResult
	err = json.Unmarshal(res, &decodeScriptResult)
	if err != nil {
		return nil, err
	}

	return &decodeScriptResult, nil
}

// DecodeScriptAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See DecodeScript for the blocking version and more details.
func (c *Client) DecodeScriptAsync(serializedScript []byte) FutureDecodeScriptResult {
	scriptHex := hex.EncodeToString(serializedScript)
	cmd := model.NewDecodeScriptCmd(scriptHex)
	return c.sendCmd(cmd)
}

// DecodeScript returns information about a script given its serialized bytes.
func (c *Client) DecodeScript(serializedScript []byte) (*model.DecodeScriptResult, error) {
	return c.DecodeScriptAsync(serializedScript).Receive()
}
