package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type Transaction interface {
	TransactionID() *externalapi.DomainTransactionID
	Transaction() *externalapi.DomainTransaction
}
