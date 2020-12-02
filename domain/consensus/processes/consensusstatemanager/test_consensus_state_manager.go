package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
)

type testConsensusStateManager struct {
	*consensusStateManager
}

// NewTestConsensusStateManager creates an instance of a TestConsensusStateManager
func NewTestConsensusStateManager(baseConsensusStateManager model.ConsensusStateManager) testapi.TestConsensusStateManager {
	return &testConsensusStateManager{consensusStateManager: baseConsensusStateManager.(*consensusStateManager)}
}

func (csm testConsensusStateManager) AddUTXOToMultiset(
	multiset model.Multiset, entry *externalapi.UTXOEntry, outpoint *externalapi.DomainOutpoint) error {

	return addUTXOToMultiset(multiset, entry, outpoint)
}

func (csm testConsensusStateManager) ResolveBlockStatus(blockHash *externalapi.DomainHash) (externalapi.BlockStatus, error) {
	return csm.resolveBlockStatus(blockHash)
}

func (csm testConsensusStateManager) VirtualFinalityPoint() (*externalapi.DomainHash, error) {
	return csm.virtualFinalityPoint()
}
