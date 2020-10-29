package serialization

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

func DomainTransactionToDbTransaction(domainTransaction *externalapi.DomainTransaction) *DbTransaction {
	panic("implement me")
}

func DbTransactionToDomainTransaction(dbTransaction *DbTransaction) (*externalapi.DomainTransaction, error) {
	panic("implement me")
}
