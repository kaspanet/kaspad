package serialization

import "github.com/kaspanet/kaspad/domain/consensus/model"

// BlockGHOSTDAGDataToDBBlockGHOSTDAGData converts BlockGHOSTDAGData to DbBlockGhostdagData
func BlockGHOSTDAGDataToDBBlockGHOSTDAGData(blockGHOSTDAGData *model.BlockGHOSTDAGData) *DbBlockGhostdagData {
	bluesAnticoneSizes := make([]*DbBluesAnticoneSizes, len(blockGHOSTDAGData.BluesAnticoneSizes))
	i := 0
	for hash, anticoneSize := range blockGHOSTDAGData.BluesAnticoneSizes {
		bluesAnticoneSizes[i] = &DbBluesAnticoneSizes{
			BlueHash:     DomainHashToDbHash(&hash),
			AnticoneSize: uint32(anticoneSize),
		}
		i++
	}

	return &DbBlockGhostdagData{
		BlueScore:          blockGHOSTDAGData.BlueScore,
		SelectedParent:     DomainHashToDbHash(blockGHOSTDAGData.SelectedParent),
		MergeSetBlues:      DomainHashesToDbHashes(blockGHOSTDAGData.MergeSetBlues),
		MergeSetReds:       DomainHashesToDbHashes(blockGHOSTDAGData.MergeSetReds),
		BluesAnticoneSizes: bluesAnticoneSizes,
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
