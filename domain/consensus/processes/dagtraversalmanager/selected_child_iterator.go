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
	highHash, lowHash     *externalapi.DomainHash
	current               *externalapi.DomainHash
	err                   error
}

func (s *selectedChildIterator) First() bool {
	s.current = s.lowHash
	return s.Next()
}

func (s *selectedChildIterator) Next() bool {
	if s.err != nil {
		return true
	}

	data, err := s.reachabilityDataStore.ReachabilityData(s.databaseContext, s.current)
	if err != nil {
		s.current = nil
		s.err = err
		return true
	}

	for _, child := range data.Children() {
		isChildInSelectedParentChainOfHighHash, err := s.dagTopologyManager.IsInSelectedParentChainOf(child, s.highHash)
		if err != nil {
			s.current = nil
			s.err = err
			return true
		}

		if isChildInSelectedParentChainOfHighHash {
			s.current = child
			return true
		}
	}
	return false
}

func (s *selectedChildIterator) Get() (*externalapi.DomainHash, error) {
	return s.current, s.err
}

func (s *selectedChildIterator) Close() error {
	return nil
}

// SelectedChildIterator returns a BlockIterator that iterates from lowHash (exclusive) to highHash (inclusive) over
// highHash's selected parent chain
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
		lowHash:               lowHash,
		current:               lowHash,
	}, nil
}
