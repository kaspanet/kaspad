package hashset

import (
	"strings"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// HashSet is an unsorted unique collection of DomainHashes
type HashSet map[externalapi.DomainHash]struct{}

// New creates and returns an empty HashSet
func New() HashSet {
	return HashSet{}
}

// NewFromSlice creates and returns a HashSet with contents according to provided slice
func NewFromSlice(hashes ...*externalapi.DomainHash) HashSet {
	set := New()

	for _, hash := range hashes {
		set.Add(hash)
	}

	return set
}

// String returns a string representation of this hash set
func (hs HashSet) String() string {
	hashStrings := make([]string, 0, len(hs))
	for hash := range hs {
		hashStrings = append(hashStrings, hash.String())
	}
	return strings.Join(hashStrings, ", ")
}

// Add appends a hash to this HashSet. If given hash already exists - does nothing
func (hs HashSet) Add(hash *externalapi.DomainHash) {
	hs[*hash] = struct{}{}
}

// Remove removes a hash from this HashSet. If given hash does not exist in HashSet - does nothing.
func (hs HashSet) Remove(hash *externalapi.DomainHash) {
	delete(hs, *hash)
}

// Contains returns true if this HashSet contains the given hash.
func (hs HashSet) Contains(hash *externalapi.DomainHash) bool {
	_, ok := hs[*hash]
	return ok
}

// Subtract creates and returns a new HashSet that contains all hashes in this HashSet minus the ones in `other`
func (hs HashSet) Subtract(other HashSet) HashSet {
	diff := New()

	for hash := range hs {
		if !other.Contains(&hash) {
			diff.Add(&hash)
		}
	}

	return diff
}

// ContainsAllInSlice returns true if this HashSet contains all hashes in given slice
func (hs HashSet) ContainsAllInSlice(slice []*externalapi.DomainHash) bool {
	for _, hash := range slice {
		if !hs.Contains(hash) {
			return false
		}
	}

	return true
}

// ToSlice converts this HashSet into a slice of hashes
func (hs HashSet) ToSlice() []*externalapi.DomainHash {
	slice := make([]*externalapi.DomainHash, 0, len(hs))

	for hash := range hs {
		slice = append(slice, &hash)
	}

	return slice
}

// Length returns the length of this HashSet
func (hs HashSet) Length() int {
	return hs.Length()
}
