package utxoserialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/utils/multiset"
)

// CalculateMultisetFromProtoUTXOSet calculates the Multiset corresponding to the given ProtuUTXOSet
func CalculateMultisetFromProtoUTXOSet(protoUTXOSet *ProtoUTXOSet) (model.Multiset, error) {
	ms := multiset.New()
	for _, utxo := range protoUTXOSet.Utxos {
		ms.Add(utxo.EntryOutpointPair)
	}
	return ms, nil
}
