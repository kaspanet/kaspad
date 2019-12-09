// Copyright (c) 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package jsonrpc

// Standard JSON-RPC 2.0 errors.
var (
	ErrRPCInvalidRequest = &RPCError{
		Code:    -32600,
		Message: "Invalid request",
	}
	ErrRPCMethodNotFound = &RPCError{
		Code:    -32601,
		Message: "Method not found",
	}
	ErrRPCInvalidParams = &RPCError{
		Code:    -32602,
		Message: "Invalid parameters",
	}
	ErrRPCInternal = &RPCError{
		Code:    -32603,
		Message: "Internal error",
	}
	ErrRPCParse = &RPCError{
		Code:    -32700,
		Message: "Parse error",
	}
)

// General application defined JSON errors.
const (
	ErrRPCMisc                RPCErrorCode = -1
	ErrRPCForbiddenBySafeMode RPCErrorCode = -2
	ErrRPCType                RPCErrorCode = -3
	ErrRPCInvalidAddressOrKey RPCErrorCode = -5
	ErrRPCOutOfMemory         RPCErrorCode = -7
	ErrRPCInvalidParameter    RPCErrorCode = -8
	ErrRPCDatabase            RPCErrorCode = -20
	ErrRPCDeserialization     RPCErrorCode = -22
	ErrRPCVerify              RPCErrorCode = -25
)

// Peer-to-peer client errors.
const (
	ErrRPCClientNotConnected      RPCErrorCode = -9
	ErrRPCClientInInitialDownload RPCErrorCode = -10
	ErrRPCClientNodeNotAdded      RPCErrorCode = -24
)

// Specific Errors related to commands. These are the ones a user of the RPC
// server are most likely to see. Generally, the codes should match one of the
// more general errors above.
const (
	ErrRPCBlockNotFound      RPCErrorCode = -5
	ErrRPCBlockCount         RPCErrorCode = -5
	ErrRPCSelectedTipHash    RPCErrorCode = -5
	ErrRPCDifficulty         RPCErrorCode = -5
	ErrRPCOutOfRange         RPCErrorCode = -1
	ErrRPCNoTxInfo           RPCErrorCode = -5
	ErrRPCNoAcceptanceIndex  RPCErrorCode = -5
	ErrRPCNoCFIndex          RPCErrorCode = -5
	ErrRPCNoNewestBlockInfo  RPCErrorCode = -5
	ErrRPCInvalidTxVout      RPCErrorCode = -5
	ErrRPCSubnetworkNotFound RPCErrorCode = -5
	ErrRPCRawTxString        RPCErrorCode = -32602
	ErrRPCDecodeHexString    RPCErrorCode = -22
	ErrRPCOrphanBlock        RPCErrorCode = -6
	ErrRPCBlockInvalid       RPCErrorCode = -5
)

// Errors that are specific to kaspad.
const (
	ErrRPCUnimplemented RPCErrorCode = -1
)
