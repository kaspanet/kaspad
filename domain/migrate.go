package domain

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
	"math"
)

func (d *domain) migrate() error {
	log.Infof("Starting migration")
	pruningPoint, err := d.Consensus().PruningPoint()
	if err != nil {
		return err
	}
	log.Infof("Current pruning point: %s", pruningPoint)

	if d.consensusConfig.Params.GenesisHash.Equal(pruningPoint) {
		err = d.initStagingConsensus(d.consensusConfig)
		if err != nil {
			return err
		}
	} else {
		err = d.InitStagingConsensusWithoutGenesis()
		if err != nil {
			return err
		}

		err = syncConsensuses(d.Consensus(), d.StagingConsensus())
		if err != nil {
			return err
		}
	}

	err = d.CommitStagingConsensus()
	if err != nil {
		return err
	}

	log.Info("Done migrating")
	return nil
}

func syncConsensuses(syncer, syncee externalapi.Consensus) error {
	pruningPointProof, err := syncer.BuildPruningPointProof()
	if err != nil {
		return err
	}

	err = syncee.ApplyPruningPointProof(pruningPointProof)
	if err != nil {
		return err
	}

	pruningPointHeaders, err := syncer.PruningPointHeaders()
	if err != nil {
		return err
	}

	err = syncee.ImportPruningPoints(pruningPointHeaders)
	if err != nil {
		return err
	}

	pruningPointAndItsAnticone, err := syncer.PruningPointAndItsAnticone()
	if err != nil {
		return err
	}

	for _, blockHash := range pruningPointAndItsAnticone {
		block, found, err := syncer.GetBlock(blockHash)
		if err != nil {
			return err
		}

		if !found {
			return errors.Errorf("block %s is missing", blockHash)
		}

		blockDAAWindowHashes, err := syncer.BlockDAAWindowHashes(blockHash)
		if err != nil {
			return err
		}

		ghostdagDataBlockHashes, err := syncer.TrustedBlockAssociatedGHOSTDAGDataBlockHashes(blockHash)
		if err != nil {
			return err
		}

		blockWithTrustedData := &externalapi.BlockWithTrustedData{
			Block:        block,
			DAAWindow:    make([]*externalapi.TrustedDataDataDAAHeader, 0, len(blockDAAWindowHashes)),
			GHOSTDAGData: make([]*externalapi.BlockGHOSTDAGDataHashPair, 0, len(ghostdagDataBlockHashes)),
		}

		for i, daaBlockHash := range blockDAAWindowHashes {
			trustedDataDataDAAHeader, err := syncer.TrustedDataDataDAAHeader(blockHash, daaBlockHash, uint64(i))
			if err != nil {
				return err
			}
			blockWithTrustedData.DAAWindow = append(blockWithTrustedData.DAAWindow, trustedDataDataDAAHeader)
		}

		for _, ghostdagDataBlockHash := range ghostdagDataBlockHashes {
			data, err := syncer.TrustedGHOSTDAGData(ghostdagDataBlockHash)
			if err != nil {
				return err
			}
			blockWithTrustedData.GHOSTDAGData = append(blockWithTrustedData.GHOSTDAGData, &externalapi.BlockGHOSTDAGDataHashPair{
				Hash:         ghostdagDataBlockHash,
				GHOSTDAGData: data,
			})
		}

		err = syncee.ValidateAndInsertBlockWithTrustedData(blockWithTrustedData, false)
		if err != nil {
			return err
		}
	}

	syncerVirtualSelectedParent, err := syncer.GetVirtualSelectedParent()
	if err != nil {
		return err
	}

	pruningPoint, err := syncer.PruningPoint()
	if err != nil {
		return err
	}

	missingBlocks, _, err := syncer.GetHashesBetween(pruningPoint, syncerVirtualSelectedParent, math.MaxUint64)
	if err != nil {
		return err
	}

	syncerTips, err := syncer.Tips()
	if err != nil {
		return err
	}

	for _, tip := range syncerTips {
		if tip.Equal(syncerVirtualSelectedParent) {
			continue
		}

		anticone, err := syncer.GetAnticone(syncerVirtualSelectedParent, tip, 0)
		if err != nil {
			return err
		}

		missingBlocks = append(missingBlocks, anticone...)
	}

	percents := 0
	for i, blocksHash := range missingBlocks {
		blockInfo, err := syncee.GetBlockInfo(blocksHash)
		if err != nil {
			return err
		}

		if blockInfo.Exists {
			continue
		}

		block, found, err := syncer.GetBlock(blocksHash)
		if err != nil {
			return err
		}

		if !found {
			return errors.Errorf("block %s is missing", blocksHash)
		}
		err = syncee.ValidateAndInsertBlock(block, false)
		if err != nil {
			return err
		}

		newPercents := 100 * i / len(missingBlocks)
		if newPercents > percents {
			percents = newPercents
			log.Infof("Processed %d%% of the blocks", 100*i/len(missingBlocks))
		}
	}

	var fromOutpoint *externalapi.DomainOutpoint
	const step = 100_000
	for {
		outpointAndUTXOEntryPairs, err := syncer.GetPruningPointUTXOs(pruningPoint, fromOutpoint, step)
		if err != nil {
			return err
		}
		fromOutpoint = outpointAndUTXOEntryPairs[len(outpointAndUTXOEntryPairs)-1].Outpoint
		err = syncee.AppendImportedPruningPointUTXOs(outpointAndUTXOEntryPairs)
		if err != nil {
			return err
		}
		if len(outpointAndUTXOEntryPairs) < step {
			break
		}
	}

	// Check that ValidateAndInsertImportedPruningPoint works given the right arguments.
	err = syncee.ValidateAndInsertImportedPruningPoint(pruningPoint)
	if err != nil {
		return err
	}

	emptyCoinbase := &externalapi.DomainCoinbaseData{
		ScriptPublicKey: &externalapi.ScriptPublicKey{
			Script:  nil,
			Version: 0,
		},
	}

	// Check that we can build a block just after importing the pruning point.
	_, err = syncee.BuildBlock(emptyCoinbase, nil)
	if err != nil {
		return err
	}

	estimatedVirtualDAAScoreTarget, err := syncer.GetVirtualDAAScore()
	if err != nil {
		return err
	}

	err = syncer.ResolveVirtual(func(virtualDAAScoreStart uint64, virtualDAAScore uint64) {
		if estimatedVirtualDAAScoreTarget-virtualDAAScoreStart <= 0 {
			percents = 100
		} else {
			percents = int(float64(virtualDAAScore-virtualDAAScoreStart) / float64(estimatedVirtualDAAScoreTarget-virtualDAAScoreStart) * 100)
		}
		log.Infof("Resolving virtual. Estimated progress: %d%%", percents)
	})
	if err != nil {
		return err
	}

	log.Infof("Resolved virtual")

	return nil
}
