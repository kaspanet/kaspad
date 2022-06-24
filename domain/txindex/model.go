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

// ConvertDomainHashToString converts the given DomainHash to a string
func ConvertDomainHashToString(blockHash *externalapi.DomainHash) string {
	return hex.EncodeToString(blockHash.ByteSlice())
}

// ConvertStringDomainHashToDomainHash converts the given string to a domainHash
func ConvertStringToDomainHash(stringDomainHash string) (*externalapi.DomainHash, error) {
	return externalapi.NewDomainHashFromString(stringDomainHash)
}

// ConvertDomainHashToString converts the given DomainHash to a string
func ConvertTXIDToString(txID *externalapi.DomainTransactionID) string {
	return hex.EncodeToString(txID.ByteSlice())
}

// ConvertStringDomainHashToDomainHash converts the given string to a domainHash
func ConvertStringTXID(stringDomainTransactionID string) (*externalapi.DomainTransactionID, error) {
	return externalapi.NewDomainTransactionIDFromString(stringDomainTransactionID)
}

