package blockparentbuilder

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/pow"
	"github.com/pkg/errors"
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

	// Late on we'll mutate direct parent hashes, so we first clone it.
	directParentHashesCopy := make([]*externalapi.DomainHash, len(directParentHashes))
	copy(directParentHashesCopy, directParentHashes)

	pruningPoint, err := bpb.pruningStore.PruningPoint(bpb.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	// The first candidates to be added should be from a parent in the future of the pruning
	// point, so later on we'll know that every block that doesn't have reachability data
	// (i.e. pruned) is necessarily in the past of the current candidates and cannot be
	// considered as a valid candidate.
	// This is why we sort the direct parent headers in a way that the first one will be
	// in the future of the pruning point.
	directParentHeaders := make([]externalapi.BlockHeader, len(directParentHashesCopy))
	firstParentInFutureOfPruningPointIndex := 0
	foundFirstParentInFutureOfPruningPoint := false
	for i, directParentHash := range directParentHashesCopy {
		isInFutureOfPruningPoint, err := bpb.dagTopologyManager.IsAncestorOf(stagingArea, pruningPoint, directParentHash)
		if err != nil {
			return nil, err
		}

		if !isInFutureOfPruningPoint {
			continue
		}

		firstParentInFutureOfPruningPointIndex = i
		foundFirstParentInFutureOfPruningPoint = true
		break
	}

	if !foundFirstParentInFutureOfPruningPoint {
		return nil, errors.New("BuildParents should get at least one parent in the future of the pruning point")
	}

	oldFirstDirectParent := directParentHashesCopy[0]
	directParentHashesCopy[0] = directParentHashesCopy[firstParentInFutureOfPruningPointIndex]
	directParentHashesCopy[firstParentInFutureOfPruningPointIndex] = oldFirstDirectParent

	for i, directParentHash := range directParentHashesCopy {
		directParentHeader, err := bpb.blockHeaderStore.BlockHeader(bpb.databaseContext, stagingArea, directParentHash)
		if err != nil {
			return nil, err
		}
		directParentHeaders[i] = directParentHeader
	}

	// Direct parents are guaranteed to be in one other's anticones so add them all to
	// all the block levels they occupy
	candidatesByLevelToReferenceBlocksMap := make(map[int]map[externalapi.DomainHash][]*externalapi.DomainHash)

	for _, directParentHeader := range directParentHeaders {
		directParentHash := consensushashing.HeaderHash(directParentHeader)
		proofOfWorkValue := pow.CalculateProofOfWorkValue(directParentHeader.ToMutable())
		for blockLevel := 0; ; blockLevel++ {
			if _, exists := candidatesByLevelToReferenceBlocksMap[blockLevel]; !exists {
				candidatesByLevelToReferenceBlocksMap[blockLevel] = make(map[externalapi.DomainHash][]*externalapi.DomainHash)
			}
			candidatesByLevelToReferenceBlocksMap[blockLevel][*directParentHash] = []*externalapi.DomainHash{directParentHash}
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

	for _, directParentHeader := range directParentHeaders {
		for blockLevel, blockLevelParentsInHeader := range directParentHeader.Parents() {
			isEmptyLevel := false
			if _, exists := candidatesByLevelToReferenceBlocksMap[blockLevel]; !exists {
				candidatesByLevelToReferenceBlocksMap[blockLevel] = make(map[externalapi.DomainHash][]*externalapi.DomainHash)
				isEmptyLevel = true
			}

			for _, parent := range blockLevelParentsInHeader {
				hasReachabilityData, err := bpb.reachabilityDataStore.HasReachabilityData(bpb.databaseContext, stagingArea, parent)
				if err != nil {
					return nil, err
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
						return nil, err
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
						return nil, err
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
