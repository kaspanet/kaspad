// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcclient

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
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
	notifyNewTx             bool
	notifyNewTxVerbose      bool
	notifyNewTxSubnetworkID *string
}

// Copy returns a deep copy of the receiver.
func (s *notificationState) Copy() *notificationState {
	var stateCopy notificationState
	stateCopy.notifyBlocks = s.notifyBlocks
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
// result waiting on the channel with the reply set to nil.  This is useful
// to ignore things such as notifications when the caller didn't specify any
// notification handlers.
func newNilFutureResult() chan *response {
	responseChan := make(chan *response, 1)
	responseChan <- &response{result: nil, err: nil}
	return responseChan
}

// NotificationHandlers defines callback function pointers to invoke with
// notifications.  Since all of the functions are nil by default, all
// notifications are effectively ignored until their handlers are set to a
// concrete callback.
type NotificationHandlers struct {
	// OnClientConnected is invoked when the client connects or reconnects
	// to the RPC server.  This callback is run async with the rest of the
	// notification handlers, and is safe for blocking client requests.
	OnClientConnected func()

	// OnBlockAdded is invoked when a block is connected to the DAG.
	// It will only be invoked if a preceding call to NotifyBlocks has been made
	// to register for the notification and the function is non-nil.
	//
	// NOTE: Deprecated. Use OnFilteredBlockAdded instead.
	OnBlockAdded func(hash *daghash.Hash, height int32, t time.Time)

	// OnFilteredBlockAdded is invoked when a block is connected to the
	// bloackDAG.  It will only be invoked if a preceding call to
	// NotifyBlocks has been made to register for the notification and the
	// function is non-nil.  Its parameters differ from OnBlockAdded: it
	// receives the block's height, header, and relevant transactions.
	OnFilteredBlockAdded func(height uint64, header *wire.BlockHeader,
		txs []*util.Tx)

	// OnRelevantTxAccepted is invoked when an unmined transaction passes
	// the client's transaction filter.
	//
	// NOTE: This is a btcsuite extension ported from
	// github.com/decred/dcrrpcclient.
	OnRelevantTxAccepted func(transaction []byte)

	// OnTxAccepted is invoked when a transaction is accepted into the
	// memory pool.  It will only be invoked if a preceding call to
	// NotifyNewTransactions with the verbose flag set to false has been
	// made to register for the notification and the function is non-nil.
	OnTxAccepted func(hash *daghash.Hash, amount util.Amount)

	// OnTxAccepted is invoked when a transaction is accepted into the
	// memory pool.  It will only be invoked if a preceding call to
	// NotifyNewTransactions with the verbose flag set to true has been
	// made to register for the notification and the function is non-nil.
	OnTxAcceptedVerbose func(txDetails *btcjson.TxRawResult)

	// OnBtcdConnected is invoked when a wallet connects or disconnects from
	// btcd.
	//
	// This will only be available when client is connected to a wallet
	// server such as btcwallet.
	OnBtcdConnected func(connected bool)

	// OnAccountBalance is invoked with account balance updates.
	//
	// This will only be available when speaking to a wallet server
	// such as btcwallet.
	OnAccountBalance func(account string, balance util.Amount, confirmed bool)

	// OnWalletLockState is invoked when a wallet is locked or unlocked.
	//
	// This will only be available when client is connected to a wallet
	// server such as btcwallet.
	OnWalletLockState func(locked bool)

	// OnUnknownNotification is invoked when an unrecognized notification
	// is received.  This typically means the notification handling code
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

	// OnFilteredBlockAdded
	case btcjson.FilteredBlockAddedNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnFilteredBlockAdded == nil {
			return
		}

		blockHeight, blockHeader, transactions, err :=
			parseFilteredBlockAddedParams(ntfn.Params)
		if err != nil {
			log.Warnf("Received invalid filtered block "+
				"connected notification: %s", err)
			return
		}

		c.ntfnHandlers.OnFilteredBlockAdded(blockHeight,
			blockHeader, transactions)

	// OnRelevantTxAccepted
	case btcjson.RelevantTxAcceptedNtfnMethod:
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
	case btcjson.TxAcceptedNtfnMethod:
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
	case btcjson.TxAcceptedVerboseNtfnMethod:
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

	// OnBtcdConnected
	case btcjson.BtcdConnectedNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnBtcdConnected == nil {
			return
		}

		connected, err := parseBtcdConnectedNtfnParams(ntfn.Params)
		if err != nil {
			log.Warnf("Received invalid btcd connected "+
				"notification: %s", err)
			return
		}

		c.ntfnHandlers.OnBtcdConnected(connected)

	// OnAccountBalance
	case btcjson.AccountBalanceNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnAccountBalance == nil {
			return
		}

		account, bal, conf, err := parseAccountBalanceNtfnParams(ntfn.Params)
		if err != nil {
			log.Warnf("Received invalid account balance "+
				"notification: %s", err)
			return
		}

		c.ntfnHandlers.OnAccountBalance(account, bal, conf)

	// OnWalletLockState
	case btcjson.WalletLockStateNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnWalletLockState == nil {
			return
		}

		// The account name is not notified, so the return value is
		// discarded.
		_, locked, err := parseWalletLockStateNtfnParams(ntfn.Params)
		if err != nil {
			log.Warnf("Received invalid wallet lock state "+
				"notification: %s", err)
			return
		}

		c.ntfnHandlers.OnWalletLockState(locked)

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
// expected notification type.  The value is the number of parameters
// of the invalid notification.
type wrongNumParams int

// Error satisifies the builtin error interface.
func (e wrongNumParams) Error() string {
	return fmt.Sprintf("wrong number of parameters (%d)", e)
}

// parseDAGNtfnParams parses out the block hash and height from the parameters
// of blockadded.
func parseDAGNtfnParams(params []json.RawMessage) (*daghash.Hash,
	int32, time.Time, error) {

	if len(params) != 3 {
		return nil, 0, time.Time{}, wrongNumParams(len(params))
	}

	// Unmarshal first parameter as a string.
	var blockHashStr string
	err := json.Unmarshal(params[0], &blockHashStr)
	if err != nil {
		return nil, 0, time.Time{}, err
	}

	// Unmarshal second parameter as an integer.
	var blockHeight int32
	err = json.Unmarshal(params[1], &blockHeight)
	if err != nil {
		return nil, 0, time.Time{}, err
	}

	// Unmarshal third parameter as unix time.
	var blockTimeUnix int64
	err = json.Unmarshal(params[2], &blockTimeUnix)
	if err != nil {
		return nil, 0, time.Time{}, err
	}

	// Create hash from block hash string.
	blockHash, err := daghash.NewHashFromStr(blockHashStr)
	if err != nil {
		return nil, 0, time.Time{}, err
	}

	// Create time.Time from unix time.
	blockTime := time.Unix(blockTimeUnix, 0)

	return blockHash, blockHeight, blockTime, nil
}

// parseFilteredBlockAddedParams parses out the parameters included in a
// filteredblockadded notification.
//
// NOTE: This is a btcd extension ported from github.com/decred/dcrrpcclient
// and requires a websocket connection.
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

// parseChainTxNtfnParams parses out the transaction and optional details about
// the block it's mined in from the parameters of recvtx and redeemingtx
// notifications.
func parseChainTxNtfnParams(params []json.RawMessage) (*util.Tx,
	*btcjson.BlockDetails, error) {

	if len(params) == 0 || len(params) > 2 {
		return nil, nil, wrongNumParams(len(params))
	}

	// Unmarshal first parameter as a string.
	var txHex string
	err := json.Unmarshal(params[0], &txHex)
	if err != nil {
		return nil, nil, err
	}

	// If present, unmarshal second optional parameter as the block details
	// JSON object.
	var block *btcjson.BlockDetails
	if len(params) > 1 {
		err = json.Unmarshal(params[1], &block)
		if err != nil {
			return nil, nil, err
		}
	}

	// Hex decode and deserialize the transaction.
	serializedTx, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, nil, err
	}
	var msgTx wire.MsgTx
	err = msgTx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, nil, err
	}

	// TODO: Change recvtx and redeemingtx callback signatures to use
	// nicer types for details about the block (block hash as a
	// daghash.Hash, block time as a time.Time, etc.).
	return util.NewTx(&msgTx), block, nil
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
func parseTxAcceptedVerboseNtfnParams(params []json.RawMessage) (*btcjson.TxRawResult,
	error) {

	if len(params) != 1 {
		return nil, wrongNumParams(len(params))
	}

	// Unmarshal first parameter as a raw transaction result object.
	var rawTx btcjson.TxRawResult
	err := json.Unmarshal(params[0], &rawTx)
	if err != nil {
		return nil, err
	}

	// TODO: change txacceptedverbose notification callbacks to use nicer
	// types for all details about the transaction (i.e. decoding hashes
	// from their string encoding).
	return &rawTx, nil
}

// parseBtcdConnectedNtfnParams parses out the connection status of btcd
// and btcwallet from the parameters of a btcdconnected notification.
func parseBtcdConnectedNtfnParams(params []json.RawMessage) (bool, error) {
	if len(params) != 1 {
		return false, wrongNumParams(len(params))
	}

	// Unmarshal first parameter as a boolean.
	var connected bool
	err := json.Unmarshal(params[0], &connected)
	if err != nil {
		return false, err
	}

	return connected, nil
}

// parseAccountBalanceNtfnParams parses out the account name, total balance,
// and whether or not the balance is confirmed or unconfirmed from the
// parameters of an accountbalance notification.
func parseAccountBalanceNtfnParams(params []json.RawMessage) (account string,
	balance util.Amount, confirmed bool, err error) {

	if len(params) != 3 {
		return "", 0, false, wrongNumParams(len(params))
	}

	// Unmarshal first parameter as a string.
	err = json.Unmarshal(params[0], &account)
	if err != nil {
		return "", 0, false, err
	}

	// Unmarshal second parameter as a floating point number.
	var fbal float64
	err = json.Unmarshal(params[1], &fbal)
	if err != nil {
		return "", 0, false, err
	}

	// Unmarshal third parameter as a boolean.
	err = json.Unmarshal(params[2], &confirmed)
	if err != nil {
		return "", 0, false, err
	}

	// Bounds check amount.
	bal, err := util.NewAmount(fbal)
	if err != nil {
		return "", 0, false, err
	}

	return account, bal, confirmed, nil
}

// parseWalletLockStateNtfnParams parses out the account name and locked
// state of an account from the parameters of a walletlockstate notification.
func parseWalletLockStateNtfnParams(params []json.RawMessage) (account string,
	locked bool, err error) {

	if len(params) != 2 {
		return "", false, wrongNumParams(len(params))
	}

	// Unmarshal first parameter as a string.
	err = json.Unmarshal(params[0], &account)
	if err != nil {
		return "", false, err
	}

	// Unmarshal second parameter as a boolean.
	err = json.Unmarshal(params[1], &locked)
	if err != nil {
		return "", false, err
	}

	return account, locked, nil
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
//
// NOTE: This is a btcd extension and requires a websocket connection.
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

	cmd := btcjson.NewNotifyBlocksCmd()
	return c.sendCmd(cmd)
}

// NotifyBlocks registers the client to receive notifications when blocks are
// connected and disconnected from the main chain.  The notifications are
// delivered to the notification handlers associated with the client.  Calling
// this function has no effect if there are no notification handlers and will
// result in an error if the client is configured to run in HTTP POST mode.
//
// The notifications delivered as a result of this call will be via OnBlockAdded
//
// NOTE: This is a btcd extension and requires a websocket connection.
func (c *Client) NotifyBlocks() error {
	return c.NotifyBlocksAsync().Receive()
}

// newOutPointFromWire constructs the btcjson representation of a transaction
// outpoint from the wire type.
func newOutPointFromWire(op *wire.OutPoint) btcjson.OutPoint {
	return btcjson.OutPoint{
		TxID:  op.TxID.String(),
		Index: op.Index,
	}
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
//
// NOTE: This is a btcd extension and requires a websocket connection.
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

	cmd := btcjson.NewNotifyNewTransactionsCmd(&verbose, subnetworkID)
	return c.sendCmd(cmd)
}

// NotifyNewTransactions registers the client to receive notifications every
// time a new transaction is accepted to the memory pool.  The notifications are
// delivered to the notification handlers associated with the client.  Calling
// this function has no effect if there are no notification handlers and will
// result in an error if the client is configured to run in HTTP POST mode.
//
// The notifications delivered as a result of this call will be via one of
// OnTxAccepted (when verbose is false) or OnTxAcceptedVerbose (when verbose is
// true).
//
// NOTE: This is a btcd extension and requires a websocket connection.
func (c *Client) NotifyNewTransactions(verbose bool, subnetworkID *string) error {
	return c.NotifyNewTransactionsAsync(verbose, subnetworkID).Receive()
}

// FutureLoadTxFilterResult is a future promise to deliver the result
// of a LoadTxFilterAsync RPC invocation (or an applicable error).
//
// NOTE: This is a btcd extension ported from github.com/decred/dcrrpcclient
// and requires a websocket connection.
type FutureLoadTxFilterResult chan *response

// Receive waits for the response promised by the future and returns an error
// if the registration was not successful.
//
// NOTE: This is a btcd extension ported from github.com/decred/dcrrpcclient
// and requires a websocket connection.
func (r FutureLoadTxFilterResult) Receive() error {
	_, err := receiveFuture(r)
	return err
}

// LoadTxFilterAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See LoadTxFilter for the blocking version and more details.
//
// NOTE: This is a btcd extension ported from github.com/decred/dcrrpcclient
// and requires a websocket connection.
func (c *Client) LoadTxFilterAsync(reload bool, addresses []util.Address,
	outPoints []wire.OutPoint) FutureLoadTxFilterResult {

	addrStrs := make([]string, len(addresses))
	for i, a := range addresses {
		addrStrs[i] = a.EncodeAddress()
	}
	outPointObjects := make([]btcjson.OutPoint, len(outPoints))
	for i := range outPoints {
		outPointObjects[i] = btcjson.OutPoint{
			TxID:  outPoints[i].TxID.String(),
			Index: outPoints[i].Index,
		}
	}

	cmd := btcjson.NewLoadTxFilterCmd(reload, addrStrs, outPointObjects)
	return c.sendCmd(cmd)
}

// LoadTxFilter loads, reloads, or adds data to a websocket client's transaction
// filter.  The filter is consistently updated based on inspected transactions
// during mempool acceptance, block acceptance, and for all rescanned blocks.
//
// NOTE: This is a btcd extension ported from github.com/decred/dcrrpcclient
// and requires a websocket connection.
func (c *Client) LoadTxFilter(reload bool, addresses []util.Address, outPoints []wire.OutPoint) error {
	return c.LoadTxFilterAsync(reload, addresses, outPoints).Receive()
}
