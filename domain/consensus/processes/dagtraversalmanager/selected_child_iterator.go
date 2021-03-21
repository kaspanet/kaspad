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
	isClosed              bool
}

func (s *selectedChildIterator) First() bool {
	if s.isClosed {
		panic("Tried using a closed SelectedChildIterator")
	}
	s.current = s.lowHash
	return s.Next()
}

func (s *selectedChildIterator) Next() bool {
	if s.isClosed {
		panic("Tried using a closed SelectedChildIterator")
	}
	if s.err != nil {
		return true
	}

	data, err := s.reachabilityDataStore.ReachabilityData(s.databaseContext, nil, s.current)
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
	if s.isClosed {
		return nil, errors.New("Tried using a closed SelectedChildIterator")
	}
	return s.current, s.err
}

func (s *selectedChildIterator) Close() error {
	if s.isClosed {
		return errors.New("Tried using a closed SelectedChildIterator")
	}
	s.isClosed = true
	s.databaseContext = nil
	s.dagTopologyManager = nil
	s.reachabilityDataStore = nil
	s.highHash = nil
	s.lowHash = nil
	s.current = nil
	s.err = nil
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
