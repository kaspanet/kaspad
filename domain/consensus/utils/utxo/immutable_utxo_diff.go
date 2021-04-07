package utxo

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

type immutableUTXODiff struct {
	mutableUTXODiff *mutableUTXODiff

	isInvalidated bool
}

func (iud *immutableUTXODiff) ToAdd() externalapi.UTXOCollection {
	if iud.isInvalidated {
		panic("Attempt to read from an invalidated UTXODiff")
	}

	return iud.mutableUTXODiff.ToAdd()
}

func (iud *immutableUTXODiff) ToRemove() externalapi.UTXOCollection {
	if iud.isInvalidated {
		panic("Attempt to read from an invalidated UTXODiff")
	}

	return iud.mutableUTXODiff.ToRemove()
}

func (iud *immutableUTXODiff) WithDiff(other externalapi.UTXODiff) (externalapi.UTXODiff, error) {
	if iud.isInvalidated {
		panic("Attempt to read from an invalidated UTXODiff")
	}

	return iud.mutableUTXODiff.WithDiff(other)
}

func (iud *immutableUTXODiff) DiffFrom(other externalapi.UTXODiff) (externalapi.UTXODiff, error) {
	if iud.isInvalidated {
		panic("Attempt to read from an invalidated UTXODiff")
	}

	return iud.mutableUTXODiff.DiffFrom(other)
}

// NewUTXODiff creates an empty UTXODiff
func NewUTXODiff() externalapi.UTXODiff {
	return newUTXODiff()
}

func newUTXODiff() *immutableUTXODiff {
	return &immutableUTXODiff{
		mutableUTXODiff: newMutableUTXODiff(),
		isInvalidated:   false,
	}
}

// NewUTXODiffFromCollections returns a new UTXODiff with the given toAdd and toRemove collections
func NewUTXODiffFromCollections(toAdd, toRemove externalapi.UTXOCollection) (externalapi.UTXODiff, error) {
	add, ok := toAdd.(utxoCollection)
	if !ok {
		return nil, errors.New("toAdd is not of type utxoCollection")
	}
	remove, ok := toRemove.(utxoCollection)
	if !ok {
		return nil, errors.New("toRemove is not of type utxoCollection")
	}
	return &immutableUTXODiff{
		mutableUTXODiff: &mutableUTXODiff{
			toAdd:    add,
			toRemove: remove,
		},
	}, nil
}

func (iud *immutableUTXODiff) CloneMutable() externalapi.MutableUTXODiff {
	if iud.isInvalidated {
		panic("Attempt to read from an invalidated UTXODiff")
	}
	return iud.cloneMutable()
}

func (iud *immutableUTXODiff) Reversed() externalapi.UTXODiff {
	if iud.isInvalidated {
		panic("Attempt to read from an invalidated UTXODiff")
	}
	return &immutableUTXODiff{
		mutableUTXODiff: iud.mutableUTXODiff.Reversed(),
		isInvalidated:   false,
	}
}

func (iud *immutableUTXODiff) cloneMutable() *mutableUTXODiff {
	if iud == nil {
		return nil
	}

	return iud.mutableUTXODiff.clone()
}

func (iud immutableUTXODiff) String() string {
	return iud.mutableUTXODiff.String()
}
