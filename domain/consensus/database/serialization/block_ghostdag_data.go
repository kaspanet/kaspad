package serialization

import "github.com/kaspanet/kaspad/domain/consensus/model"

// BlockGHOSTDAGDataToDBBlockGHOSTDAGData converts BlockGHOSTDAGData to DbBlockGhostdagData
func BlockGHOSTDAGDataToDBBlockGHOSTDAGData(blockGHOSTDAGData *model.BlockGHOSTDAGData) *DbBlockGhostdagData {
	return &DbBlockGhostdagData{
		BlueScore:          blockGHOSTDAGData.BlueScore,
		SelectedParent:     DomainHashToDbHash(blockGHOSTDAGData.SelectedParent),
		MergeSetBlues:      DomainHashesToDbHashes(blockGHOSTDAGData.MergeSetBlues),
		MergeSetReds:       DomainHashesToDbHashes(blockGHOSTDAGData.MergeSetReds),
		BluesAnticoneSizes: bluesAnticoneSizesToDBBluesAnticoneSizes(blockGHOSTDAGData.BluesAnticoneSizes),
	}
}

// DBBlockGHOSTDAGDataToBlockGHOSTDAGData converts DbBlockGhostdagData to BlockGHOSTDAGData
func DBBlockGHOSTDAGDataToBlockGHOSTDAGData(dbBlockGHOSTDAGData *DbBlockGhostdagData) (*model.BlockGHOSTDAGData, error) {
	selectedParent, err := DbHashToDomainHash(dbBlockGHOSTDAGData.SelectedParent)
	if err != nil {
		return nil, err
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

	return &model.BlockGHOSTDAGData{
		BlueScore:          dbBlockGHOSTDAGData.BlueScore,
		SelectedParent:     selectedParent,
		MergeSetBlues:      mergetSetBlues,
		MergeSetReds:       mergetSetReds,
		BluesAnticoneSizes: bluesAnticoneSizes,
	}, nil
}
