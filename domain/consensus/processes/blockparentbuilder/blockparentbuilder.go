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

	pruningPoint, err := bpb.pruningStore.PruningPoint(bpb.databaseContext, stagingArea)
	if err != nil {
		return nil, err
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
	candidatesByLevelToReferenceBlocksMap := make(map[int]map[externalapi.DomainHash][]*externalapi.DomainHash)

	for directParentHash, directParentHeader := range directParentHeaders {
		directParentHash := directParentHash // Assign to a new pointer to avoid `range` pointer reuse
		proofOfWorkValue := pow.CalculateProofOfWorkValue(directParentHeader.ToMutable())
		for blockLevel := 0; ; blockLevel++ {
			if _, exists := candidatesByLevelToReferenceBlocksMap[blockLevel]; !exists {
				candidatesByLevelToReferenceBlocksMap[blockLevel] = make(map[externalapi.DomainHash][]*externalapi.DomainHash)
			}
			candidatesByLevelToReferenceBlocksMap[blockLevel][directParentHash] = []*externalapi.DomainHash{&directParentHash}
			if proofOfWorkValue.Bit(blockLevel+1) != 0 {
				break
			}
		}
	}

	virtualGenesisChildren, err := bpb.dagTopologyManager.Children(stagingArea, model.VirtualGenesisBlockHash)
	if err != nil {
		return nil, err
	}

	virtualGenesisChildrenHeaders := make(map[externalapi.DomainHash]externalapi.BlockHeader, len(virtualGenesisChildren))
	for _, child := range virtualGenesisChildren {
		virtualGenesisChildrenHeaders[*child], err = bpb.blockHeaderStore.BlockHeader(bpb.databaseContext, stagingArea, child)
		if err != nil {
			return nil, err
		}
	}

	maybeAddDirectParentParents := func(directParentHeader externalapi.BlockHeader) error {
		for blockLevel, blockLevelParentsInHeader := range directParentHeader.Parents() {
			isEmptyLevel := false
			if _, exists := candidatesByLevelToReferenceBlocksMap[blockLevel]; !exists {
				candidatesByLevelToReferenceBlocksMap[blockLevel] = make(map[externalapi.DomainHash][]*externalapi.DomainHash)
				isEmptyLevel = true
			}

			for _, parent := range blockLevelParentsInHeader {
				hasReachabilityData, err := bpb.reachabilityDataStore.HasReachabilityData(bpb.databaseContext, stagingArea, parent)
				if err != nil {
					return err
				}

				var referenceBlocks []*externalapi.DomainHash
				if hasReachabilityData {
					referenceBlocks = []*externalapi.DomainHash{parent}
				} else {
					for childHash, childHeader := range virtualGenesisChildrenHeaders {
						childHash := childHash
						if len(childHeader.Parents()) <= blockLevel {
							continue
						}

						if childHeader.Parents()[blockLevel].Contains(parent) {
							referenceBlocks = append(referenceBlocks, &childHash)
						}
					}
				}

				if isEmptyLevel {
					candidatesByLevelToReferenceBlocksMap[blockLevel][*parent] = referenceBlocks
					continue
				}

				if !hasReachabilityData {
					continue
				}

				toRemove := hashset.New()
				isAncestorOfAnyCandidate := false
				for candidate, candidateReferences := range candidatesByLevelToReferenceBlocksMap[blockLevel] {
					candidate := candidate // Assign to a new pointer to avoid `range` pointer reuse
					isInFutureOfCurrentCandidate, err := bpb.dagTopologyManager.IsAnyAncestorOf(stagingArea, candidateReferences, parent)
					if err != nil {
						return err
					}

					if isInFutureOfCurrentCandidate {
						toRemove.Add(&candidate)
						continue
					}

					if isAncestorOfAnyCandidate {
						continue
					}

					// Maybe explicitly check the candidate if you see it has reachability data
					isAncestorOfCurrentCandidate, err := bpb.dagTopologyManager.IsAncestorOfAny(stagingArea, parent, candidateReferences)
					if err != nil {
						return err
					}

					if isAncestorOfCurrentCandidate {
						isAncestorOfAnyCandidate = true
					}
				}

				if toRemove.Length() > 0 {
					for hash := range toRemove {
						delete(candidatesByLevelToReferenceBlocksMap[blockLevel], hash)
					}
				}

				// We should add the block as a candidate if it's in the future of another candidate
				// or in the anticone of all candidates.
				if !isAncestorOfAnyCandidate || toRemove.Length() > 0 {
					candidatesByLevelToReferenceBlocksMap[blockLevel][*parent] = referenceBlocks
				}
			}
		}

		return nil
	}

	var firstAddedParent *externalapi.DomainHash
	for directParentHash, directParentHeader := range directParentHeaders {
		directParentHash := directParentHash
		isInFutureOfPruningPoint, err := bpb.dagTopologyManager.IsAncestorOf(stagingArea, pruningPoint, &directParentHash)
		if err != nil {
			return nil, err
		}

		if !isInFutureOfPruningPoint {
			continue
		}

		err = maybeAddDirectParentParents(directParentHeader)
		if err != nil {
			return nil, err
		}

		firstAddedParent = &directParentHash
		break
	}
	// Find the future-most parents for every block level. Note that for some block
	// levels it will be the indirect parents that are the most in the future
	for directParentHash, directParentHeader := range directParentHeaders {
		if directParentHash.Equal(firstAddedParent) {
			continue
		}

		err = maybeAddDirectParentParents(directParentHeader)
		if err != nil {
			return nil, err
		}
	}

	parents := make([]externalapi.BlockLevelParents, len(candidatesByLevelToReferenceBlocksMap))
	for blockLevel := 0; blockLevel < len(candidatesByLevelToReferenceBlocksMap); blockLevel++ {
		levelBlocks := make(externalapi.BlockLevelParents, 0, len(candidatesByLevelToReferenceBlocksMap[blockLevel]))
		for block := range candidatesByLevelToReferenceBlocksMap[blockLevel] {
			block := block // Assign to a new pointer to avoid `range` pointer reuse
			levelBlocks = append(levelBlocks, &block)
		}
		parents[blockLevel] = levelBlocks
	}
	return parents, nil
}
