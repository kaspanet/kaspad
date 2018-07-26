package blockdag

import (
		"fmt"
)

// utxoDiff  represents a diff between two UTXO Sets
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

// inverted returns a new utxoDiff with the ToAdd and ToRemove fields inverted
func (d *utxoDiff) inverted() *utxoDiff {
	return &utxoDiff{
		toAdd:    d.toRemove,
		toRemove: d.toAdd,
	}
}

// diff returns a new utxoDiff with the difference of this and other
// Assumes that if a txOut exists in both diffs, it's underlying values would be the same
func (d *utxoDiff) diff(other *utxoDiff) (*utxoDiff, error) {
	result := newUTXODiff()

	// Note that the following cases are not accounted for, as they are impossible
	// as long as the base UTXOSet is the same:
	// - if tx is in d.toAdd and other.toRemove
	// - if tx is in d.toRemove and other.toAdd

	// All transactions in d.toAdd:
	// If they are not in other.toAdd - should be added in result.toRemove
	// If they are in other.toRemove - base utxoSet is not the same
	for id, tx := range d.toAdd {
		for idx, txOut := range tx {
			if _, ok := other.toAdd[id][idx]; !ok {
				result.toRemove.add(id, idx, txOut)
			}
			if _, ok := other.toRemove[id][idx]; ok {
				return nil, fmt.Errorf("diff: transaction both in d.toAdd and in other.toRemove")
			}
		}
	}

	// All transactions in d.toRemove:
	// If they are not in other.toRemove - should be added in result.toAdd
	// If they are in other.toAdd - base utxoSet is not the same
	for id, tx := range d.toRemove {
		for idx, txOut := range tx {
			if _, ok := other.toRemove[id][idx]; !ok {
				result.toAdd.add(id, idx, txOut)
			}
			if _,ok := other.toAdd[id][idx]; ok {
				return nil, fmt.Errorf("diff: transaction both in d.toRemove and in other.toAdd")
			}
		}
	}

	// All transactions in other.toAdd:
	// If they are not in d.toAdd - should be added in result.toAdd
	for id, tx := range other.toAdd {
		for idx, txOut := range tx {
			if _, ok := d.toAdd[id][idx]; !ok {
				result.toAdd.add(id, idx, txOut)
			}
		}
	}

	// All transactions in other.toRemove:
	// If they are not in d.toRemove - should be added in result.toRemove
	for id, tx := range other.toRemove {
		for idx, txOut := range tx {
			if _, ok := d.toRemove[id][idx]; !ok {
				result.toRemove.add(id, idx, txOut)
			}
		}
	}

	return result, nil
}

// withDiff applies provided diff to this diff, creating a new diff, that would be the result if
// first d, and than diff were applied to the same base
func (d *utxoDiff) withDiff(diff *utxoDiff) (*utxoDiff, error) {
	result := newUTXODiff()

	// All transactions in d.toAdd:
	// If they are not in diff.toRemove - should be added in result.toAdd
	// If they are in diff.toAdd - should throw an error
	// Otherwise - should be ignored
	for id, tx := range d.toAdd {
		for idx, txOut := range tx {
			if _, ok := diff.toRemove[id][idx]; !ok {
				result.toAdd.add(id, idx, txOut)
			}
			if _, ok := diff.toAdd[id][idx]; ok {
				return nil, fmt.Errorf("withDiff: transaction both in d.toAdd and in other.toAdd")
			}
		}
	}

	// All transactions in d.toRemove:
	// If they are not in diff.toAdd - should be added in result.toRemove
	// If they are in diff.toRemove - should throw an error
	// Otherwise - should be ignored
	for id, tx := range d.toRemove {
		for idx, txOut := range tx {
			if _, ok := diff.toAdd[id][idx]; !ok {
				result.toRemove.add(id, idx, txOut)
			}
			if _, ok := diff.toRemove[id][idx]; ok {
				return nil, fmt.Errorf("withDiff: transaction both in d.toRemove and in other.toRemove")
			}
		}
	}

	// All transactions in diff.toAdd:
	// If they are not in d.toRemove - should be added in result.toAdd
	for id, tx := range diff.toAdd {
		for idx, txOut := range tx {
			if _, ok := d.toRemove[id][idx]; !ok {
				result.toAdd.add(id, idx, txOut)
			}
		}
	}

	// All transactions in diff.toRemove:
	// If they are not in d.toAdd - should be added in result.toRemove
	for id, tx := range diff.toRemove {
		for idx, txOut := range tx {
			if _, ok := d.toAdd[id][idx]; !ok {
				result.toRemove.add(id, idx, txOut)
			}
		}
	}

	return result, nil
}

// clone returns a clone of this UTXO diff
func (d *utxoDiff) clone() *utxoDiff {
	return &utxoDiff{
		toAdd:    d.toAdd.clone(),
		toRemove: d.toRemove.clone(),
	}
}

func (d utxoDiff) String() string {
	return fmt.Sprintf("toAdd: %s; toRemove: %s", d.toAdd, d.toRemove)
}