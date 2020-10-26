package utxoalgebra

import (
	"reflect"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

func diffClone(diff *model.UTXODiff) *model.UTXODiff {
	clone := &model.UTXODiff{
		ToAdd:    collectionClone(diff.ToAdd),
		ToRemove: collectionClone(diff.ToRemove),
	}
	return clone
}

func diffAddEntry(diff *model.UTXODiff, outpoint *externalapi.DomainOutpoint, entry *externalapi.UTXOEntry) error {
	if collectionContainsWithBlueScore(diff.ToRemove, outpoint, entry.BlockBlueScore) {
		collectionRemove(diff.ToRemove, outpoint)
	} else if _, exists := diff.ToAdd[*outpoint]; exists {
		return errors.Errorf("AddEntry: Cannot add outpoint %s twice", outpoint)
	} else {
		collectionAdd(diff.ToAdd, outpoint, entry)
	}
	return nil
}

func diffRemoveEntry(diff *model.UTXODiff, outpoint *externalapi.DomainOutpoint, entry *externalapi.UTXOEntry) error {
	if collectionContainsWithBlueScore(diff.ToAdd, outpoint, entry.BlockBlueScore) {
		collectionRemove(diff.ToAdd, outpoint)
	} else if _, exists := diff.ToRemove[*outpoint]; exists {
		return errors.Errorf("removeEntry: Cannot remove outpoint %s twice", outpoint)
	} else {
		collectionAdd(diff.ToRemove, outpoint, entry)
	}
	return nil
}

func diffEqual(this *model.UTXODiff, other *model.UTXODiff) bool {
	if this == nil || other == nil {
		return this == other
	}

	return reflect.DeepEqual(this.ToAdd, other.ToAdd) &&
		reflect.DeepEqual(this.ToRemove, other.ToRemove)
}
