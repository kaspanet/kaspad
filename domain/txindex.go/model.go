package txindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// TXChanges is the set of changes made to the UTXO index after
// a successful update
type TXAcceptanceChange struct {
	Added   map[externalapi.DomainTransactionID]*externalapi.DomainHash
	Removed map[externalapi.DomainTransactionID]*externalapi.DomainHash
}