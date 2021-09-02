package blockparentbuilder

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
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
	candidatesByLevelToReferenceBlockMap := make(map[int]map[externalapi.DomainHash]*externalapi.DomainHash)

	for directParentHash, directParentHeader := range directParentHeaders {
		directParentHash := directParentHash // Assign to a new pointer to avoid `range` pointer reuse
		proofOfWorkValue := pow.CalculateProofOfWorkValue(directParentHeader.ToMutable())
		for blockLevel := 0; ; blockLevel++ {
			if _, exists := candidatesByLevelToReferenceBlockMap[blockLevel]; !exists {
				candidatesByLevelToReferenceBlockMap[blockLevel] = make(map[externalapi.DomainHash]*externalapi.DomainHash)
			}
			candidatesByLevelToReferenceBlockMap[blockLevel][directParentHash] = &directParentHash
			if proofOfWorkValue.Bit(blockLevel+1) != 0 {
				break
			}
		}
	}

	// Find the future-most parents for every block level. Note that for some block
	// levels it will be the indirect parents that are the most in the future
	for _, directParentHeader := range directParentHeaders {
		for blockLevel, blockLevelParentsInHeader := range directParentHeader.Parents() {
			hasCheckedPruningPointParents := false

			isEmptyLevel := false
			if _, exists := candidatesByLevelToReferenceBlockMap[blockLevel]; !exists {
				candidatesByLevelToReferenceBlockMap[blockLevel] = make(map[externalapi.DomainHash]*externalapi.DomainHash)
				isEmptyLevel = true
			}

			for _, parent := range blockLevelParentsInHeader {
				hasReachabilityData, err := bpb.reachabilityDataStore.HasReachabilityData(bpb.databaseContext, stagingArea, parent)
				if err != nil {
					return nil, err
				}

				var blocksToCheck []*externalapi.DomainHash
				var referenceBlock *externalapi.DomainHash
				if hasReachabilityData {
					blocksToCheck = []*externalapi.DomainHash{parent}
					referenceBlock = parent
				} else {
					if hasCheckedPruningPointParents {
						continue
					}
					blocksToCheck = pruningPointParents[blockLevel]
					referenceBlock = pruningPoint
					hasCheckedPruningPointParents = true
				}

				if isEmptyLevel {
					for _, block := range blocksToCheck {
						candidatesByLevelToReferenceBlockMap[blockLevel][*block] = referenceBlock
					}
				} else {
					for _, block := range blocksToCheck {
						toRemove := hashset.New()
						isAncestorOfAnyCandidate := false
						for candidate, candidateReference := range candidatesByLevelToReferenceBlockMap[blockLevel] {
							candidate := candidate // Assign to a new pointer to avoid `range` pointer reuse
							isInFutureOfCurrentCandidate, err := bpb.isStrictAncestorOf(stagingArea, candidateReference, referenceBlock)
							if err != nil {
								return nil, err
							}

							if isInFutureOfCurrentCandidate {
								toRemove.Add(&candidate)
								continue
							}

							if isAncestorOfAnyCandidate {
								continue
							}

							isAncestorOfCurrentCandidate, err := bpb.dagTopologyManager.IsAncestorOf(stagingArea, referenceBlock, candidateReference)
							if err != nil {
								return nil, err
							}

							if isAncestorOfCurrentCandidate {
								isAncestorOfAnyCandidate = true
							}
						}

						if toRemove.Length() > 0 {
							for hash := range toRemove {
								delete(candidatesByLevelToReferenceBlockMap[blockLevel], hash)
							}
						}

						// We should add the block as a candidate if it's in the future of another candidate
						// or in the anticone of all candidates.
						if !isAncestorOfAnyCandidate || toRemove.Length() > 0 {
							candidatesByLevelToReferenceBlockMap[blockLevel][*block] = referenceBlock
						}
					}
				}
			}
		}
	}

	parents := make([]externalapi.BlockLevelParents, len(candidatesByLevelToReferenceBlockMap))
	for blockLevel := 0; blockLevel < len(candidatesByLevelToReferenceBlockMap); blockLevel++ {
		levelBlocks := make(externalapi.BlockLevelParents, 0, len(candidatesByLevelToReferenceBlockMap[blockLevel]))
		for block := range candidatesByLevelToReferenceBlockMap[blockLevel] {
			block := block // Assign to a new pointer to avoid `range` pointer reuse
			levelBlocks = append(levelBlocks, &block)
		}
		parents[blockLevel] = levelBlocks
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

func (bpb *blockParentBuilder) isStrictAncestorOf(stagingArea *model.StagingArea, blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	if blockHashA.Equal(blockHashB) {
		return false, nil
	}

	return bpb.dagTopologyManager.IsAncestorOf(stagingArea, blockHashA, blockHashB)
}
