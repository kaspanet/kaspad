package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
)

func (dtm *dagTraversalManager) Anticone(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	anticone := []*externalapi.DomainHash{}
	queue, err := dtm.consensusStateStore.Tips(nil, dtm.databaseContext)
	if err != nil {
		return nil, err
	}
	visited := hashset.New()

	for len(queue) > 0 {
		var current *externalapi.DomainHash
		current, queue = queue[0], queue[1:]

		if visited.Contains(current) {
			continue
		}

		visited.Add(current)

		currentIsAncestorOfBlock, err := dtm.dagTopologyManager.IsAncestorOf(nil, current, blockHash)
		if err != nil {
			return nil, err
		}

		if currentIsAncestorOfBlock {
			continue
		}

		blockIsAncestorOfCurrent, err := dtm.dagTopologyManager.IsAncestorOf(nil, blockHash, current)
		if err != nil {
			return nil, err
		}

		if !blockIsAncestorOfCurrent {
			anticone = append(anticone, current)
		}

		currentParents, err := dtm.dagTopologyManager.Parents(nil, current)
		if err != nil {
			return nil, err
		}

		for _, parent := range currentParents {
			queue = append(queue, parent)
		}
	}

	return anticone, nil
}
