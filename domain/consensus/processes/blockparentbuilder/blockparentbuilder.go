package blockparentbuilder

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/pow"
)

type blockParentBuilder struct {
	databaseContext    model.DBManager
	blockHeaderStore   model.BlockHeaderStore
	dagTopologyManager model.DAGTopologyManager
}

// New creates a new instance of a BlockParentBuilder
func New(
	databaseContext model.DBManager,
	blockHeaderStore model.BlockHeaderStore,
	dagTopologyManager model.DAGTopologyManager,
) model.BlockParentBuilder {
	return &blockParentBuilder{
		databaseContext:    databaseContext,
		blockHeaderStore:   blockHeaderStore,
		dagTopologyManager: dagTopologyManager,
	}
}

func (bpb *blockParentBuilder) BuildParents(stagingArea *model.StagingArea,
	directParentHashes []*externalapi.DomainHash) ([]externalapi.BlockLevelParents, error) {

	directParentHeaders := make(map[externalapi.DomainHash]externalapi.BlockHeader)
	for _, directParentHash := range directParentHashes {
		directParentHeader, err := bpb.blockHeaderStore.BlockHeader(bpb.databaseContext, stagingArea, directParentHash)
		if err != nil {
			return nil, err
		}
		directParentHeaders[*directParentHash] = directParentHeader
	}

	// Direct parents are guaranteed to be in one other's anticones so add them all to
	// all the block levels they occupy
	parentsMap := make(map[int]externalapi.BlockLevelParents)
	for directParentHash, directParentHeader := range directParentHeaders {
		directParentHash := directParentHash // Assign to a new pointer to avoid `range` pointer reuse

		proofOfWorkValue := pow.CalculateProofOfWorkValue(directParentHeader.ToMutable())
		for blockLevel := 0; proofOfWorkValue.Bit(blockLevel) == 0; blockLevel++ {
			parentsMap[blockLevel] = append(parentsMap[blockLevel], &directParentHash)
		}
	}

	// Find the future-most parents for every block level. Note that for some block
	// levels it will be the indirect parents that are the most in the future
	for _, directParentHeader := range directParentHeaders {
		for blockLevel, blockLevelParentsInHeader := range directParentHeader.Parents() {
			blockLevelParentsInMap := parentsMap[blockLevel]
			newBlockLevelParents := externalapi.BlockLevelParents{}

			// Include map parents that don't have any header parents in their future
			for _, mapBlockLevelParent := range blockLevelParentsInMap {
				isMapParentAncestorOfAnyHeaderParent, err := bpb.dagTopologyManager.IsAncestorOfAny(stagingArea, mapBlockLevelParent, blockLevelParentsInHeader)
				if err != nil {
					return nil, err
				}
				if !isMapParentAncestorOfAnyHeaderParent {
					newBlockLevelParents = append(newBlockLevelParents, mapBlockLevelParent)
				}
			}

			// Include header parents that don't have any map parents in their future
			for _, headerBlockLevelParent := range blockLevelParentsInHeader {
				isHeaderParentAncestorOfAnyMapParent, err := bpb.dagTopologyManager.IsAncestorOfAny(stagingArea, headerBlockLevelParent, blockLevelParentsInMap)
				if err != nil {
					return nil, err
				}
				if !isHeaderParentAncestorOfAnyMapParent {
					newBlockLevelParents = append(newBlockLevelParents, headerBlockLevelParent)
				}
			}

			// Include any parents that exist in both the map and the header
			for _, mapBlockLevelParent := range blockLevelParentsInMap {
				found := false
				for _, headerBlockLevelParent := range blockLevelParentsInHeader {
					if mapBlockLevelParent.Equal(headerBlockLevelParent) {
						found = true
					}
				}
				if found {
					newBlockLevelParents = append(newBlockLevelParents, mapBlockLevelParent)
				}
			}

			parentsMap[blockLevel] = newBlockLevelParents
		}
	}

	parents := make([]externalapi.BlockLevelParents, len(parentsMap))
	for blockLevel := 0; blockLevel < len(parentsMap); blockLevel++ {
		parents[blockLevel] = parentsMap[blockLevel]
	}
	return parents, nil
}
