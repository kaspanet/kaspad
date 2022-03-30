package blockparentbuilder

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
	"github.com/pkg/errors"
)

type blockParentBuilder struct {
	databaseContext       model.DBManager
	blockHeaderStore      model.BlockHeaderStore
	dagTopologyManager    model.DAGTopologyManager
	parentsManager        model.ParentsManager
	reachabilityDataStore model.ReachabilityDataStore
	pruningStore          model.PruningStore

	genesisHash   *externalapi.DomainHash
	maxBlockLevel int
}

// New creates a new instance of a BlockParentBuilder
func New(
	databaseContext model.DBManager,
	blockHeaderStore model.BlockHeaderStore,
	dagTopologyManager model.DAGTopologyManager,
	parentsManager model.ParentsManager,

	reachabilityDataStore model.ReachabilityDataStore,
	pruningStore model.PruningStore,

	genesisHash *externalapi.DomainHash,
	maxBlockLevel int,
) model.BlockParentBuilder {
	return &blockParentBuilder{
		databaseContext:    databaseContext,
		blockHeaderStore:   blockHeaderStore,
		dagTopologyManager: dagTopologyManager,
		parentsManager:     parentsManager,

		reachabilityDataStore: reachabilityDataStore,
		pruningStore:          pruningStore,
		genesisHash:           genesisHash,
		maxBlockLevel:         maxBlockLevel,
	}
}

func (bpb *blockParentBuilder) BuildParents(stagingArea *model.StagingArea,
	daaScore uint64, directParentHashes []*externalapi.DomainHash) ([]externalapi.BlockLevelParents, error) {

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

	type blockToReferences map[externalapi.DomainHash][]*externalapi.DomainHash
	candidatesByLevelToReferenceBlocksMap := make(map[int]blockToReferences)

	// Direct parents are guaranteed to be in one other's anticones so add them all to
	// all the block levels they occupy
	for _, directParentHeader := range directParentHeaders {
		directParentHash := consensushashing.HeaderHash(directParentHeader)
		blockLevel := directParentHeader.BlockLevel(bpb.maxBlockLevel)
		for i := 0; i <= blockLevel; i++ {
			if _, exists := candidatesByLevelToReferenceBlocksMap[i]; !exists {
				candidatesByLevelToReferenceBlocksMap[i] = make(map[externalapi.DomainHash][]*externalapi.DomainHash)
			}
			candidatesByLevelToReferenceBlocksMap[i][*directParentHash] = []*externalapi.DomainHash{directParentHash}
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
		for blockLevel, blockLevelParentsInHeader := range bpb.parentsManager.Parents(directParentHeader) {
			isEmptyLevel := false
			if _, exists := candidatesByLevelToReferenceBlocksMap[blockLevel]; !exists {
				candidatesByLevelToReferenceBlocksMap[blockLevel] = make(map[externalapi.DomainHash][]*externalapi.DomainHash)
				isEmptyLevel = true
			}

			for _, parent := range blockLevelParentsInHeader {
				isInFutureOfVirtualGenesisChildren := false
				hasReachabilityData, err := bpb.reachabilityDataStore.HasReachabilityData(bpb.databaseContext, stagingArea, parent)
				if err != nil {
					return nil, err
				}
				if hasReachabilityData {
					// If a block is in the future of one of the virtual genesis children it means we have the full DAG between the current block
					// and this parent, so there's no need for any indirect reference blocks, and normal reachability queries can be used.
					isInFutureOfVirtualGenesisChildren, err = bpb.dagTopologyManager.IsAnyAncestorOf(stagingArea, virtualGenesisChildren, parent)
					if err != nil {
						return nil, err
					}
				}

				// Reference blocks are the blocks that are used in reachability queries to check if
				// a candidate is in the future of another candidate. In most cases this is just the
				// block itself, but in the case where a block doesn't have reachability data we need
				// to use some blocks in its future as reference instead.
				// If we make sure to add a parent in the future of the pruning point first, we can
				// know that any pruned candidate that is in the past of some blocks in the pruning
				// point anticone should have should be a parent (in the relevant level) of one of
				// the virtual genesis children in the pruning point anticone. So we can check which
				// virtual genesis children have this block as parent and use those block as
				// reference blocks.
				var referenceBlocks []*externalapi.DomainHash
				if isInFutureOfVirtualGenesisChildren {
					referenceBlocks = []*externalapi.DomainHash{parent}
				} else {
					for childHash, childHeader := range virtualGenesisChildrenHeaders {
						childHash := childHash // Assign to a new pointer to avoid `range` pointer reuse
						if bpb.parentsManager.ParentsAtLevel(childHeader, blockLevel).Contains(parent) {
							referenceBlocks = append(referenceBlocks, &childHash)
						}
					}
				}

				if isEmptyLevel {
					candidatesByLevelToReferenceBlocksMap[blockLevel][*parent] = referenceBlocks
					continue
				}

				if !isInFutureOfVirtualGenesisChildren {
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

	parents := make([]externalapi.BlockLevelParents, 0, len(candidatesByLevelToReferenceBlocksMap))
	for blockLevel := 0; blockLevel < len(candidatesByLevelToReferenceBlocksMap); blockLevel++ {
		if blockLevel > 0 {
			if _, ok := candidatesByLevelToReferenceBlocksMap[blockLevel][*bpb.genesisHash]; ok && len(candidatesByLevelToReferenceBlocksMap[blockLevel]) == 1 {
				break
			}
		}

		levelBlocks := make(externalapi.BlockLevelParents, 0, len(candidatesByLevelToReferenceBlocksMap[blockLevel]))
		for block := range candidatesByLevelToReferenceBlocksMap[blockLevel] {
			block := block // Assign to a new pointer to avoid `range` pointer reuse
			levelBlocks = append(levelBlocks, &block)
		}

		parents = append(parents, levelBlocks)
	}
	return parents, nil
}
