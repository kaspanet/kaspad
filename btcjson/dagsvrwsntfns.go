// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// NOTE: This file is intended to house the RPC websocket notifications that are
// supported by a dag server.

package btcjson

const (
	// BlockAddedNtfnMethod is the legacy, deprecated method used for
	// notifications from the dag server that a block has been connected.
	//
	// NOTE: Deprecated. Use FilteredBlockAddedNtfnMethod instead.
	BlockAddedNtfnMethod = "blockAdded"

	// FilteredBlockAddedNtfnMethod is the new method used for
	// notifications from the dag server that a block has been connected.
	FilteredBlockAddedNtfnMethod = "filteredBlockAdded"

	// RecvTxNtfnMethod is the legacy, deprecated method used for
	// notifications from the dag server that a transaction which pays to
	// a registered address has been processed.
	//
	// NOTE: Deprecated. Use RelevantTxAcceptedNtfnMethod and
	// FilteredBlockAddedNtfnMethod instead.
	RecvTxNtfnMethod = "recvTx"

	// RedeemingTxNtfnMethod is the legacy, deprecated method used for
	// notifications from the dag server that a transaction which spends a
	// registered outpoint has been processed.
	//
	// NOTE: Deprecated. Use RelevantTxAcceptedNtfnMethod and
	// FilteredBlockAddedNtfnMethod instead.
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

// BlockAddedNtfn defines the blockAdded JSON-RPC notification.
//
// NOTE: Deprecated. Use FilteredBlockAddedNtfn instead.
type BlockAddedNtfn struct {
	Hash   string
	Height uint64
	Time   int64
}

// NewBlockAddedNtfn returns a new instance which can be used to issue a
// blockAdded JSON-RPC notification.
//
// NOTE: Deprecated. Use NewFilteredBlockAddedNtfn instead.
func NewBlockAddedNtfn(hash string, height uint64, time int64) *BlockAddedNtfn {
	return &BlockAddedNtfn{
		Hash:   hash,
		Height: height,
		Time:   time,
	}
}

// FilteredBlockAddedNtfn defines the filteredBlockAdded JSON-RPC
// notification.
type FilteredBlockAddedNtfn struct {
	Height        uint64
	Header        string
	SubscribedTxs []string
}

// NewFilteredBlockAddedNtfn returns a new instance which can be used to
// issue a filteredBlockAdded JSON-RPC notification.
func NewFilteredBlockAddedNtfn(height uint64, header string, subscribedTxs []string) *FilteredBlockAddedNtfn {
	return &FilteredBlockAddedNtfn{
		Height:        height,
		Header:        header,
		SubscribedTxs: subscribedTxs,
	}
}

// BlockDetails describes details of a tx in a block.
type BlockDetails struct {
	Height uint64 `json:"height"`
	Hash   string `json:"hash"`
	Index  int    `json:"index"`
	Time   int64  `json:"time"`
}

// RecvTxNtfn defines the recvTx JSON-RPC notification.
//
// NOTE: Deprecated. Use RelevantTxAcceptedNtfn and FilteredBlockAddedNtfn
// instead.
type RecvTxNtfn struct {
	HexTx string
	Block *BlockDetails
}

// NewRecvTxNtfn returns a new instance which can be used to issue a recvTx
// JSON-RPC notification.
//
// NOTE: Deprecated. Use NewRelevantTxAcceptedNtfn and
// NewFilteredBlockAddedNtfn instead.
func NewRecvTxNtfn(hexTx string, block *BlockDetails) *RecvTxNtfn {
	return &RecvTxNtfn{
		HexTx: hexTx,
		Block: block,
	}
}

// RedeemingTxNtfn defines the redeemingTx JSON-RPC notification.
//
// NOTE: Deprecated. Use RelevantTxAcceptedNtfn and FilteredBlockAddedNtfn
// instead.
type RedeemingTxNtfn struct {
	HexTx string
	Block *BlockDetails
}

// NewRedeemingTxNtfn returns a new instance which can be used to issue a
// redeemingTx JSON-RPC notification.
//
// NOTE: Deprecated. Use NewRelevantTxAcceptedNtfn and
// NewFilteredBlockAddedNtfn instead.
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
	Height uint64
	Time   int64
}

// NewRescanFinishedNtfn returns a new instance which can be used to issue a
// rescanFinished JSON-RPC notification.
//
// NOTE: Deprecated. Not used with rescanblocks command.
func NewRescanFinishedNtfn(hash string, height uint64, time int64) *RescanFinishedNtfn {
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
	Height uint64
	Time   int64
}

// NewRescanProgressNtfn returns a new instance which can be used to issue a
// rescanProgress JSON-RPC notification.
//
// NOTE: Deprecated. Not used with rescanblocks command.
func NewRescanProgressNtfn(hash string, height uint64, time int64) *RescanProgressNtfn {
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

	MustRegisterCmd(BlockAddedNtfnMethod, (*BlockAddedNtfn)(nil), flags)
	MustRegisterCmd(FilteredBlockAddedNtfnMethod, (*FilteredBlockAddedNtfn)(nil), flags)
	MustRegisterCmd(RecvTxNtfnMethod, (*RecvTxNtfn)(nil), flags)
	MustRegisterCmd(RedeemingTxNtfnMethod, (*RedeemingTxNtfn)(nil), flags)
	MustRegisterCmd(RescanFinishedNtfnMethod, (*RescanFinishedNtfn)(nil), flags)
	MustRegisterCmd(RescanProgressNtfnMethod, (*RescanProgressNtfn)(nil), flags)
	MustRegisterCmd(TxAcceptedNtfnMethod, (*TxAcceptedNtfn)(nil), flags)
	MustRegisterCmd(TxAcceptedVerboseNtfnMethod, (*TxAcceptedVerboseNtfn)(nil), flags)
	MustRegisterCmd(RelevantTxAcceptedNtfnMethod, (*RelevantTxAcceptedNtfn)(nil), flags)
}
