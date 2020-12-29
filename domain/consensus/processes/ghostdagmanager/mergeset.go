package ghostdagmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"sort"
)

func (gm *ghostdagManager) mergeSetWithoutSelectedParent(selecteParent *externalapi.DomainHash,
	blockParents []*externalapi.DomainHash) ([]*externalapi.DomainHash, error) {

	mergeSetMap := make(map[externalapi.DomainHash]struct{}, gm.k)
	mergeSetSlice := make([]*externalapi.DomainHash, 0, gm.k)
	selectedParentPast := make(map[externalapi.DomainHash]struct{})
	queue := []*externalapi.DomainHash{}
	// Queueing all parents (other than the selected parent itself) for processing.
	for _, parent := range blockParents {
		if parent.Equal(selecteParent) {
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
		currentParents, err := gm.dagTopologyManager.Parents(current)
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

			isAncestorOfSelectedParent, err := gm.dagTopologyManager.IsAncestorOf(parent, selecteParent)
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

	err := gm.sortMergeSet(mergeSetSlice)
	if err != nil {
		return nil, err
	}

	return mergeSetSlice, nil
}

func (gm *ghostdagManager) sortMergeSet(mergeSetSlice []*externalapi.DomainHash) error {
	var err error
	sort.Slice(mergeSetSlice, func(i, j int) bool {
		if err != nil {
			return false
		}
		isLess, lessErr := gm.less(mergeSetSlice[i], mergeSetSlice[j])
		if lessErr != nil {
			err = lessErr
			return false
		}
		return isLess
	})
	return err
}
