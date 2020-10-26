package stringers

import (
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// Outpoint stringifies an outpoint.
func Outpoint(outpoint *externalapi.DomainOutpoint) string {
	return fmt.Sprintf("%s:%d", TransactionID(&outpoint.ID), outpoint.Index)
}

// TransactionID stringifies a transaction ID.
func TransactionID(id *externalapi.DomainTransactionID) string {
	return hex.EncodeToString(id[:])
}
