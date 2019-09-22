// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpc

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/btcsuite/websocket"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/blockdag/indexers"
	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/config"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/logger"
	"github.com/daglabs/btcd/mempool"
	"github.com/daglabs/btcd/mining"
	"github.com/daglabs/btcd/mining/cpuminer"
	"github.com/daglabs/btcd/peer"
	"github.com/daglabs/btcd/server/p2p"
	"github.com/daglabs/btcd/server/serverutils"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/util/fs"
	"github.com/daglabs/btcd/util/network"
	"github.com/daglabs/btcd/util/random"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/version"
	"github.com/daglabs/btcd/wire"
)

// API version constants
const (
	jsonrpcSemverString = "1.3.0"
	jsonrpcSemverMajor  = 1
	jsonrpcSemverMinor  = 3
	jsonrpcSemverPatch  = 0
)

const (
	// rpcAuthTimeoutSeconds is the number of seconds a connection to the
	// RPC server is allowed to stay open without authenticating before it
	// is closed.
	rpcAuthTimeoutSeconds = 10

	// maxProtocolVersion is the max protocol version the server supports.
	maxProtocolVersion = 70002
)

// Errors
var (
	// ErrRPCUnimplemented is an error returned to RPC clients when the
	// provided command is recognized, but not implemented.
	ErrRPCUnimplemented = &btcjson.RPCError{
		Code:    btcjson.ErrRPCUnimplemented,
		Message: "Command unimplemented",
	}
)

type commandHandler func(*Server, interface{}, <-chan struct{}) (interface{}, error)

// rpcHandlers maps RPC command strings to appropriate handler functions.
// This is set by init because help references rpcHandlers and thus causes
// a dependency loop.
var rpcHandlers map[string]commandHandler
var rpcHandlersBeforeInit = map[string]commandHandler{
	"addManualNode":         handleAddManualNode,
	"createRawTransaction":  handleCreateRawTransaction,
	"debugLevel":            handleDebugLevel,
	"decodeRawTransaction":  handleDecodeRawTransaction,
	"decodeScript":          handleDecodeScript,
	"generate":              handleGenerate,
	"getAllManualNodesInfo": handleGetAllManualNodesInfo,
	"getBestBlock":          handleGetBestBlock,
	"getBestBlockHash":      handleGetBestBlockHash,
	"getBlock":              handleGetBlock,
	"getBlockDagInfo":       handleGetBlockDAGInfo,
	"getBlockCount":         handleGetBlockCount,
	"getBlockHash":          handleGetBlockHash,
	"getBlockHeader":        handleGetBlockHeader,
	"getBlockTemplate":      handleGetBlockTemplate,
	"getCFilter":            handleGetCFilter,
	"getCFilterHeader":      handleGetCFilterHeader,
	"getChainFromBlock":     handleGetChainFromBlock,
	"getConnectionCount":    handleGetConnectionCount,
	"getCurrentNet":         handleGetCurrentNet,
	"getDifficulty":         handleGetDifficulty,
	"getGenerate":           handleGetGenerate,
	"getHashesPerSec":       handleGetHashesPerSec,
	"getHeaders":            handleGetHeaders,
	"getTopHeaders":         handleGetTopHeaders,
	"getInfo":               handleGetInfo,
	"getManualNodeInfo":     handleGetManualNodeInfo,
	"getMempoolInfo":        handleGetMempoolInfo,
	"getMiningInfo":         handleGetMiningInfo,
	"getNetTotals":          handleGetNetTotals,
	"getNetworkHashPs":      handleGetNetworkHashPS,
	"getPeerInfo":           handleGetPeerInfo,
	"getRawMempool":         handleGetRawMempool,
	"getRawTransaction":     handleGetRawTransaction,
	"getSubnetwork":         handleGetSubnetwork,
	"getTxOut":              handleGetTxOut,
	"help":                  handleHelp,
	"node":                  handleNode,
	"ping":                  handlePing,
	"removeManualNode":      handleRemoveManualNode,
	"searchRawTransactions": handleSearchRawTransactions,
	"sendRawTransaction":    handleSendRawTransaction,
	"setGenerate":           handleSetGenerate,
	"stop":                  handleStop,
	"submitBlock":           handleSubmitBlock,
	"uptime":                handleUptime,
	"validateAddress":       handleValidateAddress,
	"verifyMessage":         handleVerifyMessage,
	"version":               handleVersion,
	"flushDbCache":          handleFlushDBCache,
}

// Commands that are currently unimplemented, but should ultimately be.
var rpcUnimplemented = map[string]struct{}{
	"estimatePriority": {},
	"getChainTips":     {},
	"getMempoolEntry":  {},
	"getNetworkInfo":   {},
	"invalidateBlock":  {},
	"preciousBlock":    {},
	"reconsiderBlock":  {},
}

// Commands that are available to a limited user
var rpcLimited = map[string]struct{}{
	// Websockets commands
	"loadTxFilter":          {},
	"notifyBlocks":          {},
	"notifyChainChanges":    {},
	"notifyNewTransactions": {},
	"notifyReceived":        {},
	"notifySpent":           {},
	"rescan":                {},
	"rescanBlocks":          {},
	"session":               {},

	// Websockets AND HTTP/S commands
	"help": {},

	// HTTP/S-only commands
	"createRawTransaction":  {},
	"decodeRawTransaction":  {},
	"decodeScript":          {},
	"getBestBlock":          {},
	"getBestBlockHash":      {},
	"getBlock":              {},
	"getBlockCount":         {},
	"getBlockHash":          {},
	"getBlockHeader":        {},
	"getCFilter":            {},
	"getCFilterHeader":      {},
	"getChainFromBlock":     {},
	"getCurrentNet":         {},
	"getDifficulty":         {},
	"getHeaders":            {},
	"getInfo":               {},
	"getNetTotals":          {},
	"getNetworkHashPs":      {},
	"getRawMempool":         {},
	"getRawTransaction":     {},
	"getTxOut":              {},
	"searchRawTransactions": {},
	"sendRawTransaction":    {},
	"submitBlock":           {},
	"uptime":                {},
	"validateAddress":       {},
	"verifyMessage":         {},
	"version":               {},
}

// internalRPCError is a convenience function to convert an internal error to
// an RPC error with the appropriate code set.  It also logs the error to the
// RPC server subsystem since internal errors really should not occur.  The
// context parameter is only used in the log message and may be empty if it's
// not needed.
func internalRPCError(errStr, context string) *btcjson.RPCError {
	logStr := errStr
	if context != "" {
		logStr = context + ": " + errStr
	}
	log.Error(logStr)
	return btcjson.NewRPCError(btcjson.ErrRPCInternal.Code, errStr)
}

// rpcDecodeHexError is a convenience function for returning a nicely formatted
// RPC error which indicates the provided hex string failed to decode.
func rpcDecodeHexError(gotHex string) *btcjson.RPCError {
	return btcjson.NewRPCError(btcjson.ErrRPCDecodeHexString,
		fmt.Sprintf("Argument must be hexadecimal string (not %q)",
			gotHex))
}

// rpcNoTxInfoError is a convenience function for returning a nicely formatted
// RPC error which indicates there is no information available for the provided
// transaction hash.
func rpcNoTxInfoError(txID *daghash.TxID) *btcjson.RPCError {
	return btcjson.NewRPCError(btcjson.ErrRPCNoTxInfo,
		fmt.Sprintf("No information available about transaction %s",
			txID))
}

// handleUnimplemented is the handler for commands that should ultimately be
// supported but are not yet implemented.
func handleUnimplemented(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return nil, ErrRPCUnimplemented
}

// handleAddManualNode handles addManualNode commands.
func handleAddManualNode(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.AddManualNodeCmd)

	oneTry := c.OneTry != nil && *c.OneTry

	addr := network.NormalizeAddress(c.Addr, s.cfg.DAGParams.DefaultPort)
	var err error
	if oneTry {
		err = s.cfg.ConnMgr.Connect(addr, false)
	} else {
		err = s.cfg.ConnMgr.Connect(addr, true)
	}

	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}

	// no data returned unless an error.
	return nil, nil
}

// handleRemoveManualNode handles removeManualNode command.
func handleRemoveManualNode(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.RemoveManualNodeCmd)

	addr := network.NormalizeAddress(c.Addr, s.cfg.DAGParams.DefaultPort)
	err := s.cfg.ConnMgr.RemoveByAddr(addr)

	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}

	// no data returned unless an error.
	return nil, nil
}

// handleNode handles node commands.
func handleNode(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.NodeCmd)

	var addr string
	var nodeID uint64
	var errN, err error
	params := s.cfg.DAGParams
	switch c.SubCmd {
	case "disconnect":
		// If we have a valid uint disconnect by node id. Otherwise,
		// attempt to disconnect by address, returning an error if a
		// valid IP address is not supplied.
		if nodeID, errN = strconv.ParseUint(c.Target, 10, 32); errN == nil {
			err = s.cfg.ConnMgr.DisconnectByID(int32(nodeID))
		} else {
			if _, _, errP := net.SplitHostPort(c.Target); errP == nil || net.ParseIP(c.Target) != nil {
				addr = network.NormalizeAddress(c.Target, params.DefaultPort)
				err = s.cfg.ConnMgr.DisconnectByAddr(addr)
			} else {
				return nil, &btcjson.RPCError{
					Code:    btcjson.ErrRPCInvalidParameter,
					Message: "invalid address or node ID",
				}
			}
		}
		if err != nil && peerExists(s.cfg.ConnMgr, addr, int32(nodeID)) {

			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCMisc,
				Message: "can't disconnect a permanent peer, use remove",
			}
		}

	case "remove":
		// If we have a valid uint disconnect by node id. Otherwise,
		// attempt to disconnect by address, returning an error if a
		// valid IP address is not supplied.
		if nodeID, errN = strconv.ParseUint(c.Target, 10, 32); errN == nil {
			err = s.cfg.ConnMgr.RemoveByID(int32(nodeID))
		} else {
			if _, _, errP := net.SplitHostPort(c.Target); errP == nil || net.ParseIP(c.Target) != nil {
				addr = network.NormalizeAddress(c.Target, params.DefaultPort)
				err = s.cfg.ConnMgr.RemoveByAddr(addr)
			} else {
				return nil, &btcjson.RPCError{
					Code:    btcjson.ErrRPCInvalidParameter,
					Message: "invalid address or node ID",
				}
			}
		}
		if err != nil && peerExists(s.cfg.ConnMgr, addr, int32(nodeID)) {
			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCMisc,
				Message: "can't remove a temporary peer, use disconnect",
			}
		}

	case "connect":
		addr = network.NormalizeAddress(c.Target, params.DefaultPort)

		// Default to temporary connections.
		subCmd := "temp"
		if c.ConnectSubCmd != nil {
			subCmd = *c.ConnectSubCmd
		}

		switch subCmd {
		case "perm", "temp":
			err = s.cfg.ConnMgr.Connect(addr, subCmd == "perm")
		default:
			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCInvalidParameter,
				Message: "invalid subcommand for node connect",
			}
		}
	default:
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidParameter,
			Message: "invalid subcommand for node",
		}
	}

	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}

	// no data returned unless an error.
	return nil, nil
}

// peerExists determines if a certain peer is currently connected given
// information about all currently connected peers. Peer existence is
// determined using either a target address or node id.
func peerExists(connMgr rpcserverConnManager, addr string, nodeID int32) bool {
	for _, p := range connMgr.ConnectedPeers() {
		if p.ToPeer().ID() == nodeID || p.ToPeer().Addr() == addr {
			return true
		}
	}
	return false
}

// messageToHex serializes a message to the wire protocol encoding using the
// latest protocol version and returns a hex-encoded string of the result.
func messageToHex(msg wire.Message) (string, error) {
	var buf bytes.Buffer
	if err := msg.BtcEncode(&buf, maxProtocolVersion); err != nil {
		context := fmt.Sprintf("Failed to encode msg of type %T", msg)
		return "", internalRPCError(err.Error(), context)
	}

	return hex.EncodeToString(buf.Bytes()), nil
}

// handleCreateRawTransaction handles createRawTransaction commands.
func handleCreateRawTransaction(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.CreateRawTransactionCmd)

	// Validate the locktime, if given.
	if c.LockTime != nil &&
		(*c.LockTime < 0 || *c.LockTime > wire.MaxTxInSequenceNum) {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidParameter,
			Message: "Locktime out of range",
		}
	}

	txIns := []*wire.TxIn{}
	// Add all transaction inputs to a new transaction after performing
	// some validity checks.
	for _, input := range c.Inputs {
		txID, err := daghash.NewTxIDFromStr(input.TxID)
		if err != nil {
			return nil, rpcDecodeHexError(input.TxID)
		}

		prevOut := wire.NewOutpoint(txID, input.Vout)
		txIn := wire.NewTxIn(prevOut, []byte{})
		if c.LockTime != nil && *c.LockTime != 0 {
			txIn.Sequence = wire.MaxTxInSequenceNum - 1
		}
		txIns = append(txIns, txIn)
	}
	mtx := wire.NewNativeMsgTx(wire.TxVersion, txIns, nil)

	// Add all transaction outputs to the transaction after performing
	// some validity checks.
	params := s.cfg.DAGParams
	for encodedAddr, amount := range c.Amounts {
		// Ensure amount is in the valid range for monetary amounts.
		if amount <= 0 || amount > util.MaxSatoshi {
			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCType,
				Message: "Invalid amount",
			}
		}

		// Decode the provided address.
		addr, err := util.DecodeAddress(encodedAddr, params.Prefix)
		if err != nil {
			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address or key: " + err.Error(),
			}
		}

		// Ensure the address is one of the supported types and that
		// the network encoded with the address matches the network the
		// server is currently on.
		switch addr.(type) {
		case *util.AddressPubKeyHash:
		case *util.AddressScriptHash:
		default:
			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address or key",
			}
		}
		if !addr.IsForPrefix(params.Prefix) {
			return nil, &btcjson.RPCError{
				Code: btcjson.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address: " + encodedAddr +
					" is for the wrong network",
			}
		}

		// Create a new script which pays to the provided address.
		scriptPubKey, err := txscript.PayToAddrScript(addr)
		if err != nil {
			context := "Failed to generate pay-to-address script"
			return nil, internalRPCError(err.Error(), context)
		}

		// Convert the amount to satoshi.
		satoshi, err := util.NewAmount(amount)
		if err != nil {
			context := "Failed to convert amount"
			return nil, internalRPCError(err.Error(), context)
		}

		txOut := wire.NewTxOut(uint64(satoshi), scriptPubKey)
		mtx.AddTxOut(txOut)
	}

	// Set the Locktime, if given.
	if c.LockTime != nil {
		mtx.LockTime = *c.LockTime
	}

	// Return the serialized and hex-encoded transaction.  Note that this
	// is intentionally not directly returning because the first return
	// value is a string and it would result in returning an empty string to
	// the client instead of nothing (nil) in the case of an error.
	mtxHex, err := messageToHex(mtx)
	if err != nil {
		return nil, err
	}
	return mtxHex, nil
}

// handleDebugLevel handles debugLevel commands.
func handleDebugLevel(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.DebugLevelCmd)

	// Special show command to list supported subsystems.
	if c.LevelSpec == "show" {
		return fmt.Sprintf("Supported subsystems %s",
			logger.SupportedSubsystems()), nil
	}

	err := logger.ParseAndSetDebugLevels(c.LevelSpec)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidParams.Code,
			Message: err.Error(),
		}
	}

	return "Done.", nil
}

// createVinList returns a slice of JSON objects for the inputs of the passed
// transaction.
func createVinList(mtx *wire.MsgTx) []btcjson.Vin {
	// Coinbase transactions only have a single txin by definition.
	vinList := make([]btcjson.Vin, len(mtx.TxIn))
	if mtx.IsCoinBase() {
		txIn := mtx.TxIn[0]
		vinList[0].Coinbase = hex.EncodeToString(txIn.SignatureScript)
		vinList[0].Sequence = txIn.Sequence
		return vinList
	}

	for i, txIn := range mtx.TxIn {
		// The disassembled string will contain [error] inline
		// if the script doesn't fully parse, so ignore the
		// error here.
		disbuf, _ := txscript.DisasmString(txIn.SignatureScript)

		vinEntry := &vinList[i]
		vinEntry.TxID = txIn.PreviousOutpoint.TxID.String()
		vinEntry.Vout = txIn.PreviousOutpoint.Index
		vinEntry.Sequence = txIn.Sequence
		vinEntry.ScriptSig = &btcjson.ScriptSig{
			Asm: disbuf,
			Hex: hex.EncodeToString(txIn.SignatureScript),
		}
	}

	return vinList
}

// createVoutList returns a slice of JSON objects for the outputs of the passed
// transaction.
func createVoutList(mtx *wire.MsgTx, chainParams *dagconfig.Params, filterAddrMap map[string]struct{}) []btcjson.Vout {
	voutList := make([]btcjson.Vout, 0, len(mtx.TxOut))
	for i, v := range mtx.TxOut {
		// The disassembled string will contain [error] inline if the
		// script doesn't fully parse, so ignore the error here.
		disbuf, _ := txscript.DisasmString(v.ScriptPubKey)

		// Ignore the error here since an error means the script
		// couldn't parse and there is no additional information about
		// it anyways.
		scriptClass, addr, _ := txscript.ExtractScriptPubKeyAddress(
			v.ScriptPubKey, chainParams)

		// Encode the addresses while checking if the address passes the
		// filter when needed.
		passesFilter := len(filterAddrMap) == 0
		var encodedAddr *string
		if addr != nil {
			encodedAddr = btcjson.String(addr.EncodeAddress())

			// If the filter doesn't already pass, make it pass if
			// the address exists in the filter.
			if _, exists := filterAddrMap[*encodedAddr]; exists {
				passesFilter = true
			}
		}

		if !passesFilter {
			continue
		}

		var vout btcjson.Vout
		vout.N = uint32(i)
		vout.Value = util.Amount(v.Value).ToBTC()
		vout.ScriptPubKey.Address = encodedAddr
		vout.ScriptPubKey.Asm = disbuf
		vout.ScriptPubKey.Hex = hex.EncodeToString(v.ScriptPubKey)
		vout.ScriptPubKey.Type = scriptClass.String()

		voutList = append(voutList, vout)
	}

	return voutList
}

// createTxRawResult converts the passed transaction and associated parameters
// to a raw transaction JSON object.
func createTxRawResult(dagParams *dagconfig.Params, mtx *wire.MsgTx,
	txID string, blkHeader *wire.BlockHeader, blkHash string,
	acceptingBlock *daghash.Hash, confirmations *uint64, isInMempool bool) (*btcjson.TxRawResult, error) {

	mtxHex, err := messageToHex(mtx)
	if err != nil {
		return nil, err
	}

	var payloadHash string
	if mtx.PayloadHash != nil {
		payloadHash = mtx.PayloadHash.String()
	}

	txReply := &btcjson.TxRawResult{
		Hex:         mtxHex,
		TxID:        txID,
		Hash:        mtx.TxHash().String(),
		Size:        int32(mtx.SerializeSize()),
		Vin:         createVinList(mtx),
		Vout:        createVoutList(mtx, dagParams, nil),
		Version:     mtx.Version,
		LockTime:    mtx.LockTime,
		Subnetwork:  mtx.SubnetworkID.String(),
		Gas:         mtx.Gas,
		PayloadHash: payloadHash,
		Payload:     hex.EncodeToString(mtx.Payload),
	}

	if blkHeader != nil {
		// This is not a typo, they are identical in bitcoind as well.
		txReply.Time = uint64(blkHeader.Timestamp.Unix())
		txReply.BlockTime = uint64(blkHeader.Timestamp.Unix())
		txReply.BlockHash = blkHash
	}

	txReply.Confirmations = confirmations
	txReply.IsInMempool = isInMempool
	if acceptingBlock != nil {
		txReply.AcceptedBy = btcjson.String(acceptingBlock.String())
	}

	return txReply, nil
}

// handleDecodeRawTransaction handles decodeRawTransaction commands.
func handleDecodeRawTransaction(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.DecodeRawTransactionCmd)

	// Deserialize the transaction.
	hexStr := c.HexTx
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	serializedTx, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpcDecodeHexError(hexStr)
	}
	var mtx wire.MsgTx
	err = mtx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCDeserialization,
			Message: "TX decode failed: " + err.Error(),
		}
	}

	// Create and return the result.
	txReply := btcjson.TxRawDecodeResult{
		TxID:     mtx.TxID().String(),
		Version:  mtx.Version,
		Locktime: mtx.LockTime,
		Vin:      createVinList(&mtx),
		Vout:     createVoutList(&mtx, s.cfg.DAGParams, nil),
	}
	return txReply, nil
}

// handleDecodeScript handles decodeScript commands.
func handleDecodeScript(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.DecodeScriptCmd)

	// Convert the hex script to bytes.
	hexStr := c.HexScript
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	script, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpcDecodeHexError(hexStr)
	}

	// The disassembled string will contain [error] inline if the script
	// doesn't fully parse, so ignore the error here.
	disbuf, _ := txscript.DisasmString(script)

	// Get information about the script.
	// Ignore the error here since an error means the script couldn't parse
	// and there is no additinal information about it anyways.
	scriptClass, addr, _ := txscript.ExtractScriptPubKeyAddress(script,
		s.cfg.DAGParams)
	var address *string
	if addr != nil {
		address = btcjson.String(addr.EncodeAddress())
	}

	// Convert the script itself to a pay-to-script-hash address.
	p2sh, err := util.NewAddressScriptHash(script, s.cfg.DAGParams.Prefix)
	if err != nil {
		context := "Failed to convert script to pay-to-script-hash"
		return nil, internalRPCError(err.Error(), context)
	}

	// Generate and return the reply.
	reply := btcjson.DecodeScriptResult{
		Asm:     disbuf,
		Type:    scriptClass.String(),
		Address: address,
	}
	if scriptClass != txscript.ScriptHashTy {
		reply.P2sh = p2sh.EncodeAddress()
	}
	return reply, nil
}

// handleGenerate handles generate commands.
func handleGenerate(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Respond with an error if there are no addresses to pay the
	// created blocks to.
	if len(config.MainConfig().MiningAddrs) == 0 {
		return nil, &btcjson.RPCError{
			Code: btcjson.ErrRPCInternal.Code,
			Message: "No payment addresses specified " +
				"via --miningaddr",
		}
	}

	if config.MainConfig().SubnetworkID != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidRequest.Code,
			Message: "`generate` is not supported on partial nodes.",
		}
	}

	// Respond with an error if there's virtually 0 chance of mining a block
	// with the CPU.
	if !s.cfg.DAGParams.GenerateSupported {
		return nil, &btcjson.RPCError{
			Code: btcjson.ErrRPCDifficulty,
			Message: fmt.Sprintf("No support for `generate` on "+
				"the current network, %s, as it's unlikely to "+
				"be possible to mine a block with the CPU.",
				s.cfg.DAGParams.Net),
		}
	}

	c := cmd.(*btcjson.GenerateCmd)

	// Respond with an error if the client is requesting 0 blocks to be generated.
	if c.NumBlocks == 0 {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInternal.Code,
			Message: "Please request a nonzero number of blocks to generate.",
		}
	}

	// Create a reply
	reply := make([]string, c.NumBlocks)

	blockHashes, err := s.cfg.CPUMiner.GenerateNBlocks(c.NumBlocks)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInternal.Code,
			Message: err.Error(),
		}
	}

	// Mine the correct number of blocks, assigning the hex representation of the
	// hash of each one to its place in the reply.
	for i, hash := range blockHashes {
		reply[i] = hash.String()
	}

	return reply, nil
}

// getManualNodesInfo handles getManualNodeInfo and getAllManualNodesInfo commands.
func getManualNodesInfo(s *Server, detailsArg *bool, node string) (interface{}, error) {

	details := detailsArg == nil || *detailsArg

	// Retrieve a list of persistent (manual) peers from the server and
	// filter the list of peers per the specified address (if any).
	peers := s.cfg.ConnMgr.PersistentPeers()
	if node != "" {
		found := false
		for i, peer := range peers {
			if peer.ToPeer().Addr() == node {
				peers = peers[i : i+1]
				found = true
			}
		}
		if !found {
			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCClientNodeNotAdded,
				Message: "Node has not been added",
			}
		}
	}

	// Without the details flag, the result is just a slice of the addresses as
	// strings.
	if !details {
		results := make([]string, 0, len(peers))
		for _, peer := range peers {
			results = append(results, peer.ToPeer().Addr())
		}
		return results, nil
	}

	// With the details flag, the result is an array of JSON objects which
	// include the result of DNS lookups for each peer.
	results := make([]*btcjson.GetManualNodeInfoResult, 0, len(peers))
	for _, rpcPeer := range peers {
		// Set the "address" of the peer which could be an ip address
		// or a domain name.
		peer := rpcPeer.ToPeer()
		var result btcjson.GetManualNodeInfoResult
		result.ManualNode = peer.Addr()
		result.Connected = btcjson.Bool(peer.Connected())

		// Split the address into host and port portions so we can do
		// a DNS lookup against the host.  When no port is specified in
		// the address, just use the address as the host.
		host, _, err := net.SplitHostPort(peer.Addr())
		if err != nil {
			host = peer.Addr()
		}

		var ipList []string
		switch {
		case net.ParseIP(host) != nil, strings.HasSuffix(host, ".onion"):
			ipList = make([]string, 1)
			ipList[0] = host
		default:
			// Do a DNS lookup for the address.  If the lookup fails, just
			// use the host.
			ips, err := serverutils.BTCDLookup(host)
			if err != nil {
				ipList = make([]string, 1)
				ipList[0] = host
				break
			}
			ipList = make([]string, 0, len(ips))
			for _, ip := range ips {
				ipList = append(ipList, ip.String())
			}
		}

		// Add the addresses and connection info to the result.
		addrs := make([]btcjson.GetManualNodeInfoResultAddr, 0, len(ipList))
		for _, ip := range ipList {
			var addr btcjson.GetManualNodeInfoResultAddr
			addr.Address = ip
			addr.Connected = "false"
			if ip == host && peer.Connected() {
				addr.Connected = logger.DirectionString(peer.Inbound())
			}
			addrs = append(addrs, addr)
		}
		result.Addresses = &addrs
		results = append(results, &result)
	}
	return results, nil
}

// handleGetManualNodeInfo handles getManualNodeInfo commands.
func handleGetManualNodeInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetManualNodeInfoCmd)
	results, err := getManualNodesInfo(s, c.Details, c.Node)
	if err != nil {
		return nil, err
	}
	if resultsNonDetailed, ok := results.([]string); ok {
		return resultsNonDetailed[0], nil
	}
	resultsDetailed := results.([]*btcjson.GetManualNodeInfoResult)
	return resultsDetailed[0], nil
}

// handleGetAllManualNodesInfo handles getAllManualNodesInfo commands.
func handleGetAllManualNodesInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetAllManualNodesInfoCmd)
	return getManualNodesInfo(s, c.Details, "")
}

// handleGetBestBlock implements the getBestBlock command.
func handleGetBestBlock(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// All other "get block" commands give either the height, the
	// hash, or both but require the block SHA.  This gets both for
	// the best block.
	result := &btcjson.GetBestBlockResult{
		Hash:   s.cfg.DAG.SelectedTipHash().String(),
		Height: s.cfg.DAG.ChainHeight(), //TODO: (Ori) This is probably wrong. Done only for compilation
	}
	return result, nil
}

// handleGetBestBlockHash implements the getBestBlockHash command.
func handleGetBestBlockHash(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.cfg.DAG.SelectedTipHash().String(), nil
}

// getDifficultyRatio returns the proof-of-work difficulty as a multiple of the
// minimum difficulty using the passed bits field from the header of a block.
func getDifficultyRatio(bits uint32, params *dagconfig.Params) float64 {
	// The minimum difficulty is the max possible proof-of-work limit bits
	// converted back to a number.  Note this is not the same as the proof of
	// work limit directly because the block difficulty is encoded in a block
	// with the compact form which loses precision.
	target := util.CompactToBig(bits)

	difficulty := new(big.Rat).SetFrac(params.PowMax, target)
	outString := difficulty.FloatString(8)
	diff, err := strconv.ParseFloat(outString, 64)
	if err != nil {
		log.Errorf("Cannot get difficulty: %s", err)
		return 0
	}
	return diff
}

// handleGetBlock implements the getBlock command.
func handleGetBlock(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetBlockCmd)

	// Load the raw block bytes from the database.
	hash, err := daghash.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, rpcDecodeHexError(c.Hash)
	}
	var blkBytes []byte
	err = s.cfg.DB.View(func(dbTx database.Tx) error {
		var err error
		blkBytes, err = dbTx.FetchBlock(hash)
		return err
	})
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

	// Handle partial blocks
	if c.Subnetwork != nil {
		requestSubnetworkID, err := subnetworkid.NewFromStr(*c.Subnetwork)
		if err != nil {
			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCInvalidRequest.Code,
				Message: "invalid subnetwork string",
			}
		}
		nodeSubnetworkID := config.MainConfig().SubnetworkID

		if requestSubnetworkID != nil {
			if nodeSubnetworkID != nil {
				if !nodeSubnetworkID.IsEqual(requestSubnetworkID) {
					return nil, &btcjson.RPCError{
						Code:    btcjson.ErrRPCInvalidRequest.Code,
						Message: "subnetwork does not match this partial node",
					}
				}
				// nothing to do - partial node stores partial blocks
			} else {
				// Deserialize the block.
				var msgBlock wire.MsgBlock
				err = msgBlock.Deserialize(bytes.NewReader(blkBytes))
				if err != nil {
					context := "Failed to deserialize block"
					return nil, internalRPCError(err.Error(), context)
				}
				msgBlock.ConvertToPartial(requestSubnetworkID)
				var b bytes.Buffer
				msgBlock.Serialize(bufio.NewWriter(&b))
				blkBytes = b.Bytes()
			}
		}
	}

	// When the verbose flag isn't set, simply return the serialized block
	// as a hex-encoded string.
	if c.Verbose != nil && !*c.Verbose {
		return hex.EncodeToString(blkBytes), nil
	}

	// The verbose flag is set, so generate the JSON object and return it.

	// Deserialize the block.
	blk, err := util.NewBlockFromBytes(blkBytes)
	if err != nil {
		context := "Failed to deserialize block"
		return nil, internalRPCError(err.Error(), context)
	}

	blockReply, err := buildGetBlockVerboseResult(s, blk, c.VerboseTx == nil || !*c.VerboseTx)
	if err != nil {
		return nil, err
	}
	return blockReply, nil
}

func buildGetBlockVerboseResult(s *Server, block *util.Block, isVerboseTx bool) (*btcjson.GetBlockVerboseResult, error) {
	hash := block.Hash()
	params := s.cfg.DAGParams
	blockHeader := block.MsgBlock().Header

	// Get the block chain height.
	blockChainHeight, err := s.cfg.DAG.BlockChainHeightByHash(hash)
	if err != nil {
		context := "Failed to obtain block height"
		return nil, internalRPCError(err.Error(), context)
	}

	// Get the hashes for the next blocks unless there are none.
	var nextHashStrings []string
	if blockChainHeight < s.cfg.DAG.ChainHeight() { //TODO: (Ori) This is probably wrong. Done only for compilation
		childHashes, err := s.cfg.DAG.ChildHashesByHash(hash)
		if err != nil {
			context := "No next block"
			return nil, internalRPCError(err.Error(), context)
		}
		nextHashStrings = daghash.Strings(childHashes)
	}

	blockConfirmations, err := s.cfg.DAG.BlockConfirmationsByHash(hash)
	if err != nil {
		context := "Could not get block confirmations"
		return nil, internalRPCError(err.Error(), context)
	}

	result := &btcjson.GetBlockVerboseResult{
		Hash:                 hash.String(),
		Version:              blockHeader.Version,
		VersionHex:           fmt.Sprintf("%08x", blockHeader.Version),
		HashMerkleRoot:       blockHeader.HashMerkleRoot.String(),
		AcceptedIDMerkleRoot: blockHeader.AcceptedIDMerkleRoot.String(),
		ParentHashes:         daghash.Strings(blockHeader.ParentHashes),
		Nonce:                blockHeader.Nonce,
		Time:                 blockHeader.Timestamp.Unix(),
		Confirmations:        blockConfirmations,
		Height:               blockChainHeight,
		Size:                 int32(block.MsgBlock().SerializeSize()),
		Bits:                 strconv.FormatInt(int64(blockHeader.Bits), 16),
		Difficulty:           getDifficultyRatio(blockHeader.Bits, params),
		NextHashes:           nextHashStrings,
	}

	if isVerboseTx {
		transactions := block.Transactions()
		txNames := make([]string, len(transactions))
		for i, tx := range transactions {
			txNames[i] = tx.ID().String()
		}

		result.Tx = txNames
	} else {
		txns := block.Transactions()
		rawTxns := make([]btcjson.TxRawResult, len(txns))
		for i, tx := range txns {
			var acceptingBlock *daghash.Hash
			var confirmations *uint64
			if s.cfg.TxIndex != nil {
				acceptingBlock, err = s.cfg.TxIndex.BlockThatAcceptedTx(s.cfg.DAG, tx.ID())
				if err != nil {
					return nil, err
				}
				txConfirmations, err := txConfirmations(s, tx.ID())
				if err != nil {
					return nil, err
				}
				confirmations = &txConfirmations
			}
			rawTxn, err := createTxRawResult(params, tx.MsgTx(), tx.ID().String(),
				&blockHeader, hash.String(), acceptingBlock, confirmations, false)
			if err != nil {
				return nil, err
			}
			rawTxns[i] = *rawTxn
		}
		result.RawTx = rawTxns
	}

	return result, nil
}

// softForkStatus converts a ThresholdState state into a human readable string
// corresponding to the particular state.
func softForkStatus(state blockdag.ThresholdState) (string, error) {
	switch state {
	case blockdag.ThresholdDefined:
		return "defined", nil
	case blockdag.ThresholdStarted:
		return "started", nil
	case blockdag.ThresholdLockedIn:
		return "lockedin", nil
	case blockdag.ThresholdActive:
		return "active", nil
	case blockdag.ThresholdFailed:
		return "failed", nil
	default:
		return "", fmt.Errorf("unknown deployment state: %s", state)
	}
}

// handleGetBlockDAGInfo implements the getBlockDagInfo command.
func handleGetBlockDAGInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Obtain a snapshot of the current best known DAG state. We'll
	// populate the response to this call primarily from this snapshot.
	params := s.cfg.DAGParams
	dag := s.cfg.DAG

	dagInfo := &btcjson.GetBlockDAGInfoResult{
		DAG:            params.Name,
		Blocks:         dag.BlockCount(),
		Headers:        dag.BlockCount(),
		TipHashes:      daghash.Strings(dag.TipHashes()),
		Difficulty:     getDifficultyRatio(dag.CurrentBits(), params),
		MedianTime:     dag.CalcPastMedianTime().Unix(),
		UTXOCommitment: dag.UTXOCommitment(),
		Pruned:         false,
		Bip9SoftForks:  make(map[string]*btcjson.Bip9SoftForkDescription),
	}

	// Finally, query the BIP0009 version bits state for all currently
	// defined BIP0009 soft-fork deployments.
	for deployment, deploymentDetails := range params.Deployments {
		// Map the integer deployment ID into a human readable
		// fork-name.
		var forkName string
		switch deployment {
		case dagconfig.DeploymentTestDummy:
			forkName = "dummy"

		default:
			return nil, &btcjson.RPCError{
				Code: btcjson.ErrRPCInternal.Code,
				Message: fmt.Sprintf("Unknown deployment %d "+
					"detected", deployment),
			}
		}

		// Query the dag for the current status of the deployment as
		// identified by its deployment ID.
		deploymentStatus, err := dag.ThresholdState(uint32(deployment))
		if err != nil {
			context := "Failed to obtain deployment status"
			return nil, internalRPCError(err.Error(), context)
		}

		// Attempt to convert the current deployment status into a
		// human readable string. If the status is unrecognized, then a
		// non-nil error is returned.
		statusString, err := softForkStatus(deploymentStatus)
		if err != nil {
			return nil, &btcjson.RPCError{
				Code: btcjson.ErrRPCInternal.Code,
				Message: fmt.Sprintf("unknown deployment status: %d",
					deploymentStatus),
			}
		}

		// Finally, populate the soft-fork description with all the
		// information gathered above.
		dagInfo.Bip9SoftForks[forkName] = &btcjson.Bip9SoftForkDescription{
			Status:    strings.ToLower(statusString),
			Bit:       deploymentDetails.BitNumber,
			StartTime: int64(deploymentDetails.StartTime),
			Timeout:   int64(deploymentDetails.ExpireTime),
		}
	}

	return dagInfo, nil
}

// handleGetBlockCount implements the getBlockCount command.
func handleGetBlockCount(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.cfg.DAG.BlockCount(), nil
}

// handleGetBlockHash implements the getBlockHash command.
// This command had been (possibly temporarily) dropped.
// Originally it relied on height, which no longer makes sense.
func handleGetBlockHash(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return nil, ErrRPCUnimplemented
}

// handleGetBlockHeader implements the getBlockHeader command.
func handleGetBlockHeader(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetBlockHeaderCmd)

	// Fetch the header from chain.
	hash, err := daghash.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, rpcDecodeHexError(c.Hash)
	}
	blockHeader, err := s.cfg.DAG.HeaderByHash(hash)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

	// When the verbose flag isn't set, simply return the serialized block
	// header as a hex-encoded string.
	if c.Verbose != nil && !*c.Verbose {
		var headerBuf bytes.Buffer
		err := blockHeader.Serialize(&headerBuf)
		if err != nil {
			context := "Failed to serialize block header"
			return nil, internalRPCError(err.Error(), context)
		}
		return hex.EncodeToString(headerBuf.Bytes()), nil
	}

	// The verbose flag is set, so generate the JSON object and return it.

	// Get the block chain height from chain.
	blockChainHeight, err := s.cfg.DAG.BlockChainHeightByHash(hash)
	if err != nil {
		context := "Failed to obtain block height"
		return nil, internalRPCError(err.Error(), context)
	}

	// Get the hashes for the next blocks unless there are none.
	var nextHashStrings []string
	if blockChainHeight < s.cfg.DAG.ChainHeight() { //TODO: (Ori) This is probably wrong. Done only for compilation
		childHashes, err := s.cfg.DAG.ChildHashesByHash(hash)
		if err != nil {
			context := "No next block"
			return nil, internalRPCError(err.Error(), context)
		}
		nextHashStrings = daghash.Strings(childHashes)
	}

	blockConfirmations, err := s.cfg.DAG.BlockConfirmationsByHash(hash)
	if err != nil {
		context := "Could not get block confirmations"
		return nil, internalRPCError(err.Error(), context)
	}

	params := s.cfg.DAGParams
	blockHeaderReply := btcjson.GetBlockHeaderVerboseResult{
		Hash:                 c.Hash,
		Confirmations:        blockConfirmations,
		Height:               blockChainHeight,
		Version:              blockHeader.Version,
		VersionHex:           fmt.Sprintf("%08x", blockHeader.Version),
		HashMerkleRoot:       blockHeader.HashMerkleRoot.String(),
		AcceptedIDMerkleRoot: blockHeader.AcceptedIDMerkleRoot.String(),
		NextHashes:           nextHashStrings,
		ParentHashes:         daghash.Strings(blockHeader.ParentHashes),
		Nonce:                uint64(blockHeader.Nonce),
		Time:                 blockHeader.Timestamp.Unix(),
		Bits:                 strconv.FormatInt(int64(blockHeader.Bits), 16),
		Difficulty:           getDifficultyRatio(blockHeader.Bits, params),
	}
	return blockHeaderReply, nil
}

// encodeLongPollID encodes the passed details into an ID that can be used to
// uniquely identify a block template.
func encodeLongPollID(parentHashes []*daghash.Hash, lastGenerated time.Time) string {
	return fmt.Sprintf("%s-%d", daghash.JoinHashesStrings(parentHashes, ""), lastGenerated.Unix())
}

// decodeLongPollID decodes an ID that is used to uniquely identify a block
// template.  This is mainly used as a mechanism to track when to update clients
// that are using long polling for block templates.  The ID consists of the
// parent blocks hashes for the associated template and the time the associated
// template was generated.
func decodeLongPollID(longPollID string) ([]*daghash.Hash, int64, error) {
	fields := strings.Split(longPollID, "-")
	if len(fields) != 2 {
		return nil, 0, errors.New("decodeLongPollID: invalid number of fields")
	}

	parentHashesStr := fields[0]
	if len(parentHashesStr)%daghash.HashSize != 0 {
		return nil, 0, errors.New("decodeLongPollID: invalid parent hashes format")
	}
	numberOfHashes := len(parentHashesStr) / daghash.HashSize

	parentHashes := make([]*daghash.Hash, 0, numberOfHashes)

	for i := 0; i < len(parentHashesStr); i += daghash.HashSize {
		hash, err := daghash.NewHashFromStr(parentHashesStr[i : i+daghash.HashSize])
		if err != nil {
			return nil, 0, fmt.Errorf("decodeLongPollID: NewHashFromStr: %s", err)
		}
		parentHashes = append(parentHashes, hash)
	}

	lastGenerated, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return nil, 0, fmt.Errorf("decodeLongPollID: Cannot parse timestamp %s: %s", fields[1], err)
	}

	return parentHashes, lastGenerated, nil
}

// handleGetCFilter implements the getCFilter command.
func handleGetCFilter(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if s.cfg.CfIndex == nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCNoCFIndex,
			Message: "The CF index must be enabled for this command",
		}
	}

	c := cmd.(*btcjson.GetCFilterCmd)
	hash, err := daghash.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, rpcDecodeHexError(c.Hash)
	}

	filterBytes, err := s.cfg.CfIndex.FilterByBlockHash(hash, c.FilterType)
	if err != nil {
		log.Debugf("Could not find committed filter for %s: %s",
			hash, err)
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

	log.Debugf("Found committed filter for %s", hash)
	return hex.EncodeToString(filterBytes), nil
}

// handleGetCFilterHeader implements the getCFilterHeader command.
func handleGetCFilterHeader(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if s.cfg.CfIndex == nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCNoCFIndex,
			Message: "The CF index must be enabled for this command",
		}
	}

	c := cmd.(*btcjson.GetCFilterHeaderCmd)
	hash, err := daghash.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, rpcDecodeHexError(c.Hash)
	}

	headerBytes, err := s.cfg.CfIndex.FilterHeaderByBlockHash(hash, c.FilterType)
	if len(headerBytes) > 0 {
		log.Debugf("Found header of committed filter for %s", hash)
	} else {
		log.Debugf("Could not find header of committed filter for %s: %s",
			hash, err)
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

	hash.SetBytes(headerBytes)
	return hash.String(), nil
}

// handleGetChainFromBlock implements the getChainFromBlock command.
func handleGetChainFromBlock(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if s.cfg.AcceptanceIndex == nil {
		return nil, &btcjson.RPCError{
			Code: btcjson.ErrRPCNoAcceptanceIndex,
			Message: "The acceptance index must be " +
				"enabled to get the selected parent chain " +
				"(specify --acceptanceindex)",
		}
	}

	c := cmd.(*btcjson.GetChainFromBlockCmd)
	var startHash *daghash.Hash
	if c.StartHash != nil {
		err := daghash.Decode(startHash, *c.StartHash)
		if err != nil {
			return nil, rpcDecodeHexError(*c.StartHash)
		}
	}

	s.cfg.DAG.RLock()
	defer s.cfg.DAG.RUnlock()

	// If startHash is not in the selected parent chain, there's nothing
	// to do; return an error.
	if !s.cfg.DAG.IsInSelectedParentChain(startHash) {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCBlockNotFound,
			Message: "Block not found in selected parent chain",
		}
	}

	// Retrieve the selected parent chain.
	selectedParentChain, err := s.cfg.DAG.SelectedParentChain(startHash)
	if err != nil {
		return nil, err
	}

	// Collect chainBlocks.
	chainBlocks, err := collectChainBlocks(s, selectedParentChain)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInternal.Code,
			Message: fmt.Sprintf("could not collect chain blocks: %s", err),
		}
	}

	result := &btcjson.GetChainFromBlockResult{
		SelectedParentChain: chainBlocks,
		Blocks:              nil,
	}

	// If the user specified to include the blocks, collect them as well.
	if c.IncludeBlocks != nil && *c.IncludeBlocks {
		getBlockVerboseResults, err := hashesToGetBlockVerboseResults(s, selectedParentChain)
		if err != nil {
			return nil, err
		}
		result.Blocks = getBlockVerboseResults
	}

	return result, nil
}

func collectChainBlocks(s *Server, selectedParentChain []*daghash.Hash) ([]btcjson.ChainBlock, error) {
	chainBlocks := make([]btcjson.ChainBlock, 0, len(selectedParentChain))
	for _, hash := range selectedParentChain {
		acceptanceData, err := s.cfg.AcceptanceIndex.TxsAcceptanceData(hash)
		if err != nil {
			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCInternal.Code,
				Message: fmt.Sprintf("could not retrieve acceptance data for block %s", hash),
			}
		}

		acceptedBlocks := make([]btcjson.AcceptedBlock, 0, len(acceptanceData))
		for blockHash, blockAcceptanceData := range acceptanceData {
			acceptedTxIds := make([]string, 0, len(blockAcceptanceData))
			for _, txAcceptanceData := range blockAcceptanceData {
				if txAcceptanceData.IsAccepted {
					acceptedTxIds = append(acceptedTxIds, txAcceptanceData.Tx.Hash().String())
				}
			}
			acceptedBlock := btcjson.AcceptedBlock{
				Hash:          blockHash.String(),
				AcceptedTxIDs: acceptedTxIds,
			}
			acceptedBlocks = append(acceptedBlocks, acceptedBlock)
		}

		chainBlock := btcjson.ChainBlock{
			Hash:           hash.String(),
			AcceptedBlocks: acceptedBlocks,
		}
		chainBlocks = append(chainBlocks, chainBlock)
	}
	return chainBlocks, nil
}

func hashesToGetBlockVerboseResults(s *Server, hashes []*daghash.Hash) ([]btcjson.GetBlockVerboseResult, error) {
	getBlockVerboseResults := make([]btcjson.GetBlockVerboseResult, 0, len(hashes))
	for _, blockHash := range hashes {
		block, err := s.cfg.DAG.BlockByHash(blockHash)
		if err != nil {
			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCInternal.Code,
				Message: fmt.Sprintf("could not retrieve block %s.", blockHash),
			}
		}
		getBlockVerboseResult, err := buildGetBlockVerboseResult(s, block, false)
		if err != nil {
			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCInternal.Code,
				Message: fmt.Sprintf("could not build getBlockVerboseResult for block %s.", blockHash),
			}
		}
		getBlockVerboseResults = append(getBlockVerboseResults, *getBlockVerboseResult)
	}
	return getBlockVerboseResults, nil
}

// handleGetConnectionCount implements the getConnectionCount command.
func handleGetConnectionCount(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.cfg.ConnMgr.ConnectedCount(), nil
}

// handleGetCurrentNet implements the getCurrentNet command.
func handleGetCurrentNet(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.cfg.DAGParams.Net, nil
}

// handleGetDifficulty implements the getDifficulty command.
func handleGetDifficulty(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return getDifficultyRatio(s.cfg.DAG.SelectedTipHeader().Bits, s.cfg.DAGParams), nil
}

// handleGetGenerate implements the getGenerate command.
func handleGetGenerate(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.cfg.CPUMiner.IsMining(), nil
}

// handleGetHashesPerSec implements the getHashesPerSec command.
func handleGetHashesPerSec(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if config.MainConfig().SubnetworkID != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidRequest.Code,
			Message: "`getHashesPerSec` is not supported on partial nodes.",
		}
	}

	return int64(s.cfg.CPUMiner.HashesPerSecond()), nil
}

// handleGetTopHeaders implements the getTopHeaders command.
func handleGetTopHeaders(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetTopHeadersCmd)

	var startHash *daghash.Hash
	if c.StartHash != nil {
		startHash = &daghash.Hash{}
		err := daghash.Decode(startHash, *c.StartHash)
		if err != nil {
			return nil, rpcDecodeHexError(*c.StartHash)
		}
	}
	headers, err := s.cfg.DAG.GetTopHeaders(startHash)
	if err != nil {
		return nil, internalRPCError(err.Error(),
			"Failed to get top headers")
	}

	// Return the serialized block headers as hex-encoded strings.
	hexBlockHeaders := make([]string, len(headers))
	var buf bytes.Buffer
	for i, h := range headers {
		err := h.Serialize(&buf)
		if err != nil {
			return nil, internalRPCError(err.Error(),
				"Failed to serialize block header")
		}
		hexBlockHeaders[i] = hex.EncodeToString(buf.Bytes())
		buf.Reset()
	}
	return hexBlockHeaders, nil
}

// handleGetHeaders implements the getHeaders command.
//
// NOTE: This is a btcsuite extension originally ported from
// github.com/decred/dcrd.
func handleGetHeaders(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetHeadersCmd)

	startHash := &daghash.ZeroHash
	if c.StartHash != "" {
		err := daghash.Decode(startHash, c.StartHash)
		if err != nil {
			return nil, rpcDecodeHexError(c.StopHash)
		}
	}
	stopHash := &daghash.ZeroHash
	if c.StopHash != "" {
		err := daghash.Decode(stopHash, c.StopHash)
		if err != nil {
			return nil, rpcDecodeHexError(c.StopHash)
		}
	}
	headers := s.cfg.SyncMgr.GetBlueBlocksHeadersBetween(startHash, stopHash)

	// Return the serialized block headers as hex-encoded strings.
	hexBlockHeaders := make([]string, len(headers))
	var buf bytes.Buffer
	for i, h := range headers {
		err := h.Serialize(&buf)
		if err != nil {
			return nil, internalRPCError(err.Error(),
				"Failed to serialize block header")
		}
		hexBlockHeaders[i] = hex.EncodeToString(buf.Bytes())
		buf.Reset()
	}
	return hexBlockHeaders, nil
}

// handleGetInfo implements the getInfo command. We only return the fields
// that are not related to wallet functionality.
func handleGetInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	ret := &btcjson.InfoDAGResult{
		Version:         int32(1000000*version.AppMajor + 10000*version.AppMinor + 100*version.AppPatch),
		ProtocolVersion: int32(maxProtocolVersion),
		Blocks:          s.cfg.DAG.BlockCount(),
		TimeOffset:      int64(s.cfg.TimeSource.Offset().Seconds()),
		Connections:     s.cfg.ConnMgr.ConnectedCount(),
		Proxy:           config.MainConfig().Proxy,
		Difficulty:      getDifficultyRatio(s.cfg.DAG.CurrentBits(), s.cfg.DAGParams),
		TestNet:         config.MainConfig().TestNet3,
		DevNet:          config.MainConfig().DevNet,
		RelayFee:        config.MainConfig().MinRelayTxFee.ToBTC(),
	}

	return ret, nil
}

// handleGetMempoolInfo implements the getMempoolInfo command.
func handleGetMempoolInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	mempoolTxns := s.cfg.TxMemPool.TxDescs()

	var numBytes int64
	for _, txD := range mempoolTxns {
		numBytes += int64(txD.Tx.MsgTx().SerializeSize())
	}

	ret := &btcjson.GetMempoolInfoResult{
		Size:  int64(len(mempoolTxns)),
		Bytes: numBytes,
	}

	return ret, nil
}

// handleGetMiningInfo implements the getMiningInfo command. We only return the
// fields that are not related to wallet functionality.
func handleGetMiningInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if config.MainConfig().SubnetworkID != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidRequest.Code,
			Message: "`getMiningInfo` is not supported on partial nodes.",
		}
	}

	// Create a default getNetworkHashPs command to use defaults and make
	// use of the existing getNetworkHashPs handler.
	gnhpsCmd := btcjson.NewGetNetworkHashPSCmd(nil, nil)
	networkHashesPerSecIface, err := handleGetNetworkHashPS(s, gnhpsCmd,
		closeChan)
	if err != nil {
		return nil, err
	}
	networkHashesPerSec, ok := networkHashesPerSecIface.(int64)
	if !ok {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInternal.Code,
			Message: "networkHashesPerSec is not an int64",
		}
	}

	selectedTipHash := s.cfg.DAG.SelectedTipHash()
	selectedBlock, err := s.cfg.DAG.BlockByHash(selectedTipHash)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInternal.Code,
			Message: "could not find block for selected tip",
		}
	}

	result := btcjson.GetMiningInfoResult{
		Blocks:           int64(s.cfg.DAG.BlockCount()),
		CurrentBlockSize: uint64(selectedBlock.MsgBlock().SerializeSize()),
		CurrentBlockTx:   uint64(len(selectedBlock.MsgBlock().Transactions)),
		Difficulty:       getDifficultyRatio(s.cfg.DAG.CurrentBits(), s.cfg.DAGParams),
		Generate:         s.cfg.CPUMiner.IsMining(),
		GenProcLimit:     s.cfg.CPUMiner.NumWorkers(),
		HashesPerSec:     int64(s.cfg.CPUMiner.HashesPerSecond()),
		NetworkHashPS:    networkHashesPerSec,
		PooledTx:         uint64(s.cfg.TxMemPool.Count()),
		TestNet:          config.MainConfig().TestNet3,
		DevNet:           config.MainConfig().DevNet,
	}
	return &result, nil
}

// handleGetNetTotals implements the getNetTotals command.
func handleGetNetTotals(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	totalBytesRecv, totalBytesSent := s.cfg.ConnMgr.NetTotals()
	reply := &btcjson.GetNetTotalsResult{
		TotalBytesRecv: totalBytesRecv,
		TotalBytesSent: totalBytesSent,
		TimeMillis:     time.Now().UTC().UnixNano() / int64(time.Millisecond),
	}
	return reply, nil
}

// handleGetNetworkHashPS implements the getNetworkHashPs command.
// This command had been (possibly temporarily) dropped.
// Originally it relied on height, which no longer makes sense.
func handleGetNetworkHashPS(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if config.MainConfig().SubnetworkID != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidRequest.Code,
			Message: "`getNetworkHashPS` is not supported on partial nodes.",
		}
	}

	return nil, ErrRPCUnimplemented
}

// handleGetPeerInfo implements the getPeerInfo command.
func handleGetPeerInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	peers := s.cfg.ConnMgr.ConnectedPeers()
	syncPeerID := s.cfg.SyncMgr.SyncPeerID()
	infos := make([]*btcjson.GetPeerInfoResult, 0, len(peers))
	for _, p := range peers {
		statsSnap := p.ToPeer().StatsSnapshot()
		info := &btcjson.GetPeerInfoResult{
			ID:          statsSnap.ID,
			Addr:        statsSnap.Addr,
			Services:    fmt.Sprintf("%08d", uint64(statsSnap.Services)),
			RelayTxes:   !p.IsTxRelayDisabled(),
			LastSend:    statsSnap.LastSend.Unix(),
			LastRecv:    statsSnap.LastRecv.Unix(),
			BytesSent:   statsSnap.BytesSent,
			BytesRecv:   statsSnap.BytesRecv,
			ConnTime:    statsSnap.ConnTime.Unix(),
			PingTime:    float64(statsSnap.LastPingMicros),
			TimeOffset:  statsSnap.TimeOffset,
			Version:     statsSnap.Version,
			SubVer:      statsSnap.UserAgent,
			Inbound:     statsSnap.Inbound,
			SelectedTip: statsSnap.SelectedTip.String(),
			BanScore:    int32(p.BanScore()),
			FeeFilter:   p.FeeFilter(),
			SyncNode:    statsSnap.ID == syncPeerID,
		}
		if p.ToPeer().LastPingNonce() != 0 {
			wait := float64(time.Since(statsSnap.LastPingTime).Nanoseconds())
			// We actually want microseconds.
			info.PingWait = wait / 1000
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// handleGetRawMempool implements the getRawMempool command.
func handleGetRawMempool(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetRawMempoolCmd)
	mp := s.cfg.TxMemPool

	if c.Verbose != nil && *c.Verbose {
		return mp.RawMempoolVerbose(), nil
	}

	// The response is simply an array of the transaction hashes if the
	// verbose flag is not set.
	descs := mp.TxDescs()
	hashStrings := make([]string, len(descs))
	for i := range hashStrings {
		hashStrings[i] = descs[i].Tx.ID().String()
	}

	return hashStrings, nil
}

// handleGetRawTransaction implements the getRawTransaction command.
func handleGetRawTransaction(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetRawTransactionCmd)

	// Convert the provided transaction hash hex to a Hash.
	txID, err := daghash.NewTxIDFromStr(c.TxID)
	if err != nil {
		return nil, rpcDecodeHexError(c.TxID)
	}

	verbose := false
	if c.Verbose != nil {
		verbose = *c.Verbose != 0
	}

	// Try to fetch the transaction from the memory pool and if that fails,
	// try the block database.
	var mtx *wire.MsgTx
	var blkHash *daghash.Hash
	isInMempool := false
	tx, err := s.cfg.TxMemPool.FetchTransaction(txID)
	if err != nil {
		if s.cfg.TxIndex == nil {
			return nil, &btcjson.RPCError{
				Code: btcjson.ErrRPCNoTxInfo,
				Message: "The transaction index must be " +
					"enabled to query the blockchain " +
					"(specify --txindex)",
			}
		}

		// Look up the location of the transaction.
		blockRegion, err := s.cfg.TxIndex.TxFirstBlockRegion(txID)
		if err != nil {
			context := "Failed to retrieve transaction location"
			return nil, internalRPCError(err.Error(), context)
		}
		if blockRegion == nil {
			return nil, rpcNoTxInfoError(txID)
		}

		// Load the raw transaction bytes from the database.
		var txBytes []byte
		err = s.cfg.DB.View(func(dbTx database.Tx) error {
			var err error
			txBytes, err = dbTx.FetchBlockRegion(blockRegion)
			return err
		})
		if err != nil {
			return nil, rpcNoTxInfoError(txID)
		}

		// When the verbose flag isn't set, simply return the serialized
		// transaction as a hex-encoded string.  This is done here to
		// avoid deserializing it only to reserialize it again later.
		if !verbose {
			return hex.EncodeToString(txBytes), nil
		}

		// Grab the block hash.
		blkHash = blockRegion.Hash

		// Deserialize the transaction
		var msgTx wire.MsgTx
		err = msgTx.Deserialize(bytes.NewReader(txBytes))
		if err != nil {
			context := "Failed to deserialize transaction"
			return nil, internalRPCError(err.Error(), context)
		}
		mtx = &msgTx
	} else {
		// When the verbose flag isn't set, simply return the
		// network-serialized transaction as a hex-encoded string.
		if !verbose {
			// Note that this is intentionally not directly
			// returning because the first return value is a
			// string and it would result in returning an empty
			// string to the client instead of nothing (nil) in the
			// case of an error.
			mtxHex, err := messageToHex(tx.MsgTx())
			if err != nil {
				return nil, err
			}
			return mtxHex, nil
		}

		mtx = tx.MsgTx()
		isInMempool = true
	}

	// The verbose flag is set, so generate the JSON object and return it.
	var blkHeader *wire.BlockHeader
	var blkHashStr string
	if blkHash != nil {
		// Fetch the header from chain.
		header, err := s.cfg.DAG.HeaderByHash(blkHash)
		if err != nil {
			context := "Failed to fetch block header"
			return nil, internalRPCError(err.Error(), context)
		}

		blkHeader = header
		blkHashStr = blkHash.String()
	}

	confirmations, err := txConfirmations(s, mtx.TxID())
	if err != nil {
		return nil, err
	}
	rawTxn, err := createTxRawResult(s.cfg.DAGParams, mtx, txID.String(),
		blkHeader, blkHashStr, nil, &confirmations, isInMempool)
	if err != nil {
		return nil, err
	}
	return *rawTxn, nil
}

// handleGetSubnetwork handles the getSubnetwork command.
func handleGetSubnetwork(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetSubnetworkCmd)

	subnetworkID, err := subnetworkid.NewFromStr(c.SubnetworkID)
	if err != nil {
		return nil, rpcDecodeHexError(c.SubnetworkID)
	}

	gasLimit, err := s.cfg.DAG.SubnetworkStore.GasLimit(subnetworkID)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCSubnetworkNotFound,
			Message: "Subnetwork not found.",
		}
	}

	subnetworkReply := &btcjson.GetSubnetworkResult{
		GasLimit: gasLimit,
	}
	return subnetworkReply, nil
}

// handleGetTxOut handles getTxOut commands.
func handleGetTxOut(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetTxOutCmd)

	// Convert the provided transaction hash hex to a Hash.
	txID, err := daghash.NewTxIDFromStr(c.TxID)
	if err != nil {
		return nil, rpcDecodeHexError(c.TxID)
	}

	// If requested and the tx is available in the mempool try to fetch it
	// from there, otherwise attempt to fetch from the block database.
	var bestBlockHash string
	var confirmations *uint64
	var value uint64
	var scriptPubKey []byte
	var isCoinbase bool
	isInMempool := false
	includeMempool := true
	if c.IncludeMempool != nil {
		includeMempool = *c.IncludeMempool
	}
	// TODO: This is racy.  It should attempt to fetch it directly and check
	// the error.
	if includeMempool && s.cfg.TxMemPool.HaveTransaction(txID) {
		tx, err := s.cfg.TxMemPool.FetchTransaction(txID)
		if err != nil {
			return nil, rpcNoTxInfoError(txID)
		}

		mtx := tx.MsgTx()
		if c.Vout > uint32(len(mtx.TxOut)-1) {
			return nil, &btcjson.RPCError{
				Code: btcjson.ErrRPCInvalidTxVout,
				Message: "Output index number (vout) does not " +
					"exist for transaction.",
			}
		}

		txOut := mtx.TxOut[c.Vout]
		if txOut == nil {
			errStr := fmt.Sprintf("Output index: %d for txid: %s "+
				"does not exist", c.Vout, txID)
			return nil, internalRPCError(errStr, "")
		}

		bestBlockHash = s.cfg.DAG.SelectedTipHash().String()
		value = txOut.Value
		scriptPubKey = txOut.ScriptPubKey
		isCoinbase = mtx.IsCoinBase()
		isInMempool = true
	} else {
		out := wire.Outpoint{TxID: *txID, Index: c.Vout}
		entry, ok := s.cfg.DAG.GetUTXOEntry(out)
		if !ok {
			return nil, rpcNoTxInfoError(txID)
		}

		// To match the behavior of the reference client, return nil
		// (JSON null) if the transaction output is spent by another
		// transaction already in the main chain.  Mined transactions
		// that are spent by a mempool transaction are not affected by
		// this.
		if entry == nil {
			return nil, nil
		}

		if s.cfg.TxIndex != nil {
			txConfirmations, err := txConfirmations(s, txID)
			if err != nil {
				return nil, internalRPCError("Output index number (vout) does not "+
					"exist for transaction.", "")
			}
			confirmations = &txConfirmations
		}

		bestBlockHash = s.cfg.DAG.SelectedTipHash().String()
		value = entry.Amount()
		scriptPubKey = entry.ScriptPubKey()
		isCoinbase = entry.IsCoinbase()
	}

	// Disassemble script into single line printable format.
	// The disassembled string will contain [error] inline if the script
	// doesn't fully parse, so ignore the error here.
	disbuf, _ := txscript.DisasmString(scriptPubKey)

	// Get further info about the script.
	// Ignore the error here since an error means the script couldn't parse
	// and there is no additional information about it anyways.
	scriptClass, addr, _ := txscript.ExtractScriptPubKeyAddress(scriptPubKey,
		s.cfg.DAGParams)
	var address *string
	if addr != nil {
		address = btcjson.String(addr.EncodeAddress())
	}

	txOutReply := &btcjson.GetTxOutResult{
		BestBlock:     bestBlockHash,
		Confirmations: confirmations,
		IsInMempool:   isInMempool,
		Value:         util.Amount(value).ToBTC(),
		ScriptPubKey: btcjson.ScriptPubKeyResult{
			Asm:     disbuf,
			Hex:     hex.EncodeToString(scriptPubKey),
			Type:    scriptClass.String(),
			Address: address,
		},
		Coinbase: isCoinbase,
	}
	return txOutReply, nil
}

// handleHelp implements the help command.
func handleHelp(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.HelpCmd)

	// Provide a usage overview of all commands when no specific command
	// was specified.
	var command string
	if c.Command != nil {
		command = *c.Command
	}
	if command == "" {
		usage, err := s.helpCacher.rpcUsage(false)
		if err != nil {
			context := "Failed to generate RPC usage"
			return nil, internalRPCError(err.Error(), context)
		}
		return usage, nil
	}

	// Check that the command asked for is supported and implemented.  Only
	// search the main list of handlers since help should not be provided
	// for commands that are unimplemented or related to wallet
	// functionality.
	if _, ok := rpcHandlers[command]; !ok {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidParameter,
			Message: "Unknown command: " + command,
		}
	}

	// Get the help for the command.
	help, err := s.helpCacher.rpcMethodHelp(command)
	if err != nil {
		context := "Failed to generate help"
		return nil, internalRPCError(err.Error(), context)
	}
	return help, nil
}

// handlePing implements the ping command.
func handlePing(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Ask server to ping \o_
	nonce, err := random.Uint64()
	if err != nil {
		return nil, internalRPCError("Not sending ping - failed to "+
			"generate nonce: "+err.Error(), "")
	}
	s.cfg.ConnMgr.BroadcastMessage(wire.NewMsgPing(nonce))

	return nil, nil
}

// retrievedTx represents a transaction that was either loaded from the
// transaction memory pool or from the database.  When a transaction is loaded
// from the database, it is loaded with the raw serialized bytes while the
// mempool has the fully deserialized structure.  This structure therefore will
// have one of the two fields set depending on where is was retrieved from.
// This is mainly done for efficiency to avoid extra serialization steps when
// possible.
type retrievedTx struct {
	txBytes []byte
	blkHash *daghash.Hash // Only set when transaction is in a block.
	tx      *util.Tx
}

// fetchInputTxos fetches the outpoints from all transactions referenced by the
// inputs to the passed transaction by checking the transaction mempool first
// then the transaction index for those already mined into blocks.
func fetchInputTxos(s *Server, tx *wire.MsgTx) (map[wire.Outpoint]wire.TxOut, error) {
	mp := s.cfg.TxMemPool
	originOutputs := make(map[wire.Outpoint]wire.TxOut)
	for txInIndex, txIn := range tx.TxIn {
		// Attempt to fetch and use the referenced transaction from the
		// memory pool.
		origin := &txIn.PreviousOutpoint
		originTx, err := mp.FetchTransaction(&origin.TxID)
		if err == nil {
			txOuts := originTx.MsgTx().TxOut
			if origin.Index >= uint32(len(txOuts)) {
				errStr := fmt.Sprintf("unable to find output "+
					"%s referenced from transaction %s:%d",
					origin, tx.TxID(), txInIndex)
				return nil, internalRPCError(errStr, "")
			}

			originOutputs[*origin] = *txOuts[origin.Index]
			continue
		}

		// Look up the location of the transaction.
		blockRegion, err := s.cfg.TxIndex.TxFirstBlockRegion(&origin.TxID)
		if err != nil {
			context := "Failed to retrieve transaction location"
			return nil, internalRPCError(err.Error(), context)
		}
		if blockRegion == nil {
			return nil, rpcNoTxInfoError(&origin.TxID)
		}

		// Load the raw transaction bytes from the database.
		var txBytes []byte
		err = s.cfg.DB.View(func(dbTx database.Tx) error {
			var err error
			txBytes, err = dbTx.FetchBlockRegion(blockRegion)
			return err
		})
		if err != nil {
			return nil, rpcNoTxInfoError(&origin.TxID)
		}

		// Deserialize the transaction
		var msgTx wire.MsgTx
		err = msgTx.Deserialize(bytes.NewReader(txBytes))
		if err != nil {
			context := "Failed to deserialize transaction"
			return nil, internalRPCError(err.Error(), context)
		}

		// Add the referenced output to the map.
		if origin.Index >= uint32(len(msgTx.TxOut)) {
			errStr := fmt.Sprintf("unable to find output %s "+
				"referenced from transaction %s:%d", origin,
				tx.TxID(), txInIndex)
			return nil, internalRPCError(errStr, "")
		}
		originOutputs[*origin] = *msgTx.TxOut[origin.Index]
	}

	return originOutputs, nil
}

// txConfirmations returns the confirmations number for the given transaction
// The confirmations number is defined as follows:
// If the transaction is in the mempool/in a red block/is a double spend -> 0
// Otherwise -> The confirmations number of the accepting block
func txConfirmations(s *Server, txID *daghash.TxID) (uint64, error) {
	if s.cfg.TxIndex == nil {
		return 0, errors.New("transaction index must be enabled (--txindex)")
	}

	acceptingBlock, err := s.cfg.TxIndex.BlockThatAcceptedTx(s.cfg.DAG, txID)
	if err != nil {
		return 0, fmt.Errorf("could not get block that accepted tx %s: %s", txID, err)
	}
	if acceptingBlock == nil {
		return 0, nil
	}

	confirmations, err := s.cfg.DAG.BlockConfirmationsByHash(acceptingBlock)
	if err != nil {
		return 0, fmt.Errorf("could not get confirmations for block that accepted tx %s: %s", txID, err)
	}

	return confirmations, nil
}

// createVinListPrevOut returns a slice of JSON objects for the inputs of the
// passed transaction.
func createVinListPrevOut(s *Server, mtx *wire.MsgTx, chainParams *dagconfig.Params, vinExtra bool, filterAddrMap map[string]struct{}) ([]btcjson.VinPrevOut, error) {
	// Use a dynamically sized list to accommodate the address filter.
	vinList := make([]btcjson.VinPrevOut, 0, len(mtx.TxIn))

	// Lookup all of the referenced transaction outputs needed to populate the
	// previous output information if requested. Coinbase transactions do not contain
	// valid inputs: block hash instead of transaction ID.
	var originOutputs map[wire.Outpoint]wire.TxOut
	if !mtx.IsCoinBase() && (vinExtra || len(filterAddrMap) > 0) {
		var err error
		originOutputs, err = fetchInputTxos(s, mtx)
		if err != nil {
			return nil, err
		}
	}

	for _, txIn := range mtx.TxIn {
		// The disassembled string will contain [error] inline
		// if the script doesn't fully parse, so ignore the
		// error here.
		disbuf, _ := txscript.DisasmString(txIn.SignatureScript)

		// Create the basic input entry without the additional optional
		// previous output details which will be added later if
		// requested and available.
		prevOut := &txIn.PreviousOutpoint
		vinEntry := btcjson.VinPrevOut{
			TxID:     prevOut.TxID.String(),
			Vout:     prevOut.Index,
			Sequence: txIn.Sequence,
			ScriptSig: &btcjson.ScriptSig{
				Asm: disbuf,
				Hex: hex.EncodeToString(txIn.SignatureScript),
			},
		}

		// Add the entry to the list now if it already passed the filter
		// since the previous output might not be available.
		passesFilter := len(filterAddrMap) == 0
		if passesFilter {
			vinList = append(vinList, vinEntry)
		}

		// Only populate previous output information if requested and
		// available.
		if len(originOutputs) == 0 {
			continue
		}
		originTxOut, ok := originOutputs[*prevOut]
		if !ok {
			continue
		}

		// Ignore the error here since an error means the script
		// couldn't parse and there is no additional information about
		// it anyways.
		_, addr, _ := txscript.ExtractScriptPubKeyAddress(
			originTxOut.ScriptPubKey, chainParams)

		var encodedAddr *string
		if addr != nil {
			// Encode the address while checking if the address passes the
			// filter when needed.
			encodedAddr = btcjson.String(addr.EncodeAddress())

			// If the filter doesn't already pass, make it pass if
			// the address exists in the filter.
			if _, exists := filterAddrMap[*encodedAddr]; exists {
				passesFilter = true
			}
		}

		// Ignore the entry if it doesn't pass the filter.
		if !passesFilter {
			continue
		}

		// Add entry to the list if it wasn't already done above.
		if len(filterAddrMap) != 0 {
			vinList = append(vinList, vinEntry)
		}

		// Update the entry with previous output information if
		// requested.
		if vinExtra {
			vinListEntry := &vinList[len(vinList)-1]
			vinListEntry.PrevOut = &btcjson.PrevOut{
				Address: encodedAddr,
				Value:   util.Amount(originTxOut.Value).ToBTC(),
			}
		}
	}

	return vinList, nil
}

// fetchMempoolTxnsForAddress queries the address index for all unconfirmed
// transactions that involve the provided address.  The results will be limited
// by the number to skip and the number requested.
func fetchMempoolTxnsForAddress(s *Server, addr util.Address, numToSkip, numRequested uint32) ([]*util.Tx, uint32) {
	// There are no entries to return when there are less available than the
	// number being skipped.
	mpTxns := s.cfg.AddrIndex.UnconfirmedTxnsForAddress(addr)
	numAvailable := uint32(len(mpTxns))
	if numToSkip > numAvailable {
		return nil, numAvailable
	}

	// Filter the available entries based on the number to skip and number
	// requested.
	rangeEnd := numToSkip + numRequested
	if rangeEnd > numAvailable {
		rangeEnd = numAvailable
	}
	return mpTxns[numToSkip:rangeEnd], numToSkip
}

// handleSearchRawTransactions implements the searchRawTransactions command.
func handleSearchRawTransactions(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Respond with an error if the address index is not enabled.
	addrIndex := s.cfg.AddrIndex
	if addrIndex == nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCMisc,
			Message: "Address index must be enabled (--addrindex)",
		}
	}

	// Override the flag for including extra previous output information in
	// each input if needed.
	c := cmd.(*btcjson.SearchRawTransactionsCmd)
	vinExtra := false
	if c.VinExtra != nil {
		vinExtra = *c.VinExtra
	}

	// Including the extra previous output information requires the
	// transaction index.  Currently the address index relies on the
	// transaction index, so this check is redundant, but it's better to be
	// safe in case the address index is ever changed to not rely on it.
	if vinExtra && s.cfg.TxIndex == nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCMisc,
			Message: "Transaction index must be enabled (--txindex)",
		}
	}

	// Attempt to decode the supplied address.
	params := s.cfg.DAGParams
	addr, err := util.DecodeAddress(c.Address, params.Prefix)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidAddressOrKey,
			Message: "Invalid address or key: " + err.Error(),
		}
	}

	// Override the default number of requested entries if needed.  Also,
	// just return now if the number of requested entries is zero to avoid
	// extra work.
	numRequested := 100
	if c.Count != nil {
		numRequested = *c.Count
		if numRequested < 0 {
			numRequested = 1
		}
	}
	if numRequested == 0 {
		return nil, nil
	}

	// Override the default number of entries to skip if needed.
	var numToSkip int
	if c.Skip != nil {
		numToSkip = *c.Skip
		if numToSkip < 0 {
			numToSkip = 0
		}
	}

	// Override the reverse flag if needed.
	var reverse bool
	if c.Reverse != nil {
		reverse = *c.Reverse
	}

	// Add transactions from mempool first if client asked for reverse
	// order.  Otherwise, they will be added last (as needed depending on
	// the requested counts).
	//
	// NOTE: This code doesn't sort by dependency.  This might be something
	// to do in the future for the client's convenience, or leave it to the
	// client.
	numSkipped := uint32(0)
	addressTxns := make([]retrievedTx, 0, numRequested)
	if reverse {
		// Transactions in the mempool are not in a block header yet,
		// so the block header field in the retieved transaction struct
		// is left nil.
		mpTxns, mpSkipped := fetchMempoolTxnsForAddress(s, addr,
			uint32(numToSkip), uint32(numRequested))
		numSkipped += mpSkipped
		for _, tx := range mpTxns {
			addressTxns = append(addressTxns, retrievedTx{tx: tx})
		}
	}

	// Fetch transactions from the database in the desired order if more are
	// needed.
	if len(addressTxns) < numRequested {
		err = s.cfg.DB.View(func(dbTx database.Tx) error {
			regions, dbSkipped, err := addrIndex.TxRegionsForAddress(
				dbTx, addr, uint32(numToSkip)-numSkipped,
				uint32(numRequested-len(addressTxns)), reverse)
			if err != nil {
				return err
			}

			// Load the raw transaction bytes from the database.
			serializedTxns, err := dbTx.FetchBlockRegions(regions)
			if err != nil {
				return err
			}

			// Add the transaction and the hash of the block it is
			// contained in to the list.  Note that the transaction
			// is left serialized here since the caller might have
			// requested non-verbose output and hence there would be
			// no point in deserializing it just to reserialize it
			// later.
			for i, serializedTx := range serializedTxns {
				addressTxns = append(addressTxns, retrievedTx{
					txBytes: serializedTx,
					blkHash: regions[i].Hash,
				})
			}
			numSkipped += dbSkipped

			return nil
		})
		if err != nil {
			context := "Failed to load address index entries"
			return nil, internalRPCError(err.Error(), context)
		}

	}

	// Add transactions from mempool last if client did not request reverse
	// order and the number of results is still under the number requested.
	if !reverse && len(addressTxns) < numRequested {
		// Transactions in the mempool are not in a block header yet,
		// so the block header field in the retieved transaction struct
		// is left nil.
		mpTxns, mpSkipped := fetchMempoolTxnsForAddress(s, addr,
			uint32(numToSkip)-numSkipped, uint32(numRequested-
				len(addressTxns)))
		numSkipped += mpSkipped
		for _, tx := range mpTxns {
			addressTxns = append(addressTxns, retrievedTx{tx: tx})
		}
	}

	// Address has never been used if neither source yielded any results.
	if len(addressTxns) == 0 {
		return []btcjson.SearchRawTransactionsResult{}, nil
	}

	// Serialize all of the transactions to hex.
	hexTxns := make([]string, len(addressTxns))
	for i := range addressTxns {
		// Simply encode the raw bytes to hex when the retrieved
		// transaction is already in serialized form.
		rtx := &addressTxns[i]
		if rtx.txBytes != nil {
			hexTxns[i] = hex.EncodeToString(rtx.txBytes)
			continue
		}

		// Serialize the transaction first and convert to hex when the
		// retrieved transaction is the deserialized structure.
		hexTxns[i], err = messageToHex(rtx.tx.MsgTx())
		if err != nil {
			return nil, err
		}
	}

	// When not in verbose mode, simply return a list of serialized txns.
	if c.Verbose != nil && !*c.Verbose {
		return hexTxns, nil
	}

	// Normalize the provided filter addresses (if any) to ensure there are
	// no duplicates.
	filterAddrMap := make(map[string]struct{})
	if c.FilterAddrs != nil && len(*c.FilterAddrs) > 0 {
		for _, addr := range *c.FilterAddrs {
			filterAddrMap[addr] = struct{}{}
		}
	}

	// The verbose flag is set, so generate the JSON object and return it.
	srtList := make([]btcjson.SearchRawTransactionsResult, len(addressTxns))
	for i := range addressTxns {
		// The deserialized transaction is needed, so deserialize the
		// retrieved transaction if it's in serialized form (which will
		// be the case when it was lookup up from the database).
		// Otherwise, use the existing deserialized transaction.
		rtx := &addressTxns[i]
		var mtx *wire.MsgTx
		if rtx.tx == nil {
			// Deserialize the transaction.
			mtx = new(wire.MsgTx)
			err := mtx.Deserialize(bytes.NewReader(rtx.txBytes))
			if err != nil {
				context := "Failed to deserialize transaction"
				return nil, internalRPCError(err.Error(),
					context)
			}
		} else {
			mtx = rtx.tx.MsgTx()
		}

		result := &srtList[i]
		result.Hex = hexTxns[i]
		result.TxID = mtx.TxID().String()
		result.Vin, err = createVinListPrevOut(s, mtx, params, vinExtra,
			filterAddrMap)
		if err != nil {
			return nil, err
		}
		result.Vout = createVoutList(mtx, params, filterAddrMap)
		result.Version = mtx.Version
		result.LockTime = mtx.LockTime

		// Transactions grabbed from the mempool aren't yet in a block,
		// so conditionally fetch block details here.  This will be
		// reflected in the final JSON output (mempool won't have
		// confirmations or block information).
		var blkHeader *wire.BlockHeader
		var blkHashStr string
		if blkHash := rtx.blkHash; blkHash != nil {
			// Fetch the header from chain.
			header, err := s.cfg.DAG.HeaderByHash(blkHash)
			if err != nil {
				return nil, &btcjson.RPCError{
					Code:    btcjson.ErrRPCBlockNotFound,
					Message: "Block not found",
				}
			}

			blkHeader = header
			blkHashStr = blkHash.String()
		}

		// Add the block information to the result if there is any.
		if blkHeader != nil {
			// This is not a typo, they are identical in Bitcoin
			// Core as well.
			result.Time = uint64(blkHeader.Timestamp.Unix())
			result.Blocktime = uint64(blkHeader.Timestamp.Unix())
			result.BlockHash = blkHashStr
		}

		// rtx.tx is only set when the transaction was retrieved from the mempool
		result.IsInMempool = rtx.tx != nil

		if s.cfg.TxIndex != nil && !result.IsInMempool {
			confirmations, err := txConfirmations(s, mtx.TxID())
			if err != nil {
				context := "Failed to obtain block confirmations"
				return nil, internalRPCError(err.Error(), context)
			}
			result.Confirmations = &confirmations
		}
	}

	return srtList, nil
}

// handleSendRawTransaction implements the sendRawTransaction command.
func handleSendRawTransaction(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.SendRawTransactionCmd)
	// Deserialize and send off to tx relay
	hexStr := c.HexTx
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	serializedTx, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpcDecodeHexError(hexStr)
	}
	var msgTx wire.MsgTx
	err = msgTx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCDeserialization,
			Message: "TX decode failed: " + err.Error(),
		}
	}

	// Use 0 for the tag to represent local node.
	tx := util.NewTx(&msgTx)
	acceptedTxs, err := s.cfg.TxMemPool.ProcessTransaction(tx, false, 0)
	if err != nil {
		// When the error is a rule error, it means the transaction was
		// simply rejected as opposed to something actually going wrong,
		// so log it as such.  Otherwise, something really did go wrong,
		// so log it as an actual error.  In both cases, a JSON-RPC
		// error is returned to the client with the deserialization
		// error code (to match bitcoind behavior).
		if _, ok := err.(mempool.RuleError); ok {
			log.Debugf("Rejected transaction %s: %s", tx.ID(),
				err)
		} else {
			log.Errorf("Failed to process transaction %s: %s",
				tx.ID(), err)
		}
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCVerify,
			Message: "TX rejected: " + err.Error(),
		}
	}

	// When the transaction was accepted it should be the first item in the
	// returned array of accepted transactions.  The only way this will not
	// be true is if the API for ProcessTransaction changes and this code is
	// not properly updated, but ensure the condition holds as a safeguard.
	//
	// Also, since an error is being returned to the caller, ensure the
	// transaction is removed from the memory pool.
	if len(acceptedTxs) == 0 || !acceptedTxs[0].Tx.ID().IsEqual(tx.ID()) {
		err := s.cfg.TxMemPool.RemoveTransaction(tx, true, true)
		if err != nil {
			return nil, err
		}

		errStr := fmt.Sprintf("transaction %s is not in accepted list",
			tx.ID())
		return nil, internalRPCError(errStr, "")
	}

	// Generate and relay inventory vectors for all newly accepted
	// transactions into the memory pool due to the original being
	// accepted.
	s.cfg.ConnMgr.RelayTransactions(acceptedTxs)

	// Notify both websocket and getBlockTemplate long poll clients of all
	// newly accepted transactions.
	s.NotifyNewTransactions(acceptedTxs)

	// Keep track of all the sendRawTransaction request txns so that they
	// can be rebroadcast if they don't make their way into a block.
	txD := acceptedTxs[0]
	iv := wire.NewInvVect(wire.InvTypeTx, (*daghash.Hash)(txD.Tx.ID()))
	s.cfg.ConnMgr.AddRebroadcastInventory(iv, txD)

	return tx.ID().String(), nil
}

// handleSetGenerate implements the setGenerate command.
func handleSetGenerate(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if config.MainConfig().SubnetworkID != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidRequest.Code,
			Message: "`setGenerate` is not supported on partial nodes.",
		}
	}

	c := cmd.(*btcjson.SetGenerateCmd)

	// Disable generation regardless of the provided generate flag if the
	// maximum number of threads (goroutines for our purposes) is 0.
	// Otherwise enable or disable it depending on the provided flag.
	generate := c.Generate
	genProcLimit := -1
	if c.GenProcLimit != nil {
		genProcLimit = *c.GenProcLimit
	}
	if genProcLimit == 0 {
		generate = false
	}

	if !generate {
		s.cfg.CPUMiner.Stop()
	} else {
		// Respond with an error if there are no addresses to pay the
		// created blocks to.
		if len(config.MainConfig().MiningAddrs) == 0 {
			return nil, &btcjson.RPCError{
				Code: btcjson.ErrRPCInternal.Code,
				Message: "No payment addresses specified " +
					"via --miningaddr",
			}
		}

		// It's safe to call start even if it's already started.
		s.cfg.CPUMiner.SetNumWorkers(int32(genProcLimit))
		s.cfg.CPUMiner.Start()
	}
	return nil, nil
}

// handleStop implements the stop command.
func handleStop(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	select {
	case s.requestProcessShutdown <- struct{}{}:
	default:
	}
	return "btcd stopping.", nil
}

// handleSubmitBlock implements the submitBlock command.
func handleSubmitBlock(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.SubmitBlockCmd)

	// Deserialize the submitted block.
	hexStr := c.HexBlock
	if len(hexStr)%2 != 0 {
		hexStr = "0" + c.HexBlock
	}
	serializedBlock, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpcDecodeHexError(hexStr)
	}

	block, err := util.NewBlockFromBytes(serializedBlock)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCDeserialization,
			Message: "Block decode failed: " + err.Error(),
		}
	}

	// Process this block using the same rules as blocks coming from other
	// nodes.  This will in turn relay it to the network like normal.
	_, err = s.cfg.SyncMgr.SubmitBlock(block, blockdag.BFNone)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCVerify,
			Message: fmt.Sprintf("Block rejected. Reason: %s", err),
		}
	}

	log.Infof("Accepted block %s via submitBlock", block.Hash())
	return nil, nil
}

// handleUptime implements the uptime command.
func handleUptime(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return time.Now().Unix() - s.cfg.StartupTime, nil
}

// handleValidateAddress implements the validateAddress command.
func handleValidateAddress(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.ValidateAddressCmd)

	result := btcjson.ValidateAddressResult{}
	addr, err := util.DecodeAddress(c.Address, s.cfg.DAGParams.Prefix)
	if err != nil {
		// Return the default value (false) for IsValid.
		return result, nil
	}

	result.Address = addr.EncodeAddress()
	result.IsValid = true

	return result, nil
}

// handleVerifyMessage implements the verifyMessage command.
func handleVerifyMessage(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.VerifyMessageCmd)

	// Decode the provided address.
	params := s.cfg.DAGParams
	addr, err := util.DecodeAddress(c.Address, params.Prefix)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidAddressOrKey,
			Message: "Invalid address or key: " + err.Error(),
		}
	}

	// Only P2PKH addresses are valid for signing.
	if _, ok := addr.(*util.AddressPubKeyHash); !ok {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCType,
			Message: "Address is not a pay-to-pubkey-hash address",
		}
	}

	// Decode base64 signature.
	sig, err := base64.StdEncoding.DecodeString(c.Signature)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCParse.Code,
			Message: "Malformed base64 encoding: " + err.Error(),
		}
	}

	// Validate the signature - this just shows that it was valid at all.
	// we will compare it with the key next.
	var buf bytes.Buffer
	wire.WriteVarString(&buf, "Bitcoin Signed Message:\n")
	wire.WriteVarString(&buf, c.Message)
	expectedMessageHash := daghash.DoubleHashB(buf.Bytes())
	pk, wasCompressed, err := btcec.RecoverCompact(btcec.S256(), sig,
		expectedMessageHash)
	if err != nil {
		// Mirror Bitcoin Core behavior, which treats error in
		// RecoverCompact as invalid signature.
		return false, nil
	}

	// Reconstruct the pubkey hash.
	var serializedPK []byte
	if wasCompressed {
		serializedPK = pk.SerializeCompressed()
	} else {
		serializedPK = pk.SerializeUncompressed()
	}
	address, err := util.NewAddressPubKeyHashFromPublicKey(serializedPK, params.Prefix)
	if err != nil {
		// Again mirror Bitcoin Core behavior, which treats error in public key
		// reconstruction as invalid signature.
		return false, nil
	}

	// Return boolean if addresses match.
	return address.EncodeAddress() == c.Address, nil
}

// handleVersion implements the version command.
//
// NOTE: This is a btcsuite extension ported from github.com/decred/dcrd.
func handleVersion(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := map[string]btcjson.VersionResult{
		"btcdjsonrpcapi": {
			VersionString: jsonrpcSemverString,
			Major:         jsonrpcSemverMajor,
			Minor:         jsonrpcSemverMinor,
			Patch:         jsonrpcSemverPatch,
		},
	}
	return result, nil
}

// handleFlushDBCache flushes the db cache to the disk.
// TODO: (Ori) This is a temporary function for dev use. It needs to be removed.
func handleFlushDBCache(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	err := s.cfg.DB.FlushCache()
	return nil, err
}

// Server provides a concurrent safe RPC server to a chain server.
type Server struct {
	started                int32
	shutdown               int32
	cfg                    rpcserverConfig
	authsha                [sha256.Size]byte
	limitauthsha           [sha256.Size]byte
	ntfnMgr                *wsNotificationManager
	numClients             int32
	statusLines            map[int]string
	statusLock             sync.RWMutex
	wg                     sync.WaitGroup
	gbtWorkState           *gbtWorkState
	helpCacher             *helpCacher
	requestProcessShutdown chan struct{}
	quit                   chan int
}

// httpStatusLine returns a response Status-Line (RFC 2616 Section 6.1)
// for the given request and response status code.  This function was lifted and
// adapted from the standard library HTTP server code since it's not exported.
func (s *Server) httpStatusLine(req *http.Request, code int) string {
	// Fast path:
	key := code
	proto11 := req.ProtoAtLeast(1, 1)
	if !proto11 {
		key = -key
	}
	s.statusLock.RLock()
	line, ok := s.statusLines[key]
	s.statusLock.RUnlock()
	if ok {
		return line
	}

	// Slow path:
	proto := "HTTP/1.0"
	if proto11 {
		proto = "HTTP/1.1"
	}
	codeStr := strconv.Itoa(code)
	text := http.StatusText(code)
	if text != "" {
		line = proto + " " + codeStr + " " + text + "\r\n"
		s.statusLock.Lock()
		s.statusLines[key] = line
		s.statusLock.Unlock()
	} else {
		text = "status code " + codeStr
		line = proto + " " + codeStr + " " + text + "\r\n"
	}

	return line
}

// writeHTTPResponseHeaders writes the necessary response headers prior to
// writing an HTTP body given a request to use for protocol negotiation, headers
// to write, a status code, and a writer.
func (s *Server) writeHTTPResponseHeaders(req *http.Request, headers http.Header, code int, w io.Writer) error {
	_, err := io.WriteString(w, s.httpStatusLine(req, code))
	if err != nil {
		return err
	}

	err = headers.Write(w)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, "\r\n")
	return err
}

// Stop is used by server.go to stop the rpc listener.
func (s *Server) Stop() error {
	if atomic.AddInt32(&s.shutdown, 1) != 1 {
		log.Infof("RPC server is already in the process of shutting down")
		return nil
	}
	log.Warnf("RPC server shutting down")
	for _, listener := range s.cfg.Listeners {
		err := listener.Close()
		if err != nil {
			log.Errorf("Problem shutting down rpc: %s", err)
			return err
		}
	}
	s.ntfnMgr.Shutdown()
	s.ntfnMgr.WaitForShutdown()
	close(s.quit)
	s.wg.Wait()
	log.Infof("RPC server shutdown complete")
	return nil
}

// RequestedProcessShutdown returns a channel that is sent to when an authorized
// RPC client requests the process to shutdown.  If the request can not be read
// immediately, it is dropped.
func (s *Server) RequestedProcessShutdown() <-chan struct{} {
	return s.requestProcessShutdown
}

// NotifyNewTransactions notifies both websocket and getBlockTemplate long
// poll clients of the passed transactions.  This function should be called
// whenever new transactions are added to the mempool.
func (s *Server) NotifyNewTransactions(txns []*mempool.TxDesc) {
	for _, txD := range txns {
		// Notify websocket clients about mempool transactions.
		s.ntfnMgr.NotifyMempoolTx(txD.Tx, true)

		// Potentially notify any getBlockTemplate long poll clients
		// about stale block templates due to the new transaction.
		s.gbtWorkState.NotifyMempoolTx(s.cfg.TxMemPool.LastUpdated())
	}
}

// limitConnections responds with a 503 service unavailable and returns true if
// adding another client would exceed the maximum allow RPC clients.
//
// This function is safe for concurrent access.
func (s *Server) limitConnections(w http.ResponseWriter, remoteAddr string) bool {
	if int(atomic.LoadInt32(&s.numClients)+1) > config.MainConfig().RPCMaxClients {
		log.Infof("Max RPC clients exceeded [%d] - "+
			"disconnecting client %s", config.MainConfig().RPCMaxClients,
			remoteAddr)
		http.Error(w, "503 Too busy.  Try again later.",
			http.StatusServiceUnavailable)
		return true
	}
	return false
}

// incrementClients adds one to the number of connected RPC clients.  Note
// this only applies to standard clients.  Websocket clients have their own
// limits and are tracked separately.
//
// This function is safe for concurrent access.
func (s *Server) incrementClients() {
	atomic.AddInt32(&s.numClients, 1)
}

// decrementClients subtracts one from the number of connected RPC clients.
// Note this only applies to standard clients.  Websocket clients have their own
// limits and are tracked separately.
//
// This function is safe for concurrent access.
func (s *Server) decrementClients() {
	atomic.AddInt32(&s.numClients, -1)
}

// checkAuth checks the HTTP Basic authentication supplied by a wallet
// or RPC client in the HTTP request r.  If the supplied authentication
// does not match the username and password expected, a non-nil error is
// returned.
//
// This check is time-constant.
//
// The first bool return value signifies auth success (true if successful) and
// the second bool return value specifies whether the user can change the state
// of the server (true) or whether the user is limited (false). The second is
// always false if the first is.
func (s *Server) checkAuth(r *http.Request, require bool) (bool, bool, error) {
	authhdr := r.Header["Authorization"]
	if len(authhdr) <= 0 {
		if require {
			log.Warnf("RPC authentication failure from %s",
				r.RemoteAddr)
			return false, false, errors.New("auth failure")
		}

		return false, false, nil
	}

	authsha := sha256.Sum256([]byte(authhdr[0]))

	// Check for limited auth first as in environments with limited users, those
	// are probably expected to have a higher volume of calls
	limitcmp := subtle.ConstantTimeCompare(authsha[:], s.limitauthsha[:])
	if limitcmp == 1 {
		return true, false, nil
	}

	// Check for admin-level auth
	cmp := subtle.ConstantTimeCompare(authsha[:], s.authsha[:])
	if cmp == 1 {
		return true, true, nil
	}

	// Request's auth doesn't match either user
	log.Warnf("RPC authentication failure from %s", r.RemoteAddr)
	return false, false, errors.New("auth failure")
}

// parsedRPCCmd represents a JSON-RPC request object that has been parsed into
// a known concrete command along with any error that might have happened while
// parsing it.
type parsedRPCCmd struct {
	id     interface{}
	method string
	cmd    interface{}
	err    *btcjson.RPCError
}

// standardCmdResult checks that a parsed command is a standard Bitcoin JSON-RPC
// command and runs the appropriate handler to reply to the command.  Any
// commands which are not recognized or not implemented will return an error
// suitable for use in replies.
func (s *Server) standardCmdResult(cmd *parsedRPCCmd, closeChan <-chan struct{}) (interface{}, error) {
	handler, ok := rpcHandlers[cmd.method]
	if ok {
		goto handled
	}
	_, ok = rpcUnimplemented[cmd.method]
	if ok {
		handler = handleUnimplemented
		goto handled
	}
	return nil, btcjson.ErrRPCMethodNotFound
handled:

	return handler(s, cmd.cmd, closeChan)
}

// parseCmd parses a JSON-RPC request object into known concrete command.  The
// err field of the returned parsedRPCCmd struct will contain an RPC error that
// is suitable for use in replies if the command is invalid in some way such as
// an unregistered command or invalid parameters.
func parseCmd(request *btcjson.Request) *parsedRPCCmd {
	var parsedCmd parsedRPCCmd
	parsedCmd.id = request.ID
	parsedCmd.method = request.Method

	cmd, err := btcjson.UnmarshalCmd(request)
	if err != nil {
		// When the error is because the method is not registered,
		// produce a method not found RPC error.
		if jerr, ok := err.(btcjson.Error); ok &&
			jerr.ErrorCode == btcjson.ErrUnregisteredMethod {

			parsedCmd.err = btcjson.ErrRPCMethodNotFound
			return &parsedCmd
		}

		// Otherwise, some type of invalid parameters is the
		// cause, so produce the equivalent RPC error.
		parsedCmd.err = btcjson.NewRPCError(
			btcjson.ErrRPCInvalidParams.Code, err.Error())
		return &parsedCmd
	}

	parsedCmd.cmd = cmd
	return &parsedCmd
}

// createMarshalledReply returns a new marshalled JSON-RPC response given the
// passed parameters.  It will automatically convert errors that are not of
// the type *btcjson.RPCError to the appropriate type as needed.
func createMarshalledReply(id, result interface{}, replyErr error) ([]byte, error) {
	var jsonErr *btcjson.RPCError
	if replyErr != nil {
		if jErr, ok := replyErr.(*btcjson.RPCError); ok {
			jsonErr = jErr
		} else {
			jsonErr = internalRPCError(replyErr.Error(), "")
		}
	}

	return btcjson.MarshalResponse(id, result, jsonErr)
}

// jsonRPCRead handles reading and responding to RPC messages.
func (s *Server) jsonRPCRead(w http.ResponseWriter, r *http.Request, isAdmin bool) {
	if atomic.LoadInt32(&s.shutdown) != 0 {
		return
	}

	// Read and close the JSON-RPC request body from the caller.
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error reading JSON message: %s",
			errCode, err), errCode)
		return
	}

	// Unfortunately, the http server doesn't provide the ability to
	// change the read deadline for the new connection and having one breaks
	// long polling.  However, not having a read deadline on the initial
	// connection would mean clients can connect and idle forever.  Thus,
	// hijack the connecton from the HTTP server, clear the read deadline,
	// and handle writing the response manually.
	hj, ok := w.(http.Hijacker)
	if !ok {
		errMsg := "webserver doesn't support hijacking"
		log.Warnf(errMsg)
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+errMsg, errCode)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		log.Warnf("Failed to hijack HTTP connection: %s", err)
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+err.Error(), errCode)
		return
	}
	defer conn.Close()
	defer buf.Flush()
	conn.SetReadDeadline(timeZeroVal)

	// Attempt to parse the raw body into a JSON-RPC request.
	var responseID interface{}
	var jsonErr error
	var result interface{}
	var request btcjson.Request
	if err := json.Unmarshal(body, &request); err != nil {
		jsonErr = &btcjson.RPCError{
			Code:    btcjson.ErrRPCParse.Code,
			Message: "Failed to parse request: " + err.Error(),
		}
	}
	if jsonErr == nil {
		// The JSON-RPC 1.0 spec defines that notifications must have their "id"
		// set to null and states that notifications do not have a response.
		//
		// A JSON-RPC 2.0 notification is a request with "json-rpc":"2.0", and
		// without an "id" member. The specification states that notifications
		// must not be responded to. JSON-RPC 2.0 permits the null value as a
		// valid request id, therefore such requests are not notifications.
		//
		// Bitcoin Core serves requests with "id":null or even an absent "id",
		// and responds to such requests with "id":null in the response.
		//
		// Btcd does not respond to any request without and "id" or "id":null,
		// regardless the indicated JSON-RPC protocol version unless RPC quirks
		// are enabled. With RPC quirks enabled, such requests will be responded
		// to if the reqeust does not indicate JSON-RPC version.
		//
		// RPC quirks can be enabled by the user to avoid compatibility issues
		// with software relying on Core's behavior.
		if request.ID == nil && !(config.MainConfig().RPCQuirks && request.JSONRPC == "") {
			return
		}

		// The parse was at least successful enough to have an ID so
		// set it for the response.
		responseID = request.ID

		// Setup a close notifier.  Since the connection is hijacked,
		// the CloseNotifer on the ResponseWriter is not available.
		closeChan := make(chan struct{}, 1)
		spawn(func() {
			_, err := conn.Read(make([]byte, 1))
			if err != nil {
				close(closeChan)
			}
		})

		// Check if the user is limited and set error if method unauthorized
		if !isAdmin {
			if _, ok := rpcLimited[request.Method]; !ok {
				jsonErr = &btcjson.RPCError{
					Code:    btcjson.ErrRPCInvalidParams.Code,
					Message: "limited user not authorized for this method",
				}
			}
		}

		if jsonErr == nil {
			// Attempt to parse the JSON-RPC request into a known concrete
			// command.
			parsedCmd := parseCmd(&request)
			if parsedCmd.err != nil {
				jsonErr = parsedCmd.err
			} else {
				result, jsonErr = s.standardCmdResult(parsedCmd, closeChan)
			}
		}
	}

	// Marshal the response.
	msg, err := createMarshalledReply(responseID, result, jsonErr)
	if err != nil {
		log.Errorf("Failed to marshal reply: %s", err)
		return
	}

	// Write the response.
	err = s.writeHTTPResponseHeaders(r, w.Header(), http.StatusOK, buf)
	if err != nil {
		log.Error(err)
		return
	}
	if _, err := buf.Write(msg); err != nil {
		log.Errorf("Failed to write marshalled reply: %s", err)
	}

	// Terminate with newline to maintain compatibility with Bitcoin Core.
	if err := buf.WriteByte('\n'); err != nil {
		log.Errorf("Failed to append terminating newline to reply: %s", err)
	}
}

// jsonAuthFail sends a message back to the client if the http auth is rejected.
func jsonAuthFail(w http.ResponseWriter) {
	w.Header().Add("WWW-Authenticate", `Basic realm="btcd RPC"`)
	http.Error(w, "401 Unauthorized.", http.StatusUnauthorized)
}

// Start is used by server.go to start the rpc listener.
func (s *Server) Start() {
	if atomic.AddInt32(&s.started, 1) != 1 {
		return
	}

	log.Trace("Starting RPC server")
	rpcServeMux := http.NewServeMux()
	httpServer := &http.Server{
		Handler: rpcServeMux,

		// Timeout connections which don't complete the initial
		// handshake within the allowed timeframe.
		ReadTimeout: time.Second * rpcAuthTimeoutSeconds,
	}
	rpcServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.Header().Set("Content-Type", "application/json")
		r.Close = true

		// Limit the number of connections to max allowed.
		if s.limitConnections(w, r.RemoteAddr) {
			return
		}

		// Keep track of the number of connected clients.
		s.incrementClients()
		defer s.decrementClients()
		_, isAdmin, err := s.checkAuth(r, true)
		if err != nil {
			jsonAuthFail(w)
			return
		}

		// Read and respond to the request.
		s.jsonRPCRead(w, r, isAdmin)
	})

	// Websocket endpoint.
	rpcServeMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		authenticated, isAdmin, err := s.checkAuth(r, false)
		if err != nil {
			jsonAuthFail(w)
			return
		}

		// Attempt to upgrade the connection to a websocket connection
		// using the default size for read/write buffers.
		ws, err := websocket.Upgrade(w, r, nil, 0, 0)
		if err != nil {
			if _, ok := err.(websocket.HandshakeError); !ok {
				log.Errorf("Unexpected websocket error: %s",
					err)
			}
			http.Error(w, "400 Bad Request.", http.StatusBadRequest)
			return
		}
		s.WebsocketHandler(ws, r.RemoteAddr, authenticated, isAdmin)
	})

	for _, listener := range s.cfg.Listeners {
		s.wg.Add(1)
		go func(listener net.Listener) {
			log.Infof("RPC server listening on %s", listener.Addr())
			httpServer.Serve(listener)
			log.Tracef("RPC listener done for %s", listener.Addr())
			s.wg.Done()
		}(listener)
	}

	s.ntfnMgr.Start()
}

// rpcserverPeer represents a peer for use with the RPC server.
//
// The interface contract requires that all of these methods are safe for
// concurrent access.
type rpcserverPeer interface {
	// ToPeer returns the underlying peer instance.
	ToPeer() *peer.Peer

	// IsTxRelayDisabled returns whether or not the peer has disabled
	// transaction relay.
	IsTxRelayDisabled() bool

	// BanScore returns the current integer value that represents how close
	// the peer is to being banned.
	BanScore() uint32

	// FeeFilter returns the requested current minimum fee rate for which
	// transactions should be announced.
	FeeFilter() int64
}

// rpcserverConnManager represents a connection manager for use with the RPC
// server.
//
// The interface contract requires that all of these methods are safe for
// concurrent access.
type rpcserverConnManager interface {
	// Connect adds the provided address as a new outbound peer.  The
	// permanent flag indicates whether or not to make the peer persistent
	// and reconnect if the connection is lost.  Attempting to connect to an
	// already existing peer will return an error.
	Connect(addr string, permanent bool) error

	// RemoveByID removes the peer associated with the provided id from the
	// list of persistent peers.  Attempting to remove an id that does not
	// exist will return an error.
	RemoveByID(id int32) error

	// RemoveByAddr removes the peer associated with the provided address
	// from the list of persistent peers.  Attempting to remove an address
	// that does not exist will return an error.
	RemoveByAddr(addr string) error

	// DisconnectByID disconnects the peer associated with the provided id.
	// This applies to both inbound and outbound peers.  Attempting to
	// remove an id that does not exist will return an error.
	DisconnectByID(id int32) error

	// DisconnectByAddr disconnects the peer associated with the provided
	// address.  This applies to both inbound and outbound peers.
	// Attempting to remove an address that does not exist will return an
	// error.
	DisconnectByAddr(addr string) error

	// ConnectedCount returns the number of currently connected peers.
	ConnectedCount() int32

	// NetTotals returns the sum of all bytes received and sent across the
	// network for all peers.
	NetTotals() (uint64, uint64)

	// ConnectedPeers returns an array consisting of all connected peers.
	ConnectedPeers() []rpcserverPeer

	// PersistentPeers returns an array consisting of all the persistent
	// peers.
	PersistentPeers() []rpcserverPeer

	// BroadcastMessage sends the provided message to all currently
	// connected peers.
	BroadcastMessage(msg wire.Message)

	// AddRebroadcastInventory adds the provided inventory to the list of
	// inventories to be rebroadcast at random intervals until they show up
	// in a block.
	AddRebroadcastInventory(iv *wire.InvVect, data interface{})

	// RelayTransactions generates and relays inventory vectors for all of
	// the passed transactions to all connected peers.
	RelayTransactions(txns []*mempool.TxDesc)
}

// rpcserverSyncManager represents a sync manager for use with the RPC server.
//
// The interface contract requires that all of these methods are safe for
// concurrent access.
type rpcserverSyncManager interface {
	// IsCurrent returns whether or not the sync manager believes the chain
	// is current as compared to the rest of the network.
	IsCurrent() bool

	// SubmitBlock submits the provided block to the network after
	// processing it locally.
	SubmitBlock(block *util.Block, flags blockdag.BehaviorFlags) (bool, error)

	// Pause pauses the sync manager until the returned channel is closed.
	Pause() chan<- struct{}

	// SyncPeerID returns the ID of the peer that is currently the peer being
	// used to sync from or 0 if there is none.
	SyncPeerID() int32

	// GetBlueBlocksHeadersBetween returns the headers of the blocks after the first known
	// block in the provided locators until the provided stop hash or the
	// current tip is reached, up to a max of wire.MaxBlockHeadersPerMsg
	// hashes.
	GetBlueBlocksHeadersBetween(startHash, stopHash *daghash.Hash) []*wire.BlockHeader
}

// rpcserverConfig is a descriptor containing the RPC server configuration.
type rpcserverConfig struct {
	// Listeners defines a slice of listeners for which the RPC server will
	// take ownership of and accept connections.  Since the RPC server takes
	// ownership of these listeners, they will be closed when the RPC server
	// is stopped.
	Listeners []net.Listener

	// StartupTime is the unix timestamp for when the server that is hosting
	// the RPC server started.
	StartupTime int64

	// ConnMgr defines the connection manager for the RPC server to use.  It
	// provides the RPC server with a means to do things such as add,
	// remove, connect, disconnect, and query peers as well as other
	// connection-related data and tasks.
	ConnMgr rpcserverConnManager

	// SyncMgr defines the sync manager for the RPC server to use.
	SyncMgr rpcserverSyncManager

	// These fields allow the RPC server to interface with the local block
	// chain data and state.
	TimeSource blockdag.MedianTimeSource
	DAG        *blockdag.BlockDAG
	DAGParams  *dagconfig.Params
	DB         database.DB

	// TxMemPool defines the transaction memory pool to interact with.
	TxMemPool *mempool.TxPool

	// These fields allow the RPC server to interface with mining.
	//
	// Generator produces block templates and the CPUMiner solves them using
	// the CPU.  CPU mining is typically only useful for test purposes when
	// doing regression or simulation testing.
	Generator *mining.BlkTmplGenerator
	CPUMiner  *cpuminer.CPUMiner

	// These fields define any optional indexes the RPC server can make use
	// of to provide additional data when queried.
	TxIndex         *indexers.TxIndex
	AddrIndex       *indexers.AddrIndex
	AcceptanceIndex *indexers.AcceptanceIndex
	CfIndex         *indexers.CfIndex
}

// setupRPCListeners returns a slice of listeners that are configured for use
// with the RPC server depending on the configuration settings for listen
// addresses and TLS.
func setupRPCListeners() ([]net.Listener, error) {
	// Setup TLS if not disabled.
	listenFunc := net.Listen
	if !config.MainConfig().DisableTLS {
		// Generate the TLS cert and key file if both don't already
		// exist.
		if !fs.FileExists(config.MainConfig().RPCKey) && !fs.FileExists(config.MainConfig().RPCCert) {
			err := serverutils.GenCertPair(config.MainConfig().RPCCert, config.MainConfig().RPCKey)
			if err != nil {
				return nil, err
			}
		}
		keypair, err := tls.LoadX509KeyPair(config.MainConfig().RPCCert, config.MainConfig().RPCKey)
		if err != nil {
			return nil, err
		}

		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{keypair},
			MinVersion:   tls.VersionTLS12,
		}

		// Change the standard net.Listen function to the tls one.
		listenFunc = func(net string, laddr string) (net.Listener, error) {
			return tls.Listen(net, laddr, &tlsConfig)
		}
	}

	netAddrs, err := p2p.ParseListeners(config.MainConfig().RPCListeners)
	if err != nil {
		return nil, err
	}

	listeners := make([]net.Listener, 0, len(netAddrs))
	for _, addr := range netAddrs {
		listener, err := listenFunc(addr.Network(), addr.String())
		if err != nil {
			log.Warnf("Can't listen on %s: %s", addr, err)
			continue
		}
		listeners = append(listeners, listener)
	}

	return listeners, nil
}

// NewRPCServer returns a new instance of the rpcServer struct.
func NewRPCServer(
	startupTime int64,
	p2pServer *p2p.Server,
	db database.DB,
	blockTemplateGenerator *mining.BlkTmplGenerator,
	cpuminer *cpuminer.CPUMiner,

) (*Server, error) {
	// Setup listeners for the configured RPC listen addresses and
	// TLS settings.
	rpcListeners, err := setupRPCListeners()
	if err != nil {
		return nil, err
	}
	if len(rpcListeners) == 0 {
		return nil, errors.New("RPCS: No valid listen address")
	}
	cfg := &rpcserverConfig{
		Listeners:       rpcListeners,
		StartupTime:     startupTime,
		ConnMgr:         &rpcConnManager{p2pServer},
		SyncMgr:         &rpcSyncMgr{p2pServer, p2pServer.SyncManager},
		TimeSource:      p2pServer.TimeSource,
		DAGParams:       p2pServer.DAGParams,
		DB:              db,
		TxMemPool:       p2pServer.TxMemPool,
		Generator:       blockTemplateGenerator,
		CPUMiner:        cpuminer,
		TxIndex:         p2pServer.TxIndex,
		AddrIndex:       p2pServer.AddrIndex,
		AcceptanceIndex: p2pServer.AcceptanceIndex,
		CfIndex:         p2pServer.CfIndex,
		DAG:             p2pServer.DAG,
	}
	rpc := Server{
		cfg:                    *cfg,
		statusLines:            make(map[int]string),
		gbtWorkState:           newGbtWorkState(cfg.TimeSource),
		helpCacher:             newHelpCacher(),
		requestProcessShutdown: make(chan struct{}),
		quit:                   make(chan int),
	}
	if config.MainConfig().RPCUser != "" && config.MainConfig().RPCPass != "" {
		login := config.MainConfig().RPCUser + ":" + config.MainConfig().RPCPass
		auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(login))
		rpc.authsha = sha256.Sum256([]byte(auth))
	}
	if config.MainConfig().RPCLimitUser != "" && config.MainConfig().RPCLimitPass != "" {
		login := config.MainConfig().RPCLimitUser + ":" + config.MainConfig().RPCLimitPass
		auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(login))
		rpc.limitauthsha = sha256.Sum256([]byte(auth))
	}
	rpc.ntfnMgr = newWsNotificationManager(&rpc)
	rpc.cfg.DAG.Subscribe(rpc.handleBlockDAGNotification)

	return &rpc, nil
}

// Callback for notifications from blockdag.  It notifies clients that are
// long polling for changes or subscribed to websockets notifications.
func (s *Server) handleBlockDAGNotification(notification *blockdag.Notification) {
	switch notification.Type {
	case blockdag.NTBlockAdded:
		data, ok := notification.Data.(*blockdag.BlockAddedNotificationData)
		if !ok {
			log.Warnf("Block added notification data is of wrong type.")
			break
		}
		block := data.Block

		tipHashes := s.cfg.DAG.TipHashes()

		// Allow any clients performing long polling via the
		// getBlockTemplate RPC to be notified when the new block causes
		// their old block template to become stale.
		s.gbtWorkState.NotifyBlockAdded(tipHashes)

		// Notify registered websocket clients of incoming block.
		s.ntfnMgr.NotifyBlockAdded(block)
	case blockdag.NTChainChanged:
		data, ok := notification.Data.(*blockdag.ChainChangedNotificationData)
		if !ok {
			log.Warnf("Chain changed notification data is of wrong type.")
			break
		}

		// Notify registered websocket clients of chain changes.
		s.ntfnMgr.NotifyChainChanged(data.RemovedChainBlockHashes,
			data.AddedChainBlockHashes)
	}
}

func init() {
	rpcHandlers = rpcHandlersBeforeInit
	rand.Seed(time.Now().UnixNano())
}
