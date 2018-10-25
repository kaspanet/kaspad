// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// NOTE: This file is intended to house the RPC websocket notifications that are
// supported by a dag server.

package btcjson

const (
	// BlockConnectedNtfnMethod is the legacy, deprecated method used for
	// notifications from the dag server that a block has been connected.
	//
	// NOTE: Deprecated. Use FilteredBlockConnectedNtfnMethod instead.
	BlockConnectedNtfnMethod = "blockConnected"

	// BlockDisconnectedNtfnMethod is the legacy, deprecated method used for
	// notifications from the dag server that a block has been
	// disconnected.
	//
	// NOTE: Deprecated. Use FilteredBlockDisconnectedNtfnMethod instead.
	BlockDisconnectedNtfnMethod = "blockDisconnected"

	// FilteredBlockConnectedNtfnMethod is the new method used for
	// notifications from the dag server that a block has been connected.
	FilteredBlockConnectedNtfnMethod = "filteredBlockConnected"

	// FilteredBlockDisconnectedNtfnMethod is the new method used for
	// notifications from the dag server that a block has been
	// disconnected.
	FilteredBlockDisconnectedNtfnMethod = "filteredBlockDisconnected"

	// RecvTxNtfnMethod is the legacy, deprecated method used for
	// notifications from the dag server that a transaction which pays to
	// a registered address has been processed.
	//
	// NOTE: Deprecated. Use RelevantTxAcceptedNtfnMethod and
	// FilteredBlockConnectedNtfnMethod instead.
	RecvTxNtfnMethod = "recvTx"

	// RedeemingTxNtfnMethod is the legacy, deprecated method used for
	// notifications from the dag server that a transaction which spends a
	// registered outpoint has been processed.
	//
	// NOTE: Deprecated. Use RelevantTxAcceptedNtfnMethod and
	// FilteredBlockConnectedNtfnMethod instead.
	RedeemingTxNtfnMethod = "redeemingTx"

	// RescanFinishedNtfnMethod is the legacy, deprecated method used for
	// notifications from the dag server that a legacy, deprecated rescan
	// operation has finished.
	//
	// NOTE: Deprecated. Not used with rescanblocks command.
	RescanFinishedNtfnMethod = "rescanFinished"

	// RescanProgressNtfnMethod is the legacy, deprecated method used for
	// notifications from the dag server that a legacy, deprecated rescan
	// operation this is underway has made progress.
	//
	// NOTE: Deprecated. Not used with rescanblocks command.
	RescanProgressNtfnMethod = "rescanProgress"

	// TxAcceptedNtfnMethod is the method used for notifications from the
	// dag server that a transaction has been accepted into the mempool.
	TxAcceptedNtfnMethod = "txAccepted"

	// TxAcceptedVerboseNtfnMethod is the method used for notifications from
	// the dag server that a transaction has been accepted into the
	// mempool.  This differs from TxAcceptedNtfnMethod in that it provides
	// more details in the notification.
	TxAcceptedVerboseNtfnMethod = "txAcceptedVerbose"

	// RelevantTxAcceptedNtfnMethod is the new method used for notifications
	// from the dag server that inform a client that a transaction that
	// matches the loaded filter was accepted by the mempool.
	RelevantTxAcceptedNtfnMethod = "relevantTxAccepted"
)

// BlockConnectedNtfn defines the blockConnected JSON-RPC notification.
//
// NOTE: Deprecated. Use FilteredBlockConnectedNtfn instead.
type BlockConnectedNtfn struct {
	Hash   string
	Height int32
	Time   int64
}

// NewBlockConnectedNtfn returns a new instance which can be used to issue a
// blockConnected JSON-RPC notification.
//
// NOTE: Deprecated. Use NewFilteredBlockConnectedNtfn instead.
func NewBlockConnectedNtfn(hash string, height int32, time int64) *BlockConnectedNtfn {
	return &BlockConnectedNtfn{
		Hash:   hash,
		Height: height,
		Time:   time,
	}
}

// BlockDisconnectedNtfn defines the blockDisconnected JSON-RPC notification.
//
// NOTE: Deprecated. Use FilteredBlockDisconnectedNtfn instead.
type BlockDisconnectedNtfn struct {
	Hash   string
	Height int32
	Time   int64
}

// NewBlockDisconnectedNtfn returns a new instance which can be used to issue a
// blockDisconnected JSON-RPC notification.
//
// NOTE: Deprecated. Use NewFilteredBlockDisconnectedNtfn instead.
func NewBlockDisconnectedNtfn(hash string, height int32, time int64) *BlockDisconnectedNtfn {
	return &BlockDisconnectedNtfn{
		Hash:   hash,
		Height: height,
		Time:   time,
	}
}

// FilteredBlockConnectedNtfn defines the filteredBlockConnected JSON-RPC
// notification.
type FilteredBlockConnectedNtfn struct {
	Height        int32
	Header        string
	SubscribedTxs []string
}

// NewFilteredBlockConnectedNtfn returns a new instance which can be used to
// issue a filteredBlockConnected JSON-RPC notification.
func NewFilteredBlockConnectedNtfn(height int32, header string, subscribedTxs []string) *FilteredBlockConnectedNtfn {
	return &FilteredBlockConnectedNtfn{
		Height:        height,
		Header:        header,
		SubscribedTxs: subscribedTxs,
	}
}

// FilteredBlockDisconnectedNtfn defines the filteredBlockDisconnected JSON-RPC
// notification.
type FilteredBlockDisconnectedNtfn struct {
	Height int32
	Header string
}

// NewFilteredBlockDisconnectedNtfn returns a new instance which can be used to
// issue a filteredBlockDisconnected JSON-RPC notification.
func NewFilteredBlockDisconnectedNtfn(height int32, header string) *FilteredBlockDisconnectedNtfn {
	return &FilteredBlockDisconnectedNtfn{
		Height: height,
		Header: header,
	}
}

// BlockDetails describes details of a tx in a block.
type BlockDetails struct {
	Height int32  `json:"height"`
	Hash   string `json:"hash"`
	Index  int    `json:"index"`
	Time   int64  `json:"time"`
}

// RecvTxNtfn defines the recvTx JSON-RPC notification.
//
// NOTE: Deprecated. Use RelevantTxAcceptedNtfn and FilteredBlockConnectedNtfn
// instead.
type RecvTxNtfn struct {
	HexTx string
	Block *BlockDetails
}

// NewRecvTxNtfn returns a new instance which can be used to issue a recvTx
// JSON-RPC notification.
//
// NOTE: Deprecated. Use NewRelevantTxAcceptedNtfn and
// NewFilteredBlockConnectedNtfn instead.
func NewRecvTxNtfn(hexTx string, block *BlockDetails) *RecvTxNtfn {
	return &RecvTxNtfn{
		HexTx: hexTx,
		Block: block,
	}
}

// RedeemingTxNtfn defines the redeemingTx JSON-RPC notification.
//
// NOTE: Deprecated. Use RelevantTxAcceptedNtfn and FilteredBlockConnectedNtfn
// instead.
type RedeemingTxNtfn struct {
	HexTx string
	Block *BlockDetails
}

// NewRedeemingTxNtfn returns a new instance which can be used to issue a
// redeemingTx JSON-RPC notification.
//
// NOTE: Deprecated. Use NewRelevantTxAcceptedNtfn and
// NewFilteredBlockConnectedNtfn instead.
func NewRedeemingTxNtfn(hexTx string, block *BlockDetails) *RedeemingTxNtfn {
	return &RedeemingTxNtfn{
		HexTx: hexTx,
		Block: block,
	}
}

// RescanFinishedNtfn defines the rescanFinished JSON-RPC notification.
//
// NOTE: Deprecated. Not used with rescanblocks command.
type RescanFinishedNtfn struct {
	Hash   string
	Height int32
	Time   int64
}

// NewRescanFinishedNtfn returns a new instance which can be used to issue a
// rescanFinished JSON-RPC notification.
//
// NOTE: Deprecated. Not used with rescanblocks command.
func NewRescanFinishedNtfn(hash string, height int32, time int64) *RescanFinishedNtfn {
	return &RescanFinishedNtfn{
		Hash:   hash,
		Height: height,
		Time:   time,
	}
}

// RescanProgressNtfn defines the rescanProgress JSON-RPC notification.
//
// NOTE: Deprecated. Not used with rescanblocks command.
type RescanProgressNtfn struct {
	Hash   string
	Height int32
	Time   int64
}

// NewRescanProgressNtfn returns a new instance which can be used to issue a
// rescanProgress JSON-RPC notification.
//
// NOTE: Deprecated. Not used with rescanblocks command.
func NewRescanProgressNtfn(hash string, height int32, time int64) *RescanProgressNtfn {
	return &RescanProgressNtfn{
		Hash:   hash,
		Height: height,
		Time:   time,
	}
}

// TxAcceptedNtfn defines the txAccepted JSON-RPC notification.
type TxAcceptedNtfn struct {
	TxID   string
	Amount float64
}

// NewTxAcceptedNtfn returns a new instance which can be used to issue a
// txAccepted JSON-RPC notification.
func NewTxAcceptedNtfn(txHash string, amount float64) *TxAcceptedNtfn {
	return &TxAcceptedNtfn{
		TxID:   txHash,
		Amount: amount,
	}
}

// TxAcceptedVerboseNtfn defines the txAcceptedVerbose JSON-RPC notification.
type TxAcceptedVerboseNtfn struct {
	RawTx TxRawResult
}

// NewTxAcceptedVerboseNtfn returns a new instance which can be used to issue a
// txAcceptedVerbose JSON-RPC notification.
func NewTxAcceptedVerboseNtfn(rawTx TxRawResult) *TxAcceptedVerboseNtfn {
	return &TxAcceptedVerboseNtfn{
		RawTx: rawTx,
	}
}

// RelevantTxAcceptedNtfn defines the parameters to the relevantTxAccepted
// JSON-RPC notification.
type RelevantTxAcceptedNtfn struct {
	Transaction string `json:"transaction"`
}

// NewRelevantTxAcceptedNtfn returns a new instance which can be used to issue a
// relevantxaccepted JSON-RPC notification.
func NewRelevantTxAcceptedNtfn(txHex string) *RelevantTxAcceptedNtfn {
	return &RelevantTxAcceptedNtfn{Transaction: txHex}
}

func init() {
	// The commands in this file are only usable by websockets and are
	// notifications.
	flags := UFWebsocketOnly | UFNotification

	MustRegisterCmd(BlockConnectedNtfnMethod, (*BlockConnectedNtfn)(nil), flags)
	MustRegisterCmd(BlockDisconnectedNtfnMethod, (*BlockDisconnectedNtfn)(nil), flags)
	MustRegisterCmd(FilteredBlockConnectedNtfnMethod, (*FilteredBlockConnectedNtfn)(nil), flags)
	MustRegisterCmd(FilteredBlockDisconnectedNtfnMethod, (*FilteredBlockDisconnectedNtfn)(nil), flags)
	MustRegisterCmd(RecvTxNtfnMethod, (*RecvTxNtfn)(nil), flags)
	MustRegisterCmd(RedeemingTxNtfnMethod, (*RedeemingTxNtfn)(nil), flags)
	MustRegisterCmd(RescanFinishedNtfnMethod, (*RescanFinishedNtfn)(nil), flags)
	MustRegisterCmd(RescanProgressNtfnMethod, (*RescanProgressNtfn)(nil), flags)
	MustRegisterCmd(TxAcceptedNtfnMethod, (*TxAcceptedNtfn)(nil), flags)
	MustRegisterCmd(TxAcceptedVerboseNtfnMethod, (*TxAcceptedVerboseNtfn)(nil), flags)
	MustRegisterCmd(RelevantTxAcceptedNtfnMethod, (*RelevantTxAcceptedNtfn)(nil), flags)
}
