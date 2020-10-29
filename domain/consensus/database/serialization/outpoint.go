package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func DomainOutpointToDbOutpoint(domainOutpoint *externalapi.DomainOutpoint) *DbOutpoint {
	return &DbOutpoint{
		TransactionID: DomainTransactionIDToDbTransactionId(&domainOutpoint.TransactionID),
		Index:         domainOutpoint.Index,
	}
}

func DbOutpointToDomainOutpoint(dbOutpoint *DbOutpoint) (*externalapi.DomainOutpoint, error) {
	domainTransactionID, err := DbTransactionIdToDomainTransactionID(dbOutpoint.TransactionID)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainOutpoint{
		TransactionID: *domainTransactionID,
		Index:         dbOutpoint.Index,
	}, nil
}
