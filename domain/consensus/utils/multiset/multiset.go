package multiset

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/pkg/errors"
)

type multiset struct {
	ms *secp256k1.MultiSet
}

func (m multiset) Add(data []byte) {
	m.ms.Add(data)
}

func (m multiset) Remove(data []byte) {
	m.ms.Remove(data)
}

func (m multiset) Hash() *externalapi.DomainHash {
	hash, err := hashes.FromBytes(m.ms.Finalize()[:])
	if err != nil {
		panic(errors.Errorf("this should never happen unless seckp hash size is different than %d",
			externalapi.DomainHashSize))
	}

	return hash
}

func (m multiset) Serialize() []byte {
	return m.ms.Serialize()[:]
}

func (m multiset) Clone() model.Multiset {
	msClone := *m.ms
	return &multiset{ms: &msClone}
}

// FromBytes deserializes the given bytes slice and returns a multiset.
func FromBytes(multisetBytes []byte) (model.Multiset, error) {
	serialized := &secp256k1.SerializedMultiSet{}
	if len(serialized) != len(multisetBytes) {
		return nil, errors.Errorf("mutliset bytes expected to be in length of %d but got %d",
			len(serialized), len(multisetBytes))
	}
	copy(serialized[:], multisetBytes)
	ms, err := secp256k1.DeserializeMultiSet(serialized)
	if err != nil {
		return nil, err
	}

	return &multiset{ms: ms}, nil
}

// New returns a new model.Multiset
func New() model.Multiset {
	return &multiset{ms: secp256k1.NewMultiset()}
}
