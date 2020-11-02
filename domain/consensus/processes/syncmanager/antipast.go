package syncmanager

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

func (sm *syncManager) antiPastHashesBetween(lowHash, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	panic("implement me")
}

func (sm *syncManager) missingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	panic("implement me")
}

func (sm *syncManager) isBlockHeaderInPruningPointFutureAndVirtualPast(blockHash *externalapi.DomainHash) (bool, error) {
	panic("implement me")
}
