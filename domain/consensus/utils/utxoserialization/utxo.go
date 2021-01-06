package utxoserialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
)

// ReadOnlyUTXOSetToProtoUTXOSet converts ReadOnlyUTXOSetIterator to ProtoUTXOSet
func ReadOnlyUTXOSetToProtoUTXOSet(iter model.ReadOnlyUTXOSetIterator) (*ProtoUTXOSet, error) {
	protoUTXOSet := &ProtoUTXOSet{
		Utxos: []*ProtoUTXO{},
	}

	for iter.Next() {
		outpoint, entry, err := iter.Get()
		if err != nil {
			return nil, err
		}

		serializedUTXOBytes, err := utxo.SerializeUTXO(entry, outpoint)
		if err != nil {
			return nil, err
		}

		protoUTXOSet.Utxos = append(protoUTXOSet.Utxos, &ProtoUTXO{
			EntryOutpointPair: serializedUTXOBytes,
		})
	}
	return protoUTXOSet, nil
}
