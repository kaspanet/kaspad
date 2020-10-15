package ghostdagmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"sort"
)

func (gm *ghostdagManager) mergeSet(selecteParent *model.DomainHash,
	blockParents []*model.DomainHash) []*model.DomainHash {

	mergeSetMap := make(map[model.DomainHash]struct{}, gm.k)
	mergeSetSlice := make([]*model.DomainHash, 0, gm.k)
	selectedParentPast := make(map[model.DomainHash]struct{})
	queue := []*model.DomainHash{}
	// Queueing all parents (other than the selected parent itself) for processing.
	for _, parent := range blockParents {
		if *parent == *selecteParent {
			continue
		}
		mergeSetMap[*parent] = struct{}{}
		mergeSetSlice = append(mergeSetSlice, parent)
		queue = append(queue, parent)
	}

	for len(queue) > 0 {
		var current *model.DomainHash
		current, queue = queue[0], queue[1:]
		// For each parent of the current block we check whether it is in the past of the selected parent. If not,
		// we add the it to the resulting anticone-set and queue it for further processing.
		currentParents := gm.dagTopologyManager.Parents(current)
		for _, parent := range currentParents {
			if _, ok := mergeSetMap[*parent]; ok {
				continue
			}

			if _, ok := selectedParentPast[*parent]; ok {
				continue
			}

			isAncestorOfSelectedParent := gm.dagTopologyManager.IsAncestorOf(parent, selecteParent)

			if isAncestorOfSelectedParent {
				selectedParentPast[*parent] = struct{}{}
				continue
			}

			mergeSetMap[*parent] = struct{}{}
			mergeSetSlice = append(mergeSetSlice, parent)
			queue = append(queue, parent)
		}
	}

	sort.Slice(mergeSetSlice, func(i, j int) bool {
		return gm.less(mergeSetSlice[i], mergeSetSlice[j])
	})

	return mergeSetSlice
}
