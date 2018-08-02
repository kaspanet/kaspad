package blockdag

import (
	"fmt"
	"errors"
)

// utxoDiff represents a diff between two UTXO Sets
type utxoDiff struct {
	toAdd    utxoCollection
	toRemove utxoCollection
}

// newUTXODiff creates a new, empty utxoDiff
func newUTXODiff() *utxoDiff {
	return &utxoDiff{
		toAdd:    utxoCollection{},
		toRemove: utxoCollection{},
	}
}

// diffFrom returns a new utxoDiff with the difference between this utxoDiff and another
// Assumes that if a txOut exists in both utxoDiffs, its underlying values would be the same
func (d *utxoDiff) diffFrom(other *utxoDiff) (*utxoDiff, error) {
	result := newUTXODiff()

	// Note that the following cases are not accounted for, as they are impossible
	// as long as the base utxoSet is the same:
	// - if utxoEntry is in d.toAdd and other.toRemove
	// - if utxoEntry is in d.toRemove and other.toAdd

	// All transactions in d.toAdd:
	// If they are not in other.toAdd - should be added in result.toRemove
	// If they are in other.toRemove - base utxoSet is not the same
	for outPoint, utxoEntry := range d.toAdd {
		if _, ok := other.toAdd[outPoint]; !ok {
			result.toRemove[outPoint] = utxoEntry
		}
		if _, ok := other.toRemove[outPoint]; ok {
			return nil, errors.New("diffFrom: transaction both in d.toAdd and in other.toRemove")
		}
	}

	// All transactions in d.toRemove:
	// If they are not in other.toRemove - should be added in result.toAdd
	// If they are in other.toAdd - base utxoSet is not the same
	for outPoint, utxoEntry := range d.toRemove {
		if _, ok := other.toRemove[outPoint]; !ok {
			result.toAdd[outPoint] = utxoEntry
		}
		if _, ok := other.toAdd[outPoint]; ok {
			return nil, errors.New("diffFrom: transaction both in d.toRemove and in other.toAdd")
		}
	}

	// All transactions in other.toAdd:
	// If they are not in d.toAdd - should be added in result.toAdd
	for outPoint, utxoEntry := range other.toAdd {
		if _, ok := d.toAdd[outPoint]; !ok {
			result.toAdd[outPoint] = utxoEntry
		}
	}

	// All transactions in other.toRemove:
	// If they are not in d.toRemove - should be added in result.toRemove
	for outPoint, utxoEntry := range other.toRemove {
		if _, ok := d.toRemove[outPoint]; !ok {
			result.toRemove[outPoint] = utxoEntry
		}
	}

	return result, nil
}

// withDiff applies provided diff to this diff, creating a new utxoDiff, that would be the result if
// first d, and than diff were applied to the same base
func (d *utxoDiff) withDiff(diff *utxoDiff) (*utxoDiff, error) {
	result := newUTXODiff()

	// All transactions in d.toAdd:
	// If they are not in diff.toRemove - should be added in result.toAdd
	// If they are in diff.toAdd - should throw an error
	// Otherwise - should be ignored
	for outPoint, utxoEntry := range d.toAdd {
		if _, ok := diff.toRemove[outPoint]; !ok {
			result.toAdd[outPoint] = utxoEntry
		}
		if _, ok := diff.toAdd[outPoint]; ok {
			return nil, errors.New("withDiff: transaction both in d.toAdd and in other.toAdd")
		}
	}

	// All transactions in d.toRemove:
	// If they are not in diff.toAdd - should be added in result.toRemove
	// If they are in diff.toRemove - should throw an error
	// Otherwise - should be ignored
	for outPoint, utxoEntry := range d.toRemove {
		if _, ok := diff.toAdd[outPoint]; !ok {
			result.toRemove[outPoint] = utxoEntry
		}
		if _, ok := diff.toRemove[outPoint]; ok {
			return nil, errors.New("withDiff: transaction both in d.toRemove and in other.toRemove")
		}
	}

	// All transactions in diff.toAdd:
	// If they are not in d.toRemove - should be added in result.toAdd
	for outPoint, utxoEntry := range diff.toAdd {
		if _, ok := d.toRemove[outPoint]; !ok {
			result.toAdd[outPoint] = utxoEntry
		}
	}

	// All transactions in diff.toRemove:
	// If they are not in d.toAdd - should be added in result.toRemove
	for outPoint, utxoEntry := range diff.toRemove {
		if _, ok := d.toAdd[outPoint]; !ok {
			result.toRemove[outPoint] = utxoEntry
		}
	}

	return result, nil
}

// clone returns a clone of this utxoDiff
func (d *utxoDiff) clone() *utxoDiff {
	return &utxoDiff{
		toAdd:    d.toAdd.clone(),
		toRemove: d.toRemove.clone(),
	}
}

func (d utxoDiff) String() string {
	return fmt.Sprintf("toAdd: %s; toRemove: %s", d.toAdd, d.toRemove)
}
