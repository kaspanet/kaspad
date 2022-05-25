package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

func (bp *blockProcessor) validateAndInsertBlockWithTrustedData(stagingArea *model.StagingArea,
	block *externalapi.BlockWithTrustedData, validateUTXO bool) (*externalapi.VirtualChangeSet, externalapi.BlockStatus, error) {

	blockHash := consensushashing.BlockHash(block.Block)
	for i, daaBlock := range block.DAAWindow {
		hash := consensushashing.HeaderHash(daaBlock.Header)
		bp.blocksWithTrustedDataDAAWindowStore.Stage(stagingArea, blockHash, uint64(i), &externalapi.BlockGHOSTDAGDataHashPair{
			Hash:         hash,
			GHOSTDAGData: daaBlock.GHOSTDAGData,
		})
		bp.blockHeaderStore.Stage(stagingArea, hash, daaBlock.Header)
	}

	blockReplacedGHOSTDAGData, err := bp.ghostdagDataWithoutPrunedBlocks(stagingArea, block.GHOSTDAGData[0].GHOSTDAGData)
	if err != nil {
		return nil, externalapi.StatusInvalid, err
	}
	bp.ghostdagDataStore.Stage(stagingArea, blockHash, blockReplacedGHOSTDAGData, false)

	for _, pair := range block.GHOSTDAGData {
		bp.ghostdagDataStore.Stage(stagingArea, pair.Hash, pair.GHOSTDAGData, true)
	}

	bp.daaBlocksStore.StageDAAScore(stagingArea, blockHash, block.Block.Header.DAAScore())
	return bp.validateAndInsertBlock(stagingArea, block.Block, false, validateUTXO, true)
}

func (bp *blockProcessor) ghostdagDataWithoutPrunedBlocks(stagingArea *model.StagingArea,
	data *externalapi.BlockGHOSTDAGData) (*externalapi.BlockGHOSTDAGData, error) {
	mergeSetBlues := make([]*externalapi.DomainHash, 0, len(data.MergeSetBlues()))
	for _, blockHash := range data.MergeSetBlues() {
		isPruned, err := bp.isPruned(stagingArea, blockHash)
		if err != nil {
			return nil, err
		}
		if isPruned {
			if data.SelectedParent().Equal(blockHash) {
				mergeSetBlues = append(mergeSetBlues, model.VirtualGenesisBlockHash)
			}
			continue
		}

		mergeSetBlues = append(mergeSetBlues, blockHash)
	}

	mergeSetReds := make([]*externalapi.DomainHash, 0, len(data.MergeSetReds()))
	for _, blockHash := range data.MergeSetReds() {
		isPruned, err := bp.isPruned(stagingArea, blockHash)
		if err != nil {
			return nil, err
		}
		if isPruned {
			continue
		}

		mergeSetReds = append(mergeSetReds, blockHash)
	}

	selectedParent := data.SelectedParent()
	isPruned, err := bp.isPruned(stagingArea, data.SelectedParent())
	if err != nil {
		return nil, err
	}

	if isPruned {
		selectedParent = model.VirtualGenesisBlockHash
	}

	return externalapi.NewBlockGHOSTDAGData(
		data.BlueScore(),
		data.BlueWork(),
		selectedParent,
		mergeSetBlues,
		mergeSetReds,
		data.BluesAnticoneSizes(),
	), nil
}

func (bp *blockProcessor) isPruned(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	_, err := bp.ghostdagDataStore.Get(bp.databaseContext, stagingArea, blockHash, false)
	if database.IsNotFoundError(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return false, nil
}
