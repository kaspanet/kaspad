package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// BlockGHOSTDAGDataHashPairToDbBlockGhostdagDataHashPair converts *externalapi.BlockGHOSTDAGDataHashPair to *DbBlockGHOSTDAGDataHashPair
func BlockGHOSTDAGDataHashPairToDbBlockGhostdagDataHashPair(pair *externalapi.BlockGHOSTDAGDataHashPair) *DbBlockGHOSTDAGDataHashPair {
	return &DbBlockGHOSTDAGDataHashPair{
		Hash:         DomainHashToDbHash(pair.Hash),
		GhostdagData: BlockGHOSTDAGDataToDBBlockGHOSTDAGData(pair.GHOSTDAGData),
	}
}

// DbBlockGHOSTDAGDataHashPairToBlockGHOSTDAGDataHashPair converts *DbBlockGHOSTDAGDataHashPair to *externalapi.BlockGHOSTDAGDataHashPair
func DbBlockGHOSTDAGDataHashPairToBlockGHOSTDAGDataHashPair(dbPair *DbBlockGHOSTDAGDataHashPair) (*externalapi.BlockGHOSTDAGDataHashPair, error) {
	hash, err := DbHashToDomainHash(dbPair.Hash)
	if err != nil {
		return nil, err
	}

	ghostdagData, err := DBBlockGHOSTDAGDataToBlockGHOSTDAGData(dbPair.GhostdagData)
	if err != nil {
		return nil, err
	}

	return &externalapi.BlockGHOSTDAGDataHashPair{
		Hash:         hash,
		GHOSTDAGData: ghostdagData,
	}, nil
}
