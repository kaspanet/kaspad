package utxo

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

type utxoOutpointEntryPair struct {
	outpoint externalapi.DomainOutpoint
	entry    externalapi.UTXOEntry
}

type utxoCollectionIterator struct {
	index int
	pairs []utxoOutpointEntryPair
}

func (uc utxoCollection) Iterator() externalapi.ReadOnlyUTXOSetIterator {
	pairs := make([]utxoOutpointEntryPair, len(uc))
	i := 0
	for outpoint, entry := range uc {
		pairs[i] = utxoOutpointEntryPair{
			outpoint: outpoint,
			entry:    entry,
		}
		i++
	}
	return &utxoCollectionIterator{index: -1, pairs: pairs}
}

func (uci *utxoCollectionIterator) First() bool {
	uci.index = 0
	return len(uci.pairs) > 0
}

func (uci *utxoCollectionIterator) Next() bool {
	uci.index++
	return uci.index < len(uci.pairs)
}

func (uci *utxoCollectionIterator) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry externalapi.UTXOEntry, err error) {
	pair := uci.pairs[uci.index]
	return &pair.outpoint, pair.entry, nil
}

func (uci *utxoCollectionIterator) WithDiff(diff externalapi.UTXODiff) (externalapi.ReadOnlyUTXOSetIterator, error) {
	d, ok := diff.(*immutableUTXODiff)
	if !ok {
		return nil, errors.New("diff is not of type *immutableUTXODiff")
	}

	return &readOnlyUTXOIteratorWithDiff{
		baseIterator:  uci,
		diff:          d,
		toAddIterator: diff.ToAdd().Iterator(),
	}, nil
}
