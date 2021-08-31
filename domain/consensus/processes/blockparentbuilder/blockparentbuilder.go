package blockparentbuilder

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/pow"
)

type blockParentBuilder struct {
	databaseContext       model.DBManager
	blockHeaderStore      model.BlockHeaderStore
	dagTopologyManager    model.DAGTopologyManager
	reachabilityDataStore model.ReachabilityDataStore
	pruningStore          model.PruningStore
}

// New creates a new instance of a BlockParentBuilder
func New(
	databaseContext model.DBManager,
	blockHeaderStore model.BlockHeaderStore,
	dagTopologyManager model.DAGTopologyManager,
	reachabilityDataStore model.ReachabilityDataStore,
	pruningStore model.PruningStore,
) model.BlockParentBuilder {
	return &blockParentBuilder{
		databaseContext:       databaseContext,
		blockHeaderStore:      blockHeaderStore,
		dagTopologyManager:    dagTopologyManager,
		reachabilityDataStore: reachabilityDataStore,
		pruningStore:          pruningStore,
	}
}

func (bpb *blockParentBuilder) BuildParents(stagingArea *model.StagingArea,
	directParentHashes []*externalapi.DomainHash) ([]externalapi.BlockLevelParents, error) {

	pruningPoint, hasPruningPoint, err := bpb.pruningPoint(stagingArea)
	if err != nil {
		return nil, err
	}
	pruningPointParents := []externalapi.BlockLevelParents{}
	if hasPruningPoint {
		var err error
		pruningPointParents, err = bpb.pruningPointParents(stagingArea, pruningPoint)
		if err != nil {
			return nil, err
		}
	}

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

		// Level 0 parents are always equal to the direct parents
		// This explicit assignment is mainly useful for tests where the proof of work is not checked
		parentsMap[0] = append(parentsMap[0], &directParentHash)

		proofOfWorkValue := pow.CalculateProofOfWorkValue(directParentHeader.ToMutable())
		for blockLevel := 1; proofOfWorkValue.Bit(blockLevel) == 0; blockLevel++ {
			parentsMap[blockLevel] = append(parentsMap[blockLevel], &directParentHash)
		}
	}

	// Find the future-most parents for every block level. Note that for some block
	// levels it will be the indirect parents that are the most in the future
	for _, directParentHeader := range directParentHeaders {
		for blockLevel, blockLevelParentsInHeader := range directParentHeader.Parents() {
			blockLevelParentsInMap := parentsMap[blockLevel]

			// Get the pruning point parents for the block level (if they exist)
			pruningPointBlockLevelParents := externalapi.BlockLevelParents{}
			if len(pruningPointParents) > blockLevel {
				pruningPointBlockLevelParents = pruningPointParents[blockLevel]
			}

			// Copy the header parents and replace any pruned header parents with
			// the pruning point parents for this block level
			headerParents := externalapi.BlockLevelParents{}
			unprocessedHeaderParentsContainPrunedBlocks := false
			for _, headerParent := range blockLevelParentsInHeader {
				hasReachabilityData, err := bpb.reachabilityDataStore.HasReachabilityData(bpb.databaseContext, stagingArea, headerParent)
				if err != nil {
					return nil, err
				}
				if !hasReachabilityData {
					unprocessedHeaderParentsContainPrunedBlocks = true
					continue
				}
				headerParents = append(headerParents, headerParent)
			}
			if unprocessedHeaderParentsContainPrunedBlocks {
				headerParents = append(headerParents, pruningPointBlockLevelParents...)
			}

			// Copy the map parents and replace any pruned header parents with
			// the pruning point parents for this block level
			mapParents := externalapi.BlockLevelParents{}
			unprocessedMapParentsContainPruningBlocks := false
			for _, mapParent := range blockLevelParentsInMap {
				hasReachabilityData, err := bpb.reachabilityDataStore.HasReachabilityData(bpb.databaseContext, stagingArea, mapParent)
				if err != nil {
					return nil, err
				}
				if !hasReachabilityData {
					unprocessedMapParentsContainPruningBlocks = true
					continue
				}
				mapParents = append(mapParents, mapParent)
			}
			if unprocessedMapParentsContainPruningBlocks {
				mapParents = append(mapParents, pruningPointBlockLevelParents...)
			}

			newBlockLevelParents := externalapi.BlockLevelParents{}

			// Include in the new parents collection for this block level any
			// parents that exist in both the map and the header
			unprocessedHeaderParents := externalapi.BlockLevelParents{}
			unprocessedMapParents := externalapi.BlockLevelParents{}
			for _, headerParent := range headerParents {
				found := false
				for _, mapParent := range mapParents {
					if headerParent.Equal(mapParent) {
						found = true
						break
					}
				}
				if found {
					newBlockLevelParents = append(newBlockLevelParents, headerParent)
					continue
				}
				unprocessedHeaderParents = append(unprocessedHeaderParents, headerParent)
			}
			for _, mapParent := range mapParents {
				found := false
				for _, headerParent := range unprocessedHeaderParents {
					if mapParent.Equal(headerParent) {
						found = true
						break
					}
				}
				if found {
					if !newBlockLevelParents.Contains(mapParent) {
						newBlockLevelParents = append(newBlockLevelParents, mapParent)
					}
					continue
				}
				unprocessedMapParents = append(unprocessedMapParents, mapParent)
			}

			// Include in the new parents collection for this block level any
			// map parents that don't have any header parents in their future
			for _, mapParent := range unprocessedMapParents {
				// If the map parent is one of the pruning point parents for
				// this level, use the pruning point for the topological
				// comparison instead
				mapParentForTopologyComparison := mapParent
				if pruningPointBlockLevelParents.Contains(mapParent) {
					mapParentForTopologyComparison = pruningPoint
				}

				foundDescendantOfMapParent := false
				for _, headerParent := range unprocessedHeaderParents {
					// If the header parent is one of the pruning point parents
					// for this level, use the pruning point for the topological
					// comparison instead (unless the map parent is also a
					// parent of the pruning point, in which case skip)
					headerParentForTopologyComparison := headerParent
					if pruningPointBlockLevelParents.Contains(headerParent) {
						if mapParentForTopologyComparison.Equal(pruningPoint) {
							continue
						}
						headerParentForTopologyComparison = pruningPoint
					}
					isMapParentAncestorOfHeaderParent, err := bpb.dagTopologyManager.IsAncestorOf(stagingArea,
						mapParentForTopologyComparison, headerParentForTopologyComparison)
					if err != nil {
						return nil, err
					}
					if isMapParentAncestorOfHeaderParent {
						foundDescendantOfMapParent = true
						break
					}
				}
				if !foundDescendantOfMapParent {
					if !newBlockLevelParents.Contains(mapParent) {
						newBlockLevelParents = append(newBlockLevelParents, mapParent)
					}
				}
			}

			// Include in the new parents collection for this block level any
			// header parents that don't have any map parents in their future
			for _, headerParent := range unprocessedHeaderParents {
				// If the header parent is one of the pruning point parents for
				// this level, use the pruning point for the topological
				// comparison instead
				headerParentForTopologyComparison := headerParent
				if pruningPointBlockLevelParents.Contains(headerParent) {
					headerParentForTopologyComparison = pruningPoint
				}

				foundDescendantOfHeaderParent := false
				for _, mapParent := range unprocessedMapParents {
					// If the map parent is one of the pruning point parents
					// for this level, use the pruning point for the topological
					// comparison instead (unless the header parent is also a
					// parent of the pruning point, in which case skip)
					mapParentForTopologyComparison := mapParent
					if pruningPointBlockLevelParents.Contains(mapParent) {
						if headerParentForTopologyComparison.Equal(pruningPoint) {
							continue
						}
						mapParentForTopologyComparison = pruningPoint
					}
					isHeaderParentAncestorOfMapParent, err := bpb.dagTopologyManager.IsAncestorOf(stagingArea,
						headerParentForTopologyComparison, mapParentForTopologyComparison)
					if err != nil {
						return nil, err
					}
					if isHeaderParentAncestorOfMapParent {
						foundDescendantOfHeaderParent = true
						break
					}
				}
				if !foundDescendantOfHeaderParent {
					if !newBlockLevelParents.Contains(headerParent) {
						newBlockLevelParents = append(newBlockLevelParents, headerParent)
					}
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

func (bpb *blockParentBuilder) pruningPoint(stagingArea *model.StagingArea) (*externalapi.DomainHash, bool, error) {
	hasPruningPoint, err := bpb.pruningStore.HasPruningPoint(bpb.databaseContext, stagingArea)
	if err != nil {
		return nil, false, err
	}
	if !hasPruningPoint {
		return nil, false, nil
	}
	pruningPoint, err := bpb.pruningStore.PruningPoint(bpb.databaseContext, stagingArea)
	if err != nil {
		return nil, false, err
	}
	return pruningPoint, true, nil
}

func (bpb *blockParentBuilder) pruningPointParents(stagingArea *model.StagingArea, pruningPoint *externalapi.DomainHash) ([]externalapi.BlockLevelParents, error) {
	pruningPointHeader, err := bpb.blockHeaderStore.BlockHeader(bpb.databaseContext, stagingArea, pruningPoint)
	if err != nil {
		return nil, err
	}
	pruningPointParents := pruningPointHeader.Parents()
	return pruningPointParents, nil
}
