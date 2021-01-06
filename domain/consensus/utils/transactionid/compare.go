package transactionid

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// Less returns true iff transaction ID a is less than hash b
func Less(a, b *externalapi.DomainTransactionID) bool {
	return externalapi.Less((*externalapi.DomainHash)(a), (*externalapi.DomainHash)(b))
}
