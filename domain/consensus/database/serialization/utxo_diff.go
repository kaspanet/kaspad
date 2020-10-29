package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// UTXODiffToDBUTXODiff converts UTXODiff to DbUtxoDiff
func UTXODiffToDBUTXODiff(diff *model.UTXODiff) *DbUtxoDiff {
	return &DbUtxoDiff{
		ToAdd:    utxoCollectionToDBUTXOCollection(diff.ToAdd),
		ToRemove: utxoCollectionToDBUTXOCollection(diff.ToRemove),
	}
}

// DBUTXODiffToUTXODiff converts DbUtxoDiff to UTXODiff
func DBUTXODiffToUTXODiff(diff *DbUtxoDiff) (*model.UTXODiff, error) {
	toAdd, err := dbUTXOCollectionToUTXOCollection(diff.ToAdd)
	if err != nil {
		return nil, err
	}

	toRemove, err := dbUTXOCollectionToUTXOCollection(diff.ToRemove)
	if err != nil {
		return nil, err
	}

	return &model.UTXODiff{
		ToAdd:    toAdd,
		ToRemove: toRemove,
	}, nil
}
