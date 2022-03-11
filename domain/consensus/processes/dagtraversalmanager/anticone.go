package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
)

func (dtm *dagTraversalManager) AnticoneFromVirtualPOV(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (
	[]*externalapi.DomainHash, error) {

	virtualParents, err := dtm.dagTopologyManager.Parents(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	return dtm.AnticoneFromBlocks(stagingArea, virtualParents, blockHash, 0)
}

func (dtm *dagTraversalManager) AnticoneFromBlocks(stagingArea *model.StagingArea, tips []*externalapi.DomainHash,
	blockHash *externalapi.DomainHash, maxTraversalAllowed uint64) (
	[]*externalapi.DomainHash, error) {

	anticone := []*externalapi.DomainHash{}
	queue := tips
	visited := hashset.New()

	traversalCounter := uint64(0)
	for len(queue) > 0 {
		var current *externalapi.DomainHash
		current, queue = queue[0], queue[1:]

		if visited.Contains(current) {
			continue
		}

		visited.Add(current)

		currentIsAncestorOfBlock, err := dtm.dagTopologyManager.IsAncestorOf(stagingArea, current, blockHash)
		if err != nil {
			return nil, err
		}

		if currentIsAncestorOfBlock {
			continue
		}

		blockIsAncestorOfCurrent, err := dtm.dagTopologyManager.IsAncestorOf(stagingArea, blockHash, current)
		if err != nil {
			return nil, err
		}

		// We count the number of blocks in past(tips) \setminus past(blockHash).
		// We don't use `len(visited)` since it includes some maximal blocks in past(blockHash) as well.
		traversalCounter++
		if maxTraversalAllowed > 0 && traversalCounter > maxTraversalAllowed {
			return nil, model.ErrReachedMaxTraversalAllowed
		}

		if !blockIsAncestorOfCurrent {
			anticone = append(anticone, current)
		}

		currentParents, err := dtm.dagTopologyManager.Parents(stagingArea, current)
		if err != nil {
			return nil, err
		}

		for _, parent := range currentParents {
			queue = append(queue, parent)
		}
	}

	return anticone, nil
}
