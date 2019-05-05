// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// NOTE: This file is intended to house the RPC websocket notifications that are
// supported by a dag server.

package btcjson

const (
	// FilteredBlockAddedNtfnMethod is the new method used for
	// notifications from the dag server that a block has been connected.
	FilteredBlockAddedNtfnMethod = "filteredBlockAdded"

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

	MustRegisterCmd(FilteredBlockAddedNtfnMethod, (*FilteredBlockAddedNtfn)(nil), flags)
	MustRegisterCmd(TxAcceptedNtfnMethod, (*TxAcceptedNtfn)(nil), flags)
	MustRegisterCmd(TxAcceptedVerboseNtfnMethod, (*TxAcceptedVerboseNtfn)(nil), flags)
	MustRegisterCmd(RelevantTxAcceptedNtfnMethod, (*RelevantTxAcceptedNtfn)(nil), flags)
}
