package ghostdagmanager

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type blockHeapNode struct {
	blockHash externalapi.DomainHash
	ghostdagData
}
