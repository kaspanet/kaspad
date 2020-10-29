package serialization

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

func DomainTransactionIDToDbTransactionId(domainTransactionID *externalapi.DomainTransactionID) *DbTransactionId {
	panic("implement me")
}

func DbTransactionIdToDomainTransactionID(dbTransactionId *DbTransactionId) (*externalapi.DomainTransactionID, error) {
	panic("implement me")
}
