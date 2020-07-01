// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcclient

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

var (
	// ErrWebsocketsRequired is an error to describe the condition where the
	// caller is trying to use a websocket-only feature, such as requesting
	// notifications or other websocket requests when the client is
	// configured to run in HTTP POST mode.
	ErrWebsocketsRequired = errors.New("a websocket connection is required " +
		"to use this feature")
)

// notificationState is used to track the current state of successfully
// registered notification so the state can be automatically re-established on
// reconnect.
type notificationState struct {
	notifyBlocks            bool
	notifyChainChanges      bool
	notifyNewTx             bool
	notifyNewTxVerbose      bool
	notifyNewTxSubnetworkID *string
}

// Copy returns a deep copy of the receiver.
func (s *notificationState) Copy() *notificationState {
	var stateCopy notificationState
	stateCopy.notifyBlocks = s.notifyBlocks
	stateCopy.notifyChainChanges = s.notifyChainChanges
	stateCopy.notifyNewTx = s.notifyNewTx
	stateCopy.notifyNewTxVerbose = s.notifyNewTxVerbose
	stateCopy.notifyNewTxSubnetworkID = s.notifyNewTxSubnetworkID

	return &stateCopy
}

// newNotificationState returns a new notification state ready to be populated.
func newNotificationState() *notificationState {
	return &notificationState{}
}

// newNilFutureResult returns a new future result channel that already has the
// result waiting on the channel with the reply set to nil. This is useful
// to ignore things such as notifications when the caller didn't specify any
// notification handlers.
func newNilFutureResult() chan *response {
	responseChan := make(chan *response, 1)
	responseChan <- &response{result: nil, err: nil}
	return responseChan
}

// NotificationHandlers defines callback function pointers to invoke with
// notifications. Since all of the functions are nil by default, all
// notifications are effectively ignored until their handlers are set to a
// concrete callback.
type NotificationHandlers struct {
	// OnClientConnected is invoked when the client connects or reconnects
	// to the RPC server. This callback is run async with the rest of the
	// notification handlers, and is safe for blocking client requests.
	OnClientConnected func()

	// OnBlockAdded is invoked when a block is connected to the DAG.
	// It will only be invoked if a preceding call to NotifyBlocks has been made
	// to register for the notification and the function is non-nil.
	//
	// NOTE: Deprecated. Use OnFilteredBlockAdded instead.
	OnBlockAdded func(hash *daghash.Hash, height int32, t mstime.Time)

	// OnFilteredBlockAdded is invoked when a block is connected to the
	// bloackDAG. It will only be invoked if a preceding call to
	// NotifyBlocks has been made to register for the notification and the
	// function is non-nil. Its parameters differ from OnBlockAdded: it
	// receives the block's blueScore, header, and relevant transactions.
	OnFilteredBlockAdded func(blueScore uint64, header *wire.BlockHeader,
		txs []*util.Tx)

	// OnChainChanged is invoked when the selected parent chain of the
	// DAG had changed. It will only be invoked if a preceding call to
	// NotifyChainChanges has been made to register for the notification and the
	// function is non-nil.
	OnChainChanged func(removedChainBlockHashes []*daghash.Hash,
		addedChainBlocks []*ChainBlock)

	// OnRelevantTxAccepted is invoked when an unmined transaction passes
	// the client's transaction filter.
	OnRelevantTxAccepted func(transaction []byte)

	// OnTxAccepted is invoked when a transaction is accepted into the
	// memory pool. It will only be invoked if a preceding call to
	// NotifyNewTransactions with the verbose flag set to false has been
	// made to register for the notification and the function is non-nil.
	OnTxAccepted func(hash *daghash.Hash, amount util.Amount)

	// OnTxAccepted is invoked when a transaction is accepted into the
	// memory pool. It will only be invoked if a preceding call to
	// NotifyNewTransactions with the verbose flag set to true has been
	// made to register for the notification and the function is non-nil.
	OnTxAcceptedVerbose func(txDetails *rpcmodel.TxRawResult)

	// OnUnknownNotification is invoked when an unrecognized notification
	// is received. This typically means the notification handling code
	// for this package needs to be updated for a new notification type or
	// the caller is using a custom notification this package does not know
	// about.
	OnUnknownNotification func(method string, params []json.RawMessage)
}

// handleNotification examines the passed notification type, performs
// conversions to get the raw notification types into higher level types and
// delivers the notification to the appropriate On<X> handler registered with
// the client.
func (c *Client) handleNotification(ntfn *rawNotification) {
	// Ignore the notification if the client is not interested in any
	// notifications.
	if c.ntfnHandlers == nil {
		return
	}

	switch ntfn.Method {

	// ChainChangedNtfnMethod
	case rpcmodel.ChainChangedNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnChainChanged == nil {
			return
		}

		removedChainBlockHashes, addedChainBlocks, err := parseChainChangedParams(ntfn.Params)
		if err != nil {
			log.Warnf("Received invalid chain changed "+
				"notification: %s", err)
			return
		}

		c.ntfnHandlers.OnChainChanged(removedChainBlockHashes, addedChainBlocks)

	// OnFilteredBlockAdded
	case rpcmodel.FilteredBlockAddedNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnFilteredBlockAdded == nil {
			return
		}

		blockBlueScore, blockHeader, transactions, err :=
			parseFilteredBlockAddedParams(ntfn.Params)
		if err != nil {
			log.Warnf("Received invalid filtered block "+
				"connected notification: %s", err)
			return
		}

		c.ntfnHandlers.OnFilteredBlockAdded(blockBlueScore,
			blockHeader, transactions)

	// OnRelevantTxAccepted
	case rpcmodel.RelevantTxAcceptedNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnRelevantTxAccepted == nil {
			return
		}

		transaction, err := parseRelevantTxAcceptedParams(ntfn.Params)
		if err != nil {
			log.Warnf("Received invalid relevanttxaccepted "+
				"notification: %s", err)
			return
		}

		c.ntfnHandlers.OnRelevantTxAccepted(transaction)

	// OnTxAccepted
	case rpcmodel.TxAcceptedNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnTxAccepted == nil {
			return
		}

		hash, amt, err := parseTxAcceptedNtfnParams(ntfn.Params)
		if err != nil {
			log.Warnf("Received invalid tx accepted "+
				"notification: %s", err)
			return
		}

		c.ntfnHandlers.OnTxAccepted(hash, amt)

	// OnTxAcceptedVerbose
	case rpcmodel.TxAcceptedVerboseNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnTxAcceptedVerbose == nil {
			return
		}

		rawTx, err := parseTxAcceptedVerboseNtfnParams(ntfn.Params)
		if err != nil {
			log.Warnf("Received invalid tx accepted verbose "+
				"notification: %s", err)
			return
		}

		c.ntfnHandlers.OnTxAcceptedVerbose(rawTx)

	// OnUnknownNotification
	default:
		if c.ntfnHandlers.OnUnknownNotification == nil {
			return
		}

		c.ntfnHandlers.OnUnknownNotification(ntfn.Method, ntfn.Params)
	}
}

// wrongNumParams is an error type describing an unparseable JSON-RPC
// notificiation due to an incorrect number of parameters for the
// expected notification type. The value is the number of parameters
// of the invalid notification.
type wrongNumParams int

// Error satisifies the builtin error interface.
func (e wrongNumParams) Error() string {
	return fmt.Sprintf("wrong number of parameters (%d)", e)
}

// ChainBlock models a block that is part of the selected parent chain.
type ChainBlock struct {
	Hash           *daghash.Hash
	AcceptedBlocks []*AcceptedBlock
}

// AcceptedBlock models a block that is included in the blues of a selected
// chain block.
type AcceptedBlock struct {
	Hash          *daghash.Hash
	AcceptedTxIDs []*daghash.TxID
}

func parseChainChangedParams(params []json.RawMessage) (removedChainBlockHashes []*daghash.Hash, addedChainBlocks []*ChainBlock,
	err error) {

	if len(params) != 1 {
		return nil, nil, wrongNumParams(len(params))
	}

	// Unmarshal first parameter as a raw transaction result object.
	var rawParam rpcmodel.ChainChangedRawParam
	err = json.Unmarshal(params[0], &rawParam)
	if err != nil {
		return nil, nil, err
	}

	removedChainBlockHashes = make([]*daghash.Hash, len(rawParam.RemovedChainBlockHashes))
	for i, hashStr := range rawParam.RemovedChainBlockHashes {
		hash, err := daghash.NewHashFromStr(hashStr)
		if err != nil {
			return nil, nil, err
		}
		removedChainBlockHashes[i] = hash
	}

	addedChainBlocks = make([]*ChainBlock, len(rawParam.AddedChainBlocks))
	for i, jsonChainBlock := range rawParam.AddedChainBlocks {
		chainBlock := &ChainBlock{
			AcceptedBlocks: make([]*AcceptedBlock, len(jsonChainBlock.AcceptedBlocks)),
		}
		hash, err := daghash.NewHashFromStr(jsonChainBlock.Hash)
		if err != nil {
			return nil, nil, err
		}
		chainBlock.Hash = hash
		for j, jsonAcceptedBlock := range jsonChainBlock.AcceptedBlocks {
			acceptedBlock := &AcceptedBlock{
				AcceptedTxIDs: make([]*daghash.TxID, len(jsonAcceptedBlock.AcceptedTxIDs)),
			}
			hash, err := daghash.NewHashFromStr(jsonAcceptedBlock.Hash)
			if err != nil {
				return nil, nil, err
			}
			acceptedBlock.Hash = hash
			for k, txIDStr := range jsonAcceptedBlock.AcceptedTxIDs {
				txID, err := daghash.NewTxIDFromStr(txIDStr)
				if err != nil {
					return nil, nil, err
				}
				acceptedBlock.AcceptedTxIDs[k] = txID
			}
			chainBlock.AcceptedBlocks[j] = acceptedBlock
		}
		addedChainBlocks[i] = chainBlock
	}

	return removedChainBlockHashes, addedChainBlocks, nil
}

// parseFilteredBlockAddedParams parses out the parameters included in a
// filteredblockadded notification.
func parseFilteredBlockAddedParams(params []json.RawMessage) (uint64,
	*wire.BlockHeader, []*util.Tx, error) {

	if len(params) < 3 {
		return 0, nil, nil, wrongNumParams(len(params))
	}

	// Unmarshal first parameter as an integer.
	var blockHeight uint64
	err := json.Unmarshal(params[0], &blockHeight)
	if err != nil {
		return 0, nil, nil, err
	}

	// Unmarshal second parameter as a slice of bytes.
	blockHeaderBytes, err := parseHexParam(params[1])
	if err != nil {
		return 0, nil, nil, err
	}

	// Deserialize block header from slice of bytes.
	var blockHeader wire.BlockHeader
	err = blockHeader.Deserialize(bytes.NewReader(blockHeaderBytes))
	if err != nil {
		return 0, nil, nil, err
	}

	// Unmarshal third parameter as a slice of hex-encoded strings.
	var hexTransactions []string
	err = json.Unmarshal(params[2], &hexTransactions)
	if err != nil {
		return 0, nil, nil, err
	}

	// Create slice of transactions from slice of strings by hex-decoding.
	transactions := make([]*util.Tx, len(hexTransactions))
	for i, hexTx := range hexTransactions {
		transaction, err := hex.DecodeString(hexTx)
		if err != nil {
			return 0, nil, nil, err
		}

		transactions[i], err = util.NewTxFromBytes(transaction)
		if err != nil {
			return 0, nil, nil, err
		}
	}

	return blockHeight, &blockHeader, transactions, nil
}

func parseHexParam(param json.RawMessage) ([]byte, error) {
	var s string
	err := json.Unmarshal(param, &s)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(s)
}

// parseRelevantTxAcceptedParams parses out the parameter included in a
// relevanttxaccepted notification.
func parseRelevantTxAcceptedParams(params []json.RawMessage) (transaction []byte, err error) {
	if len(params) < 1 {
		return nil, wrongNumParams(len(params))
	}

	return parseHexParam(params[0])
}

// parseTxAcceptedNtfnParams parses out the transaction hash and total amount
// from the parameters of a txaccepted notification.
func parseTxAcceptedNtfnParams(params []json.RawMessage) (*daghash.Hash,
	util.Amount, error) {

	if len(params) != 2 {
		return nil, 0, wrongNumParams(len(params))
	}

	// Unmarshal first parameter as a string.
	var txHashStr string
	err := json.Unmarshal(params[0], &txHashStr)
	if err != nil {
		return nil, 0, err
	}

	// Unmarshal second parameter as a floating point number.
	var famt float64
	err = json.Unmarshal(params[1], &famt)
	if err != nil {
		return nil, 0, err
	}

	// Bounds check amount.
	amt, err := util.NewAmount(famt)
	if err != nil {
		return nil, 0, err
	}

	// Decode string encoding of transaction sha.
	txHash, err := daghash.NewHashFromStr(txHashStr)
	if err != nil {
		return nil, 0, err
	}

	return txHash, amt, nil
}

// parseTxAcceptedVerboseNtfnParams parses out details about a raw transaction
// from the parameters of a txacceptedverbose notification.
func parseTxAcceptedVerboseNtfnParams(params []json.RawMessage) (*rpcmodel.TxRawResult,
	error) {

	if len(params) != 1 {
		return nil, wrongNumParams(len(params))
	}

	// Unmarshal first parameter as a raw transaction result object.
	var rawTx rpcmodel.TxRawResult
	err := json.Unmarshal(params[0], &rawTx)
	if err != nil {
		return nil, err
	}

	// TODO: change txacceptedverbose notification callbacks to use nicer
	// types for all details about the transaction (i.e. decoding hashes
	// from their string encoding).
	return &rawTx, nil
}

// FutureNotifyBlocksResult is a future promise to deliver the result of a
// NotifyBlocksAsync RPC invocation (or an applicable error).
type FutureNotifyBlocksResult chan *response

// Receive waits for the response promised by the future and returns an error
// if the registration was not successful.
func (r FutureNotifyBlocksResult) Receive() error {
	_, err := receiveFuture(r)
	return err
}

// NotifyBlocksAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See NotifyBlocks for the blocking version and more details.
func (c *Client) NotifyBlocksAsync() FutureNotifyBlocksResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	// Ignore the notification if the client is not interested in
	// notifications.
	if c.ntfnHandlers == nil {
		return newNilFutureResult()
	}

	cmd := rpcmodel.NewNotifyBlocksCmd()
	return c.sendCmd(cmd)
}

// NotifyBlocks registers the client to receive notifications when blocks are
// connected to the DAG. The notifications are delivered to the notification
// handlers associated with the client. Calling this function has no effect
// if there are no notification handlers and will result in an error if the
// client is configured to run in HTTP POST mode.
//
// The notifications delivered as a result of this call will be via OnBlockAdded
func (c *Client) NotifyBlocks() error {
	return c.NotifyBlocksAsync().Receive()
}

// FutureNotifyChainChangesResult is a future promise to deliver the result of a
// NotifyChainChangesAsync RPC invocation (or an applicable error).
type FutureNotifyChainChangesResult chan *response

// Receive waits for the response promised by the future and returns an error
// if the registration was not successful.
func (r FutureNotifyChainChangesResult) Receive() error {
	_, err := receiveFuture(r)
	return err
}

// NotifyChainChangesAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See NotifyChainChanges for the blocking version and more details.
func (c *Client) NotifyChainChangesAsync() FutureNotifyBlocksResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	// Ignore the notification if the client is not interested in
	// notifications.
	if c.ntfnHandlers == nil {
		return newNilFutureResult()
	}

	cmd := rpcmodel.NewNotifyChainChangesCmd()
	return c.sendCmd(cmd)
}

// NotifyChainChanges registers the client to receive notifications when the
// selected parent chain changes. The notifications are delivered to the
// notification handlers associated with the client. Calling this function has
// no effect if there are no notification handlers and will result in an error
// if the client is configured to run in HTTP POST mode.
//
// The notifications delivered as a result of this call will be via OnBlockAdded
func (c *Client) NotifyChainChanges() error {
	return c.NotifyChainChangesAsync().Receive()
}

// FutureNotifyNewTransactionsResult is a future promise to deliver the result
// of a NotifyNewTransactionsAsync RPC invocation (or an applicable error).
type FutureNotifyNewTransactionsResult chan *response

// Receive waits for the response promised by the future and returns an error
// if the registration was not successful.
func (r FutureNotifyNewTransactionsResult) Receive() error {
	_, err := receiveFuture(r)
	return err
}

// NotifyNewTransactionsAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See NotifyNewTransactionsAsync for the blocking version and more details.
func (c *Client) NotifyNewTransactionsAsync(verbose bool, subnetworkID *string) FutureNotifyNewTransactionsResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	// Ignore the notification if the client is not interested in
	// notifications.
	if c.ntfnHandlers == nil {
		return newNilFutureResult()
	}

	cmd := rpcmodel.NewNotifyNewTransactionsCmd(&verbose, subnetworkID)
	return c.sendCmd(cmd)
}

// NotifyNewTransactions registers the client to receive notifications every
// time a new transaction is accepted to the memory pool. The notifications are
// delivered to the notification handlers associated with the client. Calling
// this function has no effect if there are no notification handlers and will
// result in an error if the client is configured to run in HTTP POST mode.
//
// The notifications delivered as a result of this call will be via one of
// OnTxAccepted (when verbose is false) or OnTxAcceptedVerbose (when verbose is
// true).
func (c *Client) NotifyNewTransactions(verbose bool, subnetworkID *string) error {
	return c.NotifyNewTransactionsAsync(verbose, subnetworkID).Receive()
}

// FutureLoadTxFilterResult is a future promise to deliver the result
// of a LoadTxFilterAsync RPC invocation (or an applicable error).
type FutureLoadTxFilterResult chan *response

// Receive waits for the response promised by the future and returns an error
// if the registration was not successful.
func (r FutureLoadTxFilterResult) Receive() error {
	_, err := receiveFuture(r)
	return err
}

// LoadTxFilterAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See LoadTxFilter for the blocking version and more details.
func (c *Client) LoadTxFilterAsync(reload bool, addresses []util.Address,
	outpoints []wire.Outpoint) FutureLoadTxFilterResult {

	addrStrs := make([]string, len(addresses))
	for i, a := range addresses {
		addrStrs[i] = a.EncodeAddress()
	}
	outpointObjects := make([]rpcmodel.Outpoint, len(outpoints))
	for i := range outpoints {
		outpointObjects[i] = rpcmodel.Outpoint{
			TxID:  outpoints[i].TxID.String(),
			Index: outpoints[i].Index,
		}
	}

	cmd := rpcmodel.NewLoadTxFilterCmd(reload, addrStrs, outpointObjects)
	return c.sendCmd(cmd)
}

// LoadTxFilter loads, reloads, or adds data to a websocket client's transaction
// filter. The filter is consistently updated based on inspected transactions
// during mempool acceptance, block acceptance, and for all rescanned blocks.
func (c *Client) LoadTxFilter(reload bool, addresses []util.Address, outpoints []wire.Outpoint) error {
	return c.LoadTxFilterAsync(reload, addresses, outpoints).Receive()
}
