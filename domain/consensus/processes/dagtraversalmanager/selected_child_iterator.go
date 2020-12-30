package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

type selectedChildIterator struct {
	databaseContext    model.DBReader
	dagTopologyManager model.DAGTopologyManager

	reachabilityDataStore model.ReachabilityDataStore
	highHash              *externalapi.DomainHash
	current               *externalapi.DomainHash
}

func (s *selectedChildIterator) Next() bool {
	data, err := s.reachabilityDataStore.ReachabilityData(s.databaseContext, s.current)
	if err != nil {
		panic(err)
	}

	for _, child := range data.Children() {
		isChildInSelectedParentChainOfHighHash, err := s.dagTopologyManager.IsInSelectedParentChainOf(child, s.highHash)
		if err != nil {
			panic(err)
		}

		if isChildInSelectedParentChainOfHighHash {
			s.current = child
			return true
		}
	}
	return false
}

func (s *selectedChildIterator) Get() *externalapi.DomainHash {
	return s.current
}

func (dtm *dagTraversalManager) SelectedChildIterator(highHash, lowHash *externalapi.DomainHash) (model.BlockIterator, error) {
	isLowHashInSelectedParentChainOfHighHash, err := dtm.dagTopologyManager.IsInSelectedParentChainOf(lowHash, highHash)
	if err != nil {
		return nil, err
	}

	if !isLowHashInSelectedParentChainOfHighHash {
		return nil, errors.Errorf("%s is not in the selected parent chain of %s", highHash, lowHash)
	}
	return &selectedChildIterator{
		databaseContext:       dtm.databaseContext,
		dagTopologyManager:    dtm.dagTopologyManager,
		reachabilityDataStore: dtm.reachabilityDataStore,
		highHash:              highHash,
		current:               lowHash,
	}, nil
}
