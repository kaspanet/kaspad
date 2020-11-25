package consensusserialization

import (
	"bytes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"
	"io"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

const uint32Size = 4

// SerializeUTXO returns the byte-slice representation for given UTXOEntry-outpoint pair
func SerializeUTXO(entry *externalapi.UTXOEntry, outpoint *externalapi.DomainOutpoint) ([]byte, error) {
	w := &bytes.Buffer{}

	err := SerializeOutpoint(w, outpoint)
	if err != nil {
		return nil, err
	}

	err = SerializeUTXOEntry(w, entry)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// DeserializeUTXO deserializes the given byte slice to UTXOEntry-outpoint pair
func DeserializeUTXO(utxoBytes []byte) (entry *externalapi.UTXOEntry, outpoint *externalapi.DomainOutpoint, err error) {
	r := bytes.NewReader(utxoBytes)
	outpoint, err = DeserializeOutpoint(r)
	if err != nil {
		return nil, nil, err
	}

	entry, err = DeserializeUTXOEntry(r)
	if err != nil {
		return nil, nil, err
	}

	return entry, outpoint, nil
}

// SerializeOutpoint serializes provided outpoint into the writer
func SerializeOutpoint(w io.Writer, outpoint *externalapi.DomainOutpoint) error {
	_, err := w.Write(outpoint.TransactionID[:])
	if err != nil {
		return err
	}

	err = WriteElement(w, outpoint.Index)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// DeserializeOutpoint deserializes the outpoint from the provided reader
func DeserializeOutpoint(r io.Reader) (*externalapi.DomainOutpoint, error) {
	outpoint := &externalapi.DomainOutpoint{}
	transactionIDBytes := make([]byte, externalapi.DomainHashSize)
	_, err := io.ReadFull(r, transactionIDBytes)
	if err != nil {
		return nil, err
	}

	transactionID, err := transactionid.FromBytes(transactionIDBytes)
	if err != nil {
		return nil, err
	}
	outpoint.TransactionID = *transactionID

	err = readElement(r, &outpoint.Index)
	if err != nil {
		return nil, err
	}

	return outpoint, nil
}

// SerializeUTXOEntry serializes the provided UTXO entry into the provided writer
func SerializeUTXOEntry(w io.Writer, entry *externalapi.UTXOEntry) error {
	err := writeElements(w, entry.BlockBlueScore, entry.Amount, entry.IsCoinbase)
	if err != nil {
		return err
	}

	count := uint64(len(entry.ScriptPublicKey))
	err = WriteElement(w, count)
	if err != nil {
		return err
	}

	_, err = w.Write(entry.ScriptPublicKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// DeserializeUTXOEntry deserializes the UTXO entry from the provided reader
func DeserializeUTXOEntry(r io.Reader) (*externalapi.UTXOEntry, error) {
	entry := &externalapi.UTXOEntry{}
	err := readElements(r, &entry.BlockBlueScore, &entry.Amount, &entry.IsCoinbase)
	if err != nil {
		return nil, err
	}

	var scriptPublicKeyLen uint64
	err = readElement(r, &scriptPublicKeyLen)
	if err != nil {
		return nil, err
	}

	entry.ScriptPublicKey = make([]byte, scriptPublicKeyLen)
	_, err = r.Read(entry.ScriptPublicKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return entry, nil
}
