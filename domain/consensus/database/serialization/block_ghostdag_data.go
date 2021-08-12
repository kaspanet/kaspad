package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"math/big"
)

// BlockGHOSTDAGDataToDBBlockGHOSTDAGData converts BlockGHOSTDAGData to DbBlockGhostdagData
func BlockGHOSTDAGDataToDBBlockGHOSTDAGData(blockGHOSTDAGData *externalapi.BlockGHOSTDAGData) *DbBlockGhostdagData {
	var selectedParent *DbHash
	if blockGHOSTDAGData.SelectedParent() != nil {
		selectedParent = DomainHashToDbHash(blockGHOSTDAGData.SelectedParent())
	}

	return &DbBlockGhostdagData{
		BlueScore:          blockGHOSTDAGData.BlueScore(),
		BlueWork:           blockGHOSTDAGData.BlueWork().Bytes(),
		SelectedParent:     selectedParent,
		MergeSetBlues:      DomainHashesToDbHashes(blockGHOSTDAGData.MergeSetBlues()),
		MergeSetReds:       DomainHashesToDbHashes(blockGHOSTDAGData.MergeSetReds()),
		BluesAnticoneSizes: bluesAnticoneSizesToDBBluesAnticoneSizes(blockGHOSTDAGData.BluesAnticoneSizes()),
	}
}

// DBBlockGHOSTDAGDataToBlockGHOSTDAGData converts DbBlockGhostdagData to BlockGHOSTDAGData
func DBBlockGHOSTDAGDataToBlockGHOSTDAGData(dbBlockGHOSTDAGData *DbBlockGhostdagData) (*externalapi.BlockGHOSTDAGData, error) {
	var selectedParent *externalapi.DomainHash
	if dbBlockGHOSTDAGData.SelectedParent != nil {
		var err error
		selectedParent, err = DbHashToDomainHash(dbBlockGHOSTDAGData.SelectedParent)
		if err != nil {
			return nil, err
		}
	}

	mergetSetBlues, err := DbHashesToDomainHashes(dbBlockGHOSTDAGData.MergeSetBlues)
	if err != nil {
		return nil, err
	}

	mergetSetReds, err := DbHashesToDomainHashes(dbBlockGHOSTDAGData.MergeSetReds)
	if err != nil {
		return nil, err
	}

	bluesAnticoneSizes, err := dbBluesAnticoneSizesToBluesAnticoneSizes(dbBlockGHOSTDAGData.BluesAnticoneSizes)
	if err != nil {
		return nil, err
	}

	return externalapi.NewBlockGHOSTDAGData(
		dbBlockGHOSTDAGData.BlueScore,
		new(big.Int).SetBytes(dbBlockGHOSTDAGData.BlueWork),
		selectedParent,
		mergetSetBlues,
		mergetSetReds,
		bluesAnticoneSizes,
	), nil
}
