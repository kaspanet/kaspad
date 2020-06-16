package coinbasepayload

import (
	"bytes"
	"encoding/binary"
	"github.com/kaspanet/kaspad/util/binaryserializer"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

var byteOrder = binary.LittleEndian

// SerializeCoinbasePayload builds the coinbase payload based on the provided scriptPubKey and extra data.
func SerializeCoinbasePayload(blueScore uint64, scriptPubKey []byte, extraData []byte) ([]byte, error) {
	w := &bytes.Buffer{}
	err := binaryserializer.PutUint64(w, byteOrder, blueScore)
	if err != nil {
		return nil, err
	}
	err = wire.WriteVarInt(w, uint64(len(scriptPubKey)))
	if err != nil {
		return nil, err
	}
	_, err = w.Write(scriptPubKey)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(extraData)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// ErrIncorrectScriptPubKeyLen indicates that the script pub key length is not as expected.
var ErrIncorrectScriptPubKeyLen = errors.New("incorrect script pub key length")

// DeserializeCoinbasePayload deserializes the coinbase payload to its component (scriptPubKey and extra data).
func DeserializeCoinbasePayload(tx *wire.MsgTx) (blueScore uint64, scriptPubKey []byte, extraData []byte, err error) {
	r := bytes.NewReader(tx.Payload)
	blueScore, err = binaryserializer.Uint64(r, byteOrder)
	if err != nil {
		return 0, nil, nil, err
	}
	scriptPubKeyLen, err := wire.ReadVarInt(r)
	if err != nil {
		return 0, nil, nil, err
	}
	scriptPubKey = make([]byte, scriptPubKeyLen)
	n, err := r.Read(scriptPubKey)
	if err != nil {
		return 0, nil, nil, err
	}
	if uint64(n) != scriptPubKeyLen {
		return 0, nil, nil,
			errors.Wrapf(ErrIncorrectScriptPubKeyLen, "expected %d bytes in script pub key but got %d", scriptPubKeyLen, n)
	}
	extraData = make([]byte, r.Len())
	if r.Len() != 0 {
		_, err = r.Read(extraData)
		if err != nil {
			return 0, nil, nil, err
		}
	}
	return blueScore, scriptPubKey, extraData, nil
}
