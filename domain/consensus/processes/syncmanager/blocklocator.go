package syncmanager

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

func (sm *syncManager) createBlockLocator(lowHash, highHash *externalapi.DomainHash) (*externalapi.BlockLocator, error) {
	panic("implement me")
}

func (sm *syncManager) findNextBlockLocatorBoundaries(blockLocator *externalapi.BlockLocator) (lowHash, highHash *externalapi.DomainHash, err error) {
	panic("implement me")
}
