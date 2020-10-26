package utils

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/daghash"
)

func NewDomainTransactionIDFromString(str string) (*externalapi.DomainTransactionID, error) {
	hash, err := daghash.NewHashFromStr(str)
	return (*externalapi.DomainTransactionID)(hash), err
}
