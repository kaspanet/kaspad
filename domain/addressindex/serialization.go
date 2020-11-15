package addressindex

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"io"
)

// DeserializeUTXOCollection deserializes the UTXO collection from the provided reader
func DeserializeUTXOCollection(r io.Reader) (UTXOCollection, error) {
	count, err := appmessage.ReadVarInt(r)
	if err != nil {
		return nil, err
	}
	collection := UTXOCollection{}
	for i := uint64(0); i < count; i++ {
		outpoint, err := consensusserialization.DeserializeOutpoint(r)
		if err != nil {
			return nil, err
		}
		utxoEntry, err := consensusserialization.DeserializeUTXOEntry(r)
		if err != nil {
			return nil, err
		}

		collection.Add(outpoint, utxoEntry)
	}
	return collection, nil
}

// SerializeUTXOCollection serializes the provided UTXO collection into the provided writer
func SerializeUTXOCollection(w io.Writer, collection UTXOCollection) error {
	err := appmessage.WriteVarInt(w, uint64(len(collection)))
	if err != nil {
		return err
	}
	for outpoint, utxoEntry := range collection {
		err := consensusserialization.SerializeOutpoint(w, &outpoint)
		if err != nil {
			return err
		}

		err = consensusserialization.SerializeUTXOEntry(w, utxoEntry)
		if err != nil {
			return err
		}
	}
	return nil
}
