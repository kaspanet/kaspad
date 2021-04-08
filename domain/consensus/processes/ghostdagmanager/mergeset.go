package ghostdagmanager

import (
	"sort"

	"github.com/kaspanet/kaspad/domain/consensus/model"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (gm *ghostdagManager) mergeSetWithoutSelectedParent(stagingArea *model.StagingArea,
	selectedParent *externalapi.DomainHash, blockParents []*externalapi.DomainHash) ([]*externalapi.DomainHash, error) {

	mergeSetMap := make(map[externalapi.DomainHash]struct{}, gm.k)
	mergeSetSlice := make([]*externalapi.DomainHash, 0, gm.k)
	selectedParentPast := make(map[externalapi.DomainHash]struct{})
	queue := []*externalapi.DomainHash{}
	// Queueing all parents (other than the selected parent itself) for processing.
	for _, parent := range blockParents {
		if parent.Equal(selectedParent) {
			continue
		}
		mergeSetMap[*parent] = struct{}{}
		mergeSetSlice = append(mergeSetSlice, parent)
		queue = append(queue, parent)
	}

	for len(queue) > 0 {
		var current *externalapi.DomainHash
		current, queue = queue[0], queue[1:]
		// For each parent of the current block we check whether it is in the past of the selected parent. If not,
		// we add the it to the resulting anticone-set and queue it for further processing.
		currentParents, err := gm.dagTopologyManager.Parents(stagingArea, current)
		if err != nil {
			return nil, err
		}
		for _, parent := range currentParents {
			if _, ok := mergeSetMap[*parent]; ok {
				continue
			}

			if _, ok := selectedParentPast[*parent]; ok {
				continue
			}

			isAncestorOfSelectedParent, err := gm.dagTopologyManager.IsAncestorOf(stagingArea, parent, selectedParent)
			if err != nil {
				return nil, err
			}

			if isAncestorOfSelectedParent {
				selectedParentPast[*parent] = struct{}{}
				continue
			}

			mergeSetMap[*parent] = struct{}{}
			mergeSetSlice = append(mergeSetSlice, parent)
			queue = append(queue, parent)
		}
	}

	err := gm.sortMergeSet(stagingArea, mergeSetSlice)
	if err != nil {
		return nil, err
	}

	return mergeSetSlice, nil
}

func (gm *ghostdagManager) sortMergeSet(stagingArea *model.StagingArea, mergeSetSlice []*externalapi.DomainHash) error {
	var err error
	sort.Slice(mergeSetSlice, func(i, j int) bool {
		if err != nil {
			return false
		}
		isLess, lessErr := gm.less(stagingArea, mergeSetSlice[i], mergeSetSlice[j])
		if lessErr != nil {
			err = lessErr
			return false
		}
		return isLess
	})
	return err
}

func (gm *ghostdagManager) GetSortedMergeSet(stagingArea *model.StagingArea,
	current *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {

	currentGhostdagData, err := gm.ghostdagDataStore.Get(gm.databaseContext, stagingArea, current)
	if err != nil {
		return nil, err
	}

	blueMergeSet := currentGhostdagData.MergeSetBlues()
	redMergeSet := currentGhostdagData.MergeSetReds()
	sortedMergeSet := make([]*externalapi.DomainHash, 0, len(blueMergeSet)+len(redMergeSet))
	selectedParent, blueMergeSet := blueMergeSet[0], blueMergeSet[1:]
	sortedMergeSet = append(sortedMergeSet, selectedParent)
	i, j := 0, 0
	for i < len(blueMergeSet) && j < len(redMergeSet) {
		currentBlue := blueMergeSet[i]
		currentBlueGhostdagData, err := gm.ghostdagDataStore.Get(gm.databaseContext, stagingArea, currentBlue)
		if err != nil {
			return nil, err
		}
		currentRed := redMergeSet[j]
		currentRedGhostdagData, err := gm.ghostdagDataStore.Get(gm.databaseContext, stagingArea, currentRed)
		if err != nil {
			return nil, err
		}
		if gm.Less(currentBlue, currentBlueGhostdagData, currentRed, currentRedGhostdagData) {
			sortedMergeSet = append(sortedMergeSet, currentBlue)
			i++
		} else {
			sortedMergeSet = append(sortedMergeSet, currentRed)
			j++
		}
	}
	sortedMergeSet = append(sortedMergeSet, blueMergeSet[i:]...)
	sortedMergeSet = append(sortedMergeSet, redMergeSet[j:]...)

	return sortedMergeSet, nil
}
