package consensusstatemanager

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

func (csm *consensusStateManager) updateVirtual(tips []*externalapi.DomainHash) error {

	return nil
}

func (csm *consensusStateManager) selectVirtualParents(tips []*externalapi.DomainHash) []*externalapi.DomainHash {
	var newVirtualParents []*externalapi.DomainHash

	return newVirtualParents
}
