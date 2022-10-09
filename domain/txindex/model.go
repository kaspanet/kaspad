package txindex

import (
	"encoding/hex"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// TXAcceptanceChange is the set of changes made to the TX index after
// a successful update
type TXAcceptanceChange struct {
	Added   map[externalapi.DomainTransactionID]*externalapi.DomainHash
	Removed map[externalapi.DomainTransactionID]*externalapi.DomainHash
}

//TxIDsToBlockHashes is a map of TxIDs to corrospnding blockHashes
type TxIDsToBlockHashes map[*externalapi.DomainTransactionID]*externalapi.DomainHash

//TxIDsToBlocks is a map of TxIDs to corrospnding blocks
type TxIDsToBlocks map[*externalapi.DomainTransactionID]*externalapi.DomainBlock

// ConvertDomainHashToString converts the given DomainHash to a string
func ConvertDomainHashToString(blockHash *externalapi.DomainHash) string {
	return hex.EncodeToString(blockHash.ByteSlice())
}

// ConvertStringToDomainHash converts the given string to a domainHash
func ConvertStringToDomainHash(stringDomainHash string) (*externalapi.DomainHash, error) {
	return externalapi.NewDomainHashFromString(stringDomainHash)
}

// ConvertTXIDToString converts the given DomainHash to a string
func ConvertTXIDToString(txID *externalapi.DomainTransactionID) string {
	return hex.EncodeToString(txID.ByteSlice())
}

// ConvertStringTXID converts the given string to a domainHash
func ConvertStringTXID(stringDomainTransactionID string) (*externalapi.DomainTransactionID, error) {
	return externalapi.NewDomainTransactionIDFromString(stringDomainTransactionID)
}
