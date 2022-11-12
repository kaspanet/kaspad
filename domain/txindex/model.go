package txindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// TXsChanges is the set of changes made to the TX index after
// a successful update
type TXsChanges struct {
	Added   TxChange
	Removed TxChange
}

type AddrsChanges struct {
	AddedSent   AddrsChange
	RemovedSent AddrsChange
	AddedReceived   AddrsChange
	RemovedReceived AddrsChange
}

type TxChange map[externalapi.DomainTransactionID]*TxData
type AddrsChange map[ScriptPublicKeyString][]*externalapi.DomainTransactionID

type VirtualBlueScore uint64

type ScriptPublicKeyString string
//TxData holds tx data stored in the TXIndex database

type TxData struct {
	IncludingBlockHash *externalapi.DomainHash
	AcceptingBlockHash *externalapi.DomainHash
	IncludingIndex     uint32
}

//TxIDsToTxIndexData is a map of TxIDs to corrospnding TxIndexData
type TxIDsToTxIndexData map[externalapi.DomainTransactionID]*TxData

//TxIDsToBlockHashes is a map of TxIDs to corrospnding blockHashes
type TxIDsToBlockHashes map[externalapi.DomainTransactionID]*externalapi.DomainHash

//TxIDsToBlocks is a map of TxIDs to corrospnding blocks
type TxIDsToBlocks map[externalapi.DomainTransactionID]*externalapi.DomainBlock

//TxIDsToConfirmations is a map of TxIDs to corrospnding Confirmations
type TxIDsToConfirmations map[externalapi.DomainTransactionID]int64

//TxIDsToBlueScores is a map of TxIDs to corrospnding Confirmations
type TxIDsToBlueScores map[externalapi.DomainTransactionID]uint64
