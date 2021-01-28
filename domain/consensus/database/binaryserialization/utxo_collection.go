package binaryserialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/util/binaryserializer"
	"io"
)

// SerializeUTXOCollection serializes the given utxoCollection into the given writer
func SerializeUTXOCollection(writer io.Writer, utxoCollection model.UTXOCollection) error {
	length := uint64(utxoCollection.Len())
	err := binaryserializer.PutUint64(writer, length)
	if err != nil {
		return err
	}

	utxoIterator := utxoCollection.Iterator()
	for ok := utxoIterator.First(); ok; ok = utxoIterator.Next() {
		outpoint, utxoEntry, err := utxoIterator.Get()
		if err != nil {
			return err
		}
		err = utxo.SerializeUTXOIntoWriter(writer, utxoEntry, outpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeserializeUTXOCollection deserializes a utxoCollection out of the given reader
func DeserializeUTXOCollection(reader io.Reader) (model.UTXOCollection, error) {
	length, err := binaryserializer.Uint64(reader)
	if err != nil {
		return nil, err
	}

	utxoMap := make(map[externalapi.DomainOutpoint]externalapi.UTXOEntry, length)
	for i := uint64(0); i < length; i++ {
		utxoEntry, outpoint, err := utxo.DeserializeUTXOOutOfReader(reader)
		if err != nil {
			return nil, err
		}
		utxoMap[*outpoint] = utxoEntry
	}

	utxoCollection := utxo.NewUTXOCollection(utxoMap)
	return utxoCollection, nil
}
