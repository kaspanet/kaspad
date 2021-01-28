package binaryserialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"io"
)

// SerializeUTXODiff serializes the given utxoDiff into the given writer
func SerializeUTXODiff(writer io.Writer, utxoDiff model.UTXODiff) error {
	err := SerializeUTXOCollection(writer, utxoDiff.ToAdd())
	if err != nil {
		return err
	}
	return SerializeUTXOCollection(writer, utxoDiff.ToRemove())
}

// DeserializeUTXODiff deserializes a utxoDiff out of the given reader
func DeserializeUTXODiff(reader io.Reader) (model.UTXODiff, error) {
	toAdd, err := DeserializeUTXOCollection(reader)
	if err != nil {
		return nil, err
	}
	toRemove, err := DeserializeUTXOCollection(reader)
	if err != nil {
		return nil, err
	}
	return utxo.NewUTXODiffFromCollections(toAdd, toRemove)
}
