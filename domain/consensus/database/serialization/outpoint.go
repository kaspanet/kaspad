package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// DomainOutpointToDbOutpoint converts DomainOutpoint to DbOutpoint
func DomainOutpointToDbOutpoint(domainOutpoint *externalapi.DomainOutpoint) *DbOutpoint {
	return &DbOutpoint{
		TransactionID: DomainTransactionIDToDbTransactionID(&domainOutpoint.TransactionID),
		Index:         domainOutpoint.Index,
	}
}

// DbOutpointToDomainOutpoint converts DbOutpoint to DomainOutpoint
func DbOutpointToDomainOutpoint(dbOutpoint *DbOutpoint) (*externalapi.DomainOutpoint, error) {
	domainTransactionID, err := DbTransactionIDToDomainTransactionID(dbOutpoint.TransactionID)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainOutpoint{
		TransactionID: *domainTransactionID,
		Index:         dbOutpoint.Index,
	}, nil
}
