package utxo

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

type utxoDiff struct {
	toAdd    utxoCollection
	toRemove utxoCollection
}

// NewUTXODiff creates an empty UTXODiff
func NewUTXODiff() model.UTXODiff {
	return newUTXODiff()
}

func newUTXODiff() *utxoDiff {
	return &utxoDiff{
		toAdd:    utxoCollection{},
		toRemove: utxoCollection{},
	}
}

// NewUTXODiffFromCollections returns a new UTXODiff with the given toAdd and toRemove collections
func NewUTXODiffFromCollections(toAdd, toRemove model.UTXOCollection) (model.UTXODiff, error) {
	add, ok := toAdd.(utxoCollection)
	if !ok {
		return nil, errors.New("toAdd is not of type utxoCollection")
	}
	remove, ok := toRemove.(utxoCollection)
	if !ok {
		return nil, errors.New("toRemove is not of type utxoCollection")
	}
	return &utxoDiff{
		toAdd:    add,
		toRemove: remove,
	}, nil
}

func (d *utxoDiff) String() string {
	return fmt.Sprintf("toAdd: %s; toRemove: %s", d.toAdd, d.toRemove)
}

func (d *utxoDiff) WithDiff(other model.UTXODiff) (model.UTXODiff, error) {
	o, ok := other.(*utxoDiff)
	if !ok {
		return nil, errors.New("other is not of type *utxoDiff")
	}

	return withDiff(d, o)
}

func (d *utxoDiff) DiffFrom(other model.UTXODiff) (model.UTXODiff, error) {
	o, ok := other.(*utxoDiff)
	if !ok {
		return nil, errors.New("other is not of type *utxoDiff")
	}

	return diffFrom(d, o)
}

func (d *utxoDiff) ToAdd() model.UTXOCollection {
	return d.toAdd
}

func (d *utxoDiff) ToRemove() model.UTXOCollection {
	return d.toRemove
}

// Clone returns a clone of utxoDiff
func (d *utxoDiff) Clone() model.UTXODiff {
	return d.clone()
}

func (d *utxoDiff) clone() *utxoDiff {
	if d == nil {
		return nil
	}

	return &utxoDiff{
		toAdd:    d.toAdd.Clone(),
		toRemove: d.toRemove.Clone(),
	}
}

func (d *utxoDiff) CloneMutable() model.MutableUTXODiff {
	return d.cloneMutable()
}

func (d *utxoDiff) cloneMutable() *mutableUTXODiff {
	if d == nil {
		return nil
	}

	return &mutableUTXODiff{utxoDiff: d.clone()}
}

func (d *utxoDiff) addEntry(outpoint *externalapi.DomainOutpoint, entry externalapi.UTXOEntry) error {
	if d.toRemove.containsWithBlueScore(outpoint, entry.BlockBlueScore()) {
		d.toRemove.remove(outpoint)
	} else if d.toAdd.Contains(outpoint) {
		return errors.Errorf("AddEntry: Cannot add outpoint %s twice", outpoint)
	} else {
		d.toAdd.add(outpoint, entry)
	}
	return nil
}

func (d *utxoDiff) removeEntry(outpoint *externalapi.DomainOutpoint, entry externalapi.UTXOEntry) error {
	if d.toAdd.containsWithBlueScore(outpoint, entry.BlockBlueScore()) {
		d.toAdd.remove(outpoint)
	} else if d.toRemove.Contains(outpoint) {
		return errors.Errorf("removeEntry: Cannot remove outpoint %s twice", outpoint)
	} else {
		d.toRemove.add(outpoint, entry)
	}
	return nil
}
