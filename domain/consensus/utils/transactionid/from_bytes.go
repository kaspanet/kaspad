package transactionid

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// FromBytes creates a DomainTransactionID from the given byte slice
func FromBytes(transactionIDBytes []byte) (*externalapi.DomainTransactionID, error) {
	if len(transactionIDBytes) != externalapi.DomainHashSize {
		return nil, errors.Errorf("invalid hash size. Want: %d, got: %d",
			externalapi.DomainHashSize, len(transactionIDBytes))
	}
	var domainTransactionID externalapi.DomainTransactionID
	copy(domainTransactionID[:], transactionIDBytes)
	return &domainTransactionID, nil
}
