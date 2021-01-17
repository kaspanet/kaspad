package utxo

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/pkg/errors"
)

type immutableUTXODiff struct {
	mutableUTXODiff *mutableUTXODiff

	isInvalidated bool
}

func (iud *immutableUTXODiff) ToAdd() model.UTXOCollection {
	if iud.isInvalidated {
		panic("Attempt to read from an invalidated UTXODiff")
	}

	return iud.mutableUTXODiff.ToAdd()
}

func (iud *immutableUTXODiff) ToRemove() model.UTXOCollection {
	if iud.isInvalidated {
		panic("Attempt to read from an invalidated UTXODiff")
	}

	return iud.mutableUTXODiff.ToRemove()
}

func (iud *immutableUTXODiff) WithDiff(other model.UTXODiff) (model.UTXODiff, error) {
	if iud.isInvalidated {
		panic("Attempt to read from an invalidated UTXODiff")
	}

	return iud.mutableUTXODiff.WithDiff(other)
}

func (iud *immutableUTXODiff) DiffFrom(other model.UTXODiff) (model.UTXODiff, error) {
	if iud.isInvalidated {
		panic("Attempt to read from an invalidated UTXODiff")
	}

	return iud.mutableUTXODiff.DiffFrom(other)
}

// NewUTXODiff creates an empty UTXODiff
func NewUTXODiff() model.UTXODiff {
	return newUTXODiff()
}

func newUTXODiff() *immutableUTXODiff {
	return &immutableUTXODiff{
		mutableUTXODiff: newMutableUTXODiff(),
		isInvalidated:   false,
	}
}

// NewUTXODiffFromCollections returns a new UTXODiff with the given toAdd and toRemove collections
func NewUTXODiffFromCollections(toAdd, toRemove model.UTXOCollection) (model.UTXODiff, error) {
	add, ok := toAdd.(Collection)
	if !ok {
		return nil, errors.New("toAdd is not of type Collection")
	}
	remove, ok := toRemove.(Collection)
	if !ok {
		return nil, errors.New("toRemove is not of type Collection")
	}
	return &immutableUTXODiff{
		mutableUTXODiff: &mutableUTXODiff{
			toAdd:    add,
			toRemove: remove,
		},
	}, nil
}

func (iud *immutableUTXODiff) CloneMutable() model.MutableUTXODiff {
	return iud.cloneMutable()
}

func (iud *immutableUTXODiff) cloneMutable() *mutableUTXODiff {
	if iud == nil {
		return nil
	}

	return iud.mutableUTXODiff.clone()
}
