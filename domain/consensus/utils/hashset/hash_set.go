package hashset

import (
	"strings"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type HashSet map[externalapi.DomainHash]struct{}

func New() HashSet {
	return HashSet{}
}

func NewFromSlice(hashes ...*externalapi.DomainHash) HashSet {
	set := New()

	for _, hash := range hashes {
		set.Add(hash)
	}

	return set
}

func (hs HashSet) String() string {
	hashStrings := make([]string, 0, len(hs))
	for hash := range hs {
		hashStrings = append(hashStrings, hash.String())
	}
	return strings.Join(hashStrings, ", ")
}

func (hs HashSet) Add(hash *externalapi.DomainHash) {
	hs[*hash] = struct{}{}
}

func (hs HashSet) Remove(hash *externalapi.DomainHash) {
	delete(hs, *hash)
}

func (hs HashSet) Contains(hash *externalapi.DomainHash) bool {
	_, ok := hs[*hash]
	return ok
}

func (hs HashSet) Subtract(other HashSet) HashSet {
	diff := New()

	for hash := range hs {
		if !other.Contains(&hash) {
			diff.Add(&hash)
		}
	}

	return diff
}
func (hs HashSet) ContainsAllInSlice(slice []*externalapi.DomainHash) bool {
	for _, hash := range slice {
		if !hs.Contains(hash) {
			return false
		}
	}

	return true
}

func (hs HashSet) ToSlice() []*externalapi.DomainHash {
	slice := make([]*externalapi.DomainHash, 0, len(hs))

	for hash := range hs {
		slice = append(slice, &hash)
	}

	return slice
}
