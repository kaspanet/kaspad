package txindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// TXAcceptanceChange is the set of changes made to the TX index after
// a successful update
type TXAcceptanceChange struct {
	Added   map[externalapi.DomainTransactionID]*TxData
	Removed map[externalapi.DomainTransactionID]*TxData
}

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
