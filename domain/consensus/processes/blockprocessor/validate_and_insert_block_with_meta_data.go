package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

func (bp *blockProcessor) validateAndInsertBlockWithMetaData(stagingArea *model.StagingArea, block *externalapi.BlockWithMetaData, validateUTXO bool) (*externalapi.BlockInsertionResult, error) {
	blockHash := consensushashing.BlockHash(block.Block)
	for _, pair := range block.DAAWindow {
		hash := consensushashing.HeaderHash(pair.Header)
		bp.blockHeaderStore.Stage(stagingArea, hash, pair.Header)
		ghostdagData, err := bp.replaceGHOSTDAGDataSelectedParentWithVirtualGenesisBlockHashIfNeeded(stagingArea, pair.GHOSTDAGData)
		if err != nil {
			return nil, err
		}

		bp.ghostdagDataStore.Stage(stagingArea, hash, ghostdagData)
	}

	for _, pair := range block.GHOSTDAGData {
		ghostdagData, err := bp.replaceGHOSTDAGDataSelectedParentWithVirtualGenesisBlockHashIfNeeded(stagingArea, pair.GHOSTDAGData)
		if err != nil {
			return nil, err
		}

		bp.ghostdagDataStore.Stage(stagingArea, pair.Hash, ghostdagData)
	}

	bp.daaBlocksStore.StageDAAScore(stagingArea, blockHash, block.DAAScore)
	return bp.validateAndInsertBlock(stagingArea, block.Block, false, validateUTXO, true)
}

func (bp *blockProcessor) replaceGHOSTDAGDataSelectedParentWithVirtualGenesisBlockHashIfNeeded(stagingArea *model.StagingArea, data *externalapi.BlockGHOSTDAGData) (*externalapi.BlockGHOSTDAGData, error) {
	exists, err := bp.blockStatusStore.Exists(bp.databaseContext, stagingArea, data.SelectedParent())
	if err != nil {
		return nil, err
	}

	if exists {
		return data, nil
	}

	return externalapi.NewBlockGHOSTDAGData(
		data.BlueScore(),
		data.BlueWork(),
		model.VirtualGenesisBlockHash,
		data.MergeSetBlues(),
		data.MergeSetReds(),
		data.BluesAnticoneSizes(),
	), nil
}
