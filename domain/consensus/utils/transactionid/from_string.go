package transactionid

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
)

// NewDomainTransactionIDFromString creates a new DomainTransactionID from the given string
func NewDomainTransactionIDFromString(str string) (*externalapi.DomainTransactionID, error) {
	hash, err := hashes.FromString(str)
	return (*externalapi.DomainTransactionID)(hash), err
}
