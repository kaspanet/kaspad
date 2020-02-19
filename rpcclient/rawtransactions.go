// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcclient

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/kaspanet/kaspad/util/pointers"

	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// FutureGetRawTransactionResult is a future promise to deliver the result of a
// GetRawTransactionAsync RPC invocation (or an applicable error).
type FutureGetRawTransactionResult chan *response

// Receive waits for the response promised by the future and returns a
// transaction given its hash.
func (r FutureGetRawTransactionResult) Receive() (*util.Tx, error) {
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
	return util.NewTx(&msgTx), nil
}

// GetRawTransactionAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See GetRawTransaction for the blocking version and more details.
func (c *Client) GetRawTransactionAsync(txID *daghash.TxID) FutureGetRawTransactionResult {
	id := ""
	if txID != nil {
		id = txID.String()
	}

	cmd := rpcmodel.NewGetRawTransactionCmd(id, pointers.Int(0))
	return c.sendCmd(cmd)
}

// GetRawTransaction returns a transaction given its ID.
//
// See GetRawTransactionVerbose to obtain additional information about the
// transaction.
func (c *Client) GetRawTransaction(txID *daghash.TxID) (*util.Tx, error) {
	return c.GetRawTransactionAsync(txID).Receive()
}

// FutureGetRawTransactionVerboseResult is a future promise to deliver the
// result of a GetRawTransactionVerboseAsync RPC invocation (or an applicable
// error).
type FutureGetRawTransactionVerboseResult chan *response

// Receive waits for the response promised by the future and returns information
// about a transaction given its hash.
func (r FutureGetRawTransactionVerboseResult) Receive() (*rpcmodel.TxRawResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a gettrawtransaction result object.
	var rawTxResult rpcmodel.TxRawResult
	err = json.Unmarshal(res, &rawTxResult)
	if err != nil {
		return nil, err
	}

	return &rawTxResult, nil
}

// GetRawTransactionVerboseAsync returns an instance of a type that can be used
// to get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See GetRawTransactionVerbose for the blocking version and more details.
func (c *Client) GetRawTransactionVerboseAsync(txID *daghash.TxID) FutureGetRawTransactionVerboseResult {
	id := ""
	if txID != nil {
		id = txID.String()
	}

	cmd := rpcmodel.NewGetRawTransactionCmd(id, pointers.Int(1))
	return c.sendCmd(cmd)
}

// GetRawTransactionVerbose returns information about a transaction given
// its hash.
//
// See GetRawTransaction to obtain only the transaction already deserialized.
func (c *Client) GetRawTransactionVerbose(txID *daghash.TxID) (*rpcmodel.TxRawResult, error) {
	return c.GetRawTransactionVerboseAsync(txID).Receive()
}

// FutureDecodeRawTransactionResult is a future promise to deliver the result
// of a DecodeRawTransactionAsync RPC invocation (or an applicable error).
type FutureDecodeRawTransactionResult chan *response

// Receive waits for the response promised by the future and returns information
// about a transaction given its serialized bytes.
func (r FutureDecodeRawTransactionResult) Receive() (*rpcmodel.TxRawResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a decoderawtransaction result object.
	var rawTxResult rpcmodel.TxRawResult
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
	cmd := rpcmodel.NewDecodeRawTransactionCmd(txHex)
	return c.sendCmd(cmd)
}

// DecodeRawTransaction returns information about a transaction given its
// serialized bytes.
func (c *Client) DecodeRawTransaction(serializedTx []byte) (*rpcmodel.TxRawResult, error) {
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
func (c *Client) CreateRawTransactionAsync(inputs []rpcmodel.TransactionInput,
	amounts map[util.Address]util.Amount, lockTime *uint64) FutureCreateRawTransactionResult {

	convertedAmts := make(map[string]float64, len(amounts))
	for addr, amount := range amounts {
		convertedAmts[addr.String()] = amount.ToKAS()
	}
	cmd := rpcmodel.NewCreateRawTransactionCmd(inputs, convertedAmts, lockTime)
	return c.sendCmd(cmd)
}

// CreateRawTransaction returns a new transaction spending the provided inputs
// and sending to the provided addresses.
func (c *Client) CreateRawTransaction(inputs []rpcmodel.TransactionInput,
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

	cmd := rpcmodel.NewSendRawTransactionCmd(txHex, &allowHighFees)
	return c.sendCmd(cmd)
}

// SendRawTransaction submits the encoded transaction to the server which will
// then relay it to the network.
func (c *Client) SendRawTransaction(tx *wire.MsgTx, allowHighFees bool) (*daghash.TxID, error) {
	return c.SendRawTransactionAsync(tx, allowHighFees).Receive()
}

// FutureSearchRawTransactionsResult is a future promise to deliver the result
// of the SearchRawTransactionsAsync RPC invocation (or an applicable error).
type FutureSearchRawTransactionsResult chan *response

// Receive waits for the response promised by the future and returns the
// found raw transactions.
func (r FutureSearchRawTransactionsResult) Receive() ([]*wire.MsgTx, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal as an array of strings.
	var searchRawTxnsResult []string
	err = json.Unmarshal(res, &searchRawTxnsResult)
	if err != nil {
		return nil, err
	}

	// Decode and deserialize each transaction.
	msgTxns := make([]*wire.MsgTx, 0, len(searchRawTxnsResult))
	for _, hexTx := range searchRawTxnsResult {
		// Decode the serialized transaction hex to raw bytes.
		serializedTx, err := hex.DecodeString(hexTx)
		if err != nil {
			return nil, err
		}

		// Deserialize the transaction and add it to the result slice.
		var msgTx wire.MsgTx
		err = msgTx.Deserialize(bytes.NewReader(serializedTx))
		if err != nil {
			return nil, err
		}
		msgTxns = append(msgTxns, &msgTx)
	}

	return msgTxns, nil
}

// SearchRawTransactionsAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See SearchRawTransactions for the blocking version and more details.
func (c *Client) SearchRawTransactionsAsync(address util.Address, skip, count int, reverse bool, filterAddrs []string) FutureSearchRawTransactionsResult {
	addr := address.EncodeAddress()
	verbose := pointers.Bool(false)
	cmd := rpcmodel.NewSearchRawTransactionsCmd(addr, verbose, &skip, &count,
		nil, &reverse, &filterAddrs)
	return c.sendCmd(cmd)
}

// SearchRawTransactions returns transactions that involve the passed address.
//
// NOTE: RPC servers do not typically provide this capability unless it has
// specifically been enabled.
//
// See SearchRawTransactionsVerbose to retrieve a list of data structures with
// information about the transactions instead of the transactions themselves.
func (c *Client) SearchRawTransactions(address util.Address, skip, count int, reverse bool, filterAddrs []string) ([]*wire.MsgTx, error) {
	return c.SearchRawTransactionsAsync(address, skip, count, reverse, filterAddrs).Receive()
}

// FutureSearchRawTransactionsVerboseResult is a future promise to deliver the
// result of the SearchRawTransactionsVerboseAsync RPC invocation (or an
// applicable error).
type FutureSearchRawTransactionsVerboseResult chan *response

// Receive waits for the response promised by the future and returns the
// found raw transactions.
func (r FutureSearchRawTransactionsVerboseResult) Receive() ([]*rpcmodel.SearchRawTransactionsResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal as an array of raw transaction results.
	var result []*rpcmodel.SearchRawTransactionsResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// SearchRawTransactionsVerboseAsync returns an instance of a type that can be
// used to get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See SearchRawTransactionsVerbose for the blocking version and more details.
func (c *Client) SearchRawTransactionsVerboseAsync(address util.Address, skip,
	count int, includePrevOut, reverse bool, filterAddrs *[]string) FutureSearchRawTransactionsVerboseResult {

	addr := address.EncodeAddress()
	verbose := pointers.Bool(true)
	var prevOut *bool
	if includePrevOut {
		prevOut = pointers.Bool(true)
	}
	cmd := rpcmodel.NewSearchRawTransactionsCmd(addr, verbose, &skip, &count,
		prevOut, &reverse, filterAddrs)
	return c.sendCmd(cmd)
}

// SearchRawTransactionsVerbose returns a list of data structures that describe
// transactions which involve the passed address.
//
// NOTE: RPC servers do not typically provide this capability unless it has
// specifically been enabled.
//
// See SearchRawTransactions to retrieve a list of raw transactions instead.
func (c *Client) SearchRawTransactionsVerbose(address util.Address, skip,
	count int, includePrevOut, reverse bool, filterAddrs []string) ([]*rpcmodel.SearchRawTransactionsResult, error) {

	return c.SearchRawTransactionsVerboseAsync(address, skip, count,
		includePrevOut, reverse, &filterAddrs).Receive()
}

// FutureDecodeScriptResult is a future promise to deliver the result
// of a DecodeScriptAsync RPC invocation (or an applicable error).
type FutureDecodeScriptResult chan *response

// Receive waits for the response promised by the future and returns information
// about a script given its serialized bytes.
func (r FutureDecodeScriptResult) Receive() (*rpcmodel.DecodeScriptResult, error) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a decodescript result object.
	var decodeScriptResult rpcmodel.DecodeScriptResult
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
	cmd := rpcmodel.NewDecodeScriptCmd(scriptHex)
	return c.sendCmd(cmd)
}

// DecodeScript returns information about a script given its serialized bytes.
func (c *Client) DecodeScript(serializedScript []byte) (*rpcmodel.DecodeScriptResult, error) {
	return c.DecodeScriptAsync(serializedScript).Receive()
}
