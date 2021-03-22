package multiset

import (
	"github.com/kaspanet/go-muhash"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

type multiset struct {
	ms *muhash.MuHash
}

func (m multiset) Add(data []byte) {
	m.ms.Add(data)
}

func (m multiset) Remove(data []byte) {
	m.ms.Remove(data)
}

func (m multiset) Hash() *externalapi.DomainHash {
	finalizedHash := m.ms.Finalize()
	return externalapi.NewDomainHashFromByteArray(finalizedHash.AsArray())
}

func (m multiset) Serialize() []byte {
	return m.ms.Serialize()[:]
}

func (m multiset) Clone() model.Multiset {
	return &multiset{ms: m.ms.Clone()}
}

// FromBytes deserializes the given bytes slice and returns a multiset.
func FromBytes(multisetBytes []byte) (model.Multiset, error) {
	serialized := &muhash.SerializedMuHash{}
	if len(serialized) != len(multisetBytes) {
		return nil, errors.Errorf("mutliset bytes expected to be in length of %d but got %d",
			len(serialized), len(multisetBytes))
	}
	copy(serialized[:], multisetBytes)
	ms, err := muhash.DeserializeMuHash(serialized)
	if err != nil {
		return nil, err
	}

	return &multiset{ms: ms}, nil
}

// New returns a new model.Multiset
func New() model.Multiset {
	return &multiset{ms: muhash.NewMuHash()}
}
