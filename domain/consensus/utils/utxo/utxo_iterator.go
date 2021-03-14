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
	index    int
	pairs    []utxoOutpointEntryPair
	isClosed bool
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
	if uci.isClosed {
		panic("Tried using a closed utxoCollectionIterator")
	}
	uci.index = 0
	return len(uci.pairs) > 0
}

func (uci *utxoCollectionIterator) Next() bool {
	if uci.isClosed {
		panic("Tried using a closed utxoCollectionIterator")
	}
	uci.index++
	return uci.index < len(uci.pairs)
}

func (uci *utxoCollectionIterator) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry externalapi.UTXOEntry, err error) {
	if uci.isClosed {
		return nil, nil, errors.New("Tried using a closed utxoCollectionIterator")
	}
	pair := uci.pairs[uci.index]
	return &pair.outpoint, pair.entry, nil
}

func (uci *utxoCollectionIterator) WithDiff(diff externalapi.UTXODiff) (externalapi.ReadOnlyUTXOSetIterator, error) {
	if uci.isClosed {
		return nil, errors.New("Tried using a closed utxoCollectionIterator")
	}
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

func (uci *utxoCollectionIterator) Close() error {
	if uci.isClosed {
		return errors.New("Tried using a closed utxoCollectionIterator")
	}
	uci.isClosed = true
	uci.pairs = nil
	return nil
}
