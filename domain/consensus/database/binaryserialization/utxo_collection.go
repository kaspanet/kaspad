package binaryserialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/util/binaryserializer"
	"io"
)

func SerializeUTXOCollection(writer io.Writer, utxoCollection model.UTXOCollection) error {
	length := uint64(utxoCollection.Len())
	err := binaryserializer.PutUint64(writer, length)
	if err != nil {
		return err
	}
	return nil
}

func DeserializeUTXOCollection(reader io.Reader) (model.UTXOCollection, error) {
	length, err := binaryserializer.Uint64(reader)
	if err != nil {
		return nil, err
	}
	utxoMap := make(map[externalapi.DomainOutpoint]externalapi.UTXOEntry, length)
	utxoCollection := utxo.NewUTXOCollection(utxoMap)
	return utxoCollection, nil
}
