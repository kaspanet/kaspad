package transactionid

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
)

// Less returns true iff transaction ID a is less than hash b
func Less(a, b *externalapi.DomainTransactionID) bool {
	return hashes.Less((*externalapi.DomainHash)(a), (*externalapi.DomainHash)(b))
}
