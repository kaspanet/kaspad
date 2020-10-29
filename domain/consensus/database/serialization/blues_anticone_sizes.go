package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

func bluesAnticoneSizesToDBBluesAnticoneSizes(bluesAnticoneSizes map[externalapi.DomainHash]model.KType) []*DbBluesAnticoneSizes {
	dbBluesAnticoneSizes := make([]*DbBluesAnticoneSizes, len(bluesAnticoneSizes))
	i := 0
	for hash, anticoneSize := range bluesAnticoneSizes {
		dbBluesAnticoneSizes[i] = &DbBluesAnticoneSizes{
			BlueHash:     DomainHashToDbHash(&hash),
			AnticoneSize: uint32(anticoneSize),
		}
		i++
	}

	return dbBluesAnticoneSizes
}

func dbBluesAnticoneSizesToBluesAnticoneSizes(dbBluesAnticoneSizes []*DbBluesAnticoneSizes) (map[externalapi.DomainHash]model.KType, error) {
	bluesAnticoneSizes := make(map[externalapi.DomainHash]model.KType, len(dbBluesAnticoneSizes))

	for _, data := range dbBluesAnticoneSizes {
		hash, err := DbHashToDomainHash(data.BlueHash)
		if err != nil {
			return nil, err
		}

		bluesAnticoneSizes[*hash], err = uint32ToKType(data.AnticoneSize)
		if err != nil {
			return nil, err
		}
	}

	return bluesAnticoneSizes, nil
}

func uint32ToKType(n uint32) (model.KType, error) {
	convertedN := model.KType(n)
	if uint32(convertedN) != n {
		return 0, errors.Errorf("cannot convert %d to KType without losing data", n)
	}
	return convertedN, nil
}
