package utxo

import (
	"bytes"
	"io"

	"github.com/kaspanet/kaspad/domain/consensus/utils/serialization"

	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

const uint32Size = 4

// SerializeUTXO returns the byte-slice representation for given UTXOEntry-outpoint pair
func SerializeUTXO(entry *externalapi.UTXOEntry, outpoint *externalapi.DomainOutpoint) ([]byte, error) {
	w := &bytes.Buffer{}

	err := serializeOutpoint(w, outpoint)
	if err != nil {
		return nil, err
	}

	err = serializeUTXOEntry(w, entry)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// DeserializeUTXO deserializes the given byte slice to UTXOEntry-outpoint pair
func DeserializeUTXO(utxoBytes []byte) (entry *externalapi.UTXOEntry, outpoint *externalapi.DomainOutpoint, err error) {
	r := bytes.NewReader(utxoBytes)
	outpoint, err = deserializeOutpoint(r)
	if err != nil {
		return nil, nil, err
	}

	entry, err = deserializeUTXOEntry(r)
	if err != nil {
		return nil, nil, err
	}

	return entry, outpoint, nil
}

func serializeOutpoint(w io.Writer, outpoint *externalapi.DomainOutpoint) error {
	_, err := w.Write(outpoint.TransactionID[:])
	if err != nil {
		return err
	}

	err = serialization.WriteElement(w, outpoint.Index)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func deserializeOutpoint(r io.Reader) (*externalapi.DomainOutpoint, error) {
	transactionIDBytes := make([]byte, externalapi.DomainHashSize)
	_, err := io.ReadFull(r, transactionIDBytes)
	if err != nil {
		return nil, err
	}

	transactionID, err := transactionid.FromBytes(transactionIDBytes)
	if err != nil {
		return nil, err
	}

	indexBytes := make([]byte, uint32Size)
	_, err = io.ReadFull(r, indexBytes)
	if err != nil {
		return nil, err
	}

	var index uint32
	err = serialization.ReadElement(r, &index)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainOutpoint{
		TransactionID: *transactionID,
		Index:         index,
	}, nil
}

func serializeUTXOEntry(w io.Writer, entry *externalapi.UTXOEntry) error {
	err := serialization.WriteElements(w, entry.BlockBlueScore, entry.Amount, entry.IsCoinbase)
	if err != nil {
		return err
	}

	count := uint64(len(entry.ScriptPublicKey))
	err = serialization.WriteElement(w, count)
	if err != nil {
		return err
	}

	_, err = w.Write(entry.ScriptPublicKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func deserializeUTXOEntry(r io.Reader) (*externalapi.UTXOEntry, error) {
	entry := &externalapi.UTXOEntry{}
	err := serialization.ReadElements(r, entry.BlockBlueScore, entry.Amount, entry.IsCoinbase)
	if err != nil {
		return nil, err
	}

	count := uint64(len(entry.ScriptPublicKey))
	err = serialization.ReadElement(r, count)
	if err != nil {
		return nil, err
	}

	_, err = r.Read(entry.ScriptPublicKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return entry, nil
}
