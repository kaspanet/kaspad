package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

type selectedChildIterator struct {
	dagTraversalManager model.DAGTraversalManager

	includeLowHash    bool
	highHash, lowHash *externalapi.DomainHash
	current           *externalapi.DomainHash
	err               error
	isClosed          bool
	stagingArea       *model.StagingArea
}

func (s *selectedChildIterator) First() bool {
	if s.isClosed {
		panic("Tried using a closed SelectedChildIterator")
	}
	s.current = s.lowHash
	if s.includeLowHash {
		return true
	}

	return s.Next()
}

func (s *selectedChildIterator) Next() bool {
	if s.isClosed {
		panic("Tried using a closed SelectedChildIterator")
	}
	if s.err != nil {
		return true
	}

	selectedChild, err := s.dagTraversalManager.SelectedChild(s.stagingArea, s.highHash, s.current)
	if errors.Is(err, errNoSelectedChild) {
		return false
	}
	if err != nil {
		s.current = nil
		s.err = err
		return true
	}

	s.current = selectedChild
	return true
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
	s.highHash = nil
	s.lowHash = nil
	s.current = nil
	s.err = nil
	return nil
}

// SelectedChildIterator returns a BlockIterator that iterates from lowHash (exclusive) to highHash (inclusive) over
// highHash's selected parent chain
func (dtm *dagTraversalManager) SelectedChildIterator(stagingArea *model.StagingArea,
	highHash, lowHash *externalapi.DomainHash, includeLowHash bool) (model.BlockIterator, error) {

	isLowHashInSelectedParentChainOfHighHash, err := dtm.dagTopologyManager.IsInSelectedParentChainOf(
		stagingArea, lowHash, highHash)
	if err != nil {
		return nil, err
	}

	if !isLowHashInSelectedParentChainOfHighHash {
		return nil, errors.Errorf("%s is not in the selected parent chain of %s", highHash, lowHash)
	}
	return &selectedChildIterator{
		dagTraversalManager: dtm,
		includeLowHash:      includeLowHash,
		highHash:            highHash,
		lowHash:             lowHash,
		current:             lowHash,
		stagingArea:         stagingArea,
	}, nil
}

var errNoSelectedChild = errors.New("errNoSelectedChild")

func (dtm *dagTraversalManager) SelectedChild(stagingArea *model.StagingArea,
	context, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {

	data, err := dtm.reachabilityDataStore.ReachabilityData(dtm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	for _, child := range data.Children() {
		isChildInSelectedParentChainOfHighHash, err := dtm.dagTopologyManager.IsInSelectedParentChainOf(
			stagingArea, child, context)
		if err != nil {
			return nil, err
		}

		if isChildInSelectedParentChainOfHighHash {
			return child, nil
		}
	}

	return nil, errors.Wrapf(errNoSelectedChild, "no selected child for %s from the point of view of %s",
		blockHash, context)
}
