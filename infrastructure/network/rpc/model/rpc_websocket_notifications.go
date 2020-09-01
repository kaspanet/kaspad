// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// NOTE: This file is intended to house the RPC websocket notifications that are
// supported by a kaspa rpc server.

package model

const (
	// FilteredBlockAddedNtfnMethod is the new method used for
	// notifications from the kaspa rpc server that a block has been connected.
	FilteredBlockAddedNtfnMethod = "filteredBlockAdded"

	// TxAcceptedNtfnMethod is the method used for notifications from the
	// kaspa rpc server that a transaction has been accepted into the mempool.
	TxAcceptedNtfnMethod = "txAccepted"

	// TxAcceptedVerboseNtfnMethod is the method used for notifications from
	// the kaspa rpc server that a transaction has been accepted into the
	// mempool. This differs from TxAcceptedNtfnMethod in that it provides
	// more details in the notification.
	TxAcceptedVerboseNtfnMethod = "txAcceptedVerbose"

	// RelevantTxAcceptedNtfnMethod is the new method used for notifications
	// from the kaspa rpc server that inform a client that a transaction that
	// matches the loaded filter was accepted by the mempool.
	RelevantTxAcceptedNtfnMethod = "relevantTxAccepted"

	// ChainChangedNtfnMethod is the new method used for notifications
	// from the kaspa rpc server that inform a client that the selected chain
	// has changed.
	ChainChangedNtfnMethod = "chainChanged"

	// FinalityConflictNtfnMethod is the new method used for notifications
	// from the kaspa rpc server that inform a client that a finality conflict
	// has occured.
	FinalityConflictNtfnMethod = "finalityConflict"

	// FinalityConflictResolvedNtfnMethod is the new method used for notifications
	// from the kaspa rpc server that inform a client that a finality conflict
	// has been resolved.
	FinalityConflictResolvedNtfnMethod = "finalityConflictResolved"
)

// FilteredBlockAddedNtfn defines the filteredBlockAdded JSON-RPC
// notification.
type FilteredBlockAddedNtfn struct {
	BlueScore     uint64
	Header        string
	SubscribedTxs []string
}

// NewFilteredBlockAddedNtfn returns a new instance which can be used to
// issue a filteredBlockAdded JSON-RPC notification.
func NewFilteredBlockAddedNtfn(blueScore uint64, header string, subscribedTxs []string) *FilteredBlockAddedNtfn {
	return &FilteredBlockAddedNtfn{
		BlueScore:     blueScore,
		Header:        header,
		SubscribedTxs: subscribedTxs,
	}
}

// ChainChangedNtfn defines the chainChanged JSON-RPC
// notification.
type ChainChangedNtfn struct {
	ChainChangedRawParam ChainChangedRawParam
}

// ChainChangedRawParam is the first parameter
// of ChainChangedNtfn which contains all the
// remove chain block hashes and the added
// chain blocks.
type ChainChangedRawParam struct {
	RemovedChainBlockHashes []string     `json:"removedChainBlockHashes"`
	AddedChainBlocks        []ChainBlock `json:"addedChainBlocks"`
}

// NewChainChangedNtfn returns a new instance which can be used to
// issue a chainChanged JSON-RPC notification.
func NewChainChangedNtfn(removedChainBlockHashes []string,
	addedChainBlocks []ChainBlock) *ChainChangedNtfn {
	return &ChainChangedNtfn{ChainChangedRawParam: ChainChangedRawParam{
		RemovedChainBlockHashes: removedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
	}}
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

// FinalityConflictNtfn  defines the parameters to the finalityConflict
// JSON-RPC notification.
type FinalityConflictNtfn struct {
	ViolatingBlockHash string `json:"violatingBlockHash"`
	ConflictTime       int64  `json:"conflictTime"`
}

// NewFinalityConflictNtfn returns a new instance which can be used to issue a
// finalityConflict JSON-RPC notification.
func NewFinalityConflictNtfn(violatingBlockHash string, conflictTime int64) *FinalityConflictNtfn {
	return &FinalityConflictNtfn{
		ViolatingBlockHash: violatingBlockHash,
		ConflictTime:       conflictTime,
	}
}

// FinalityConflictResolvedNtfn defines the parameters to the
// finalityConflictResolved JSON-RPC notification.
type FinalityConflictResolvedNtfn struct {
	FinalityBlockHash string
	ResolutionTime    int64 `json:"resolutionTime"`
}

// NewFinalityConflictResolvedNtfn returns a new instance which can be used to issue a
// finalityConflictResolved JSON-RPC notification.
func NewFinalityConflictResolvedNtfn(finalityBlockHash string, resolutionTime int64) *FinalityConflictResolvedNtfn {
	return &FinalityConflictResolvedNtfn{
		FinalityBlockHash: finalityBlockHash,
		ResolutionTime:    resolutionTime,
	}
}

func init() {
	// The commands in this file are only usable by websockets and are
	// notifications.
	flags := UFWebsocketOnly | UFNotification

	MustRegisterCommand(FilteredBlockAddedNtfnMethod, (*FilteredBlockAddedNtfn)(nil), flags)
	MustRegisterCommand(TxAcceptedNtfnMethod, (*TxAcceptedNtfn)(nil), flags)
	MustRegisterCommand(TxAcceptedVerboseNtfnMethod, (*TxAcceptedVerboseNtfn)(nil), flags)
	MustRegisterCommand(RelevantTxAcceptedNtfnMethod, (*RelevantTxAcceptedNtfn)(nil), flags)
	MustRegisterCommand(ChainChangedNtfnMethod, (*ChainChangedNtfn)(nil), flags)
	MustRegisterCommand(FinalityConflictNtfnMethod, (*FinalityConflictNtfn)(nil), flags)
	MustRegisterCommand(FinalityConflictResolvedNtfnMethod, (*FinalityConflictResolvedNtfn)(nil), flags)
}
