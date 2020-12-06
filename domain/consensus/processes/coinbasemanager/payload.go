package coinbasemanager

import (
	"encoding/binary"
	"math"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

var byteOrder = binary.LittleEndian

const uint64Len = 8
const lengthOfscriptPubKeyLength = 1

// serializeCoinbasePayload builds the coinbase payload based on the provided scriptPubKey and extra data.
func (c coinbaseManager) serializeCoinbasePayload(blueScore uint64, coinbaseData *externalapi.DomainCoinbaseData) ([]byte, error) {
	scriptPubKeyLength := len(coinbaseData.ScriptPublicKey)
	if uint64(scriptPubKeyLength) > c.coinbasePayloadScriptPublicKeyMaxLength {
		return nil, errors.Wrapf(ruleerrors.ErrBadCoinbasePayloadLen, "coinbase's payload script public key is "+
			"longer than the max allowed length of %d", c.coinbasePayloadScriptPublicKeyMaxLength)
	}

	payload := make([]byte, uint64Len+lengthOfscriptPubKeyLength+scriptPubKeyLength+len(coinbaseData.ExtraData))
	byteOrder.PutUint64(payload[:uint64Len], blueScore)
	if len(coinbaseData.ScriptPublicKey) > math.MaxUint8 {
		return nil, errors.Errorf("script public key is bigger than %d", math.MaxUint8)
	}
	payload[uint64Len] = uint8(len(coinbaseData.ScriptPublicKey))
	copy(payload[uint64Len+lengthOfscriptPubKeyLength:], coinbaseData.ScriptPublicKey)
	copy(payload[uint64Len+lengthOfscriptPubKeyLength+scriptPubKeyLength:], coinbaseData.ExtraData)
	return payload, nil
}

// ExtractCoinbaseDataAndBlueScore deserializes the coinbase payload to its component (scriptPubKey and extra data).
func (c coinbaseManager) ExtractCoinbaseDataAndBlueScore(coinbaseTx *externalapi.DomainTransaction) (blueScore uint64,
	coinbaseData *externalapi.DomainCoinbaseData, err error) {

	minLength := uint64Len + lengthOfscriptPubKeyLength
	if len(coinbaseTx.Payload) < minLength {
		return 0, nil, errors.Wrapf(ruleerrors.ErrBadCoinbasePayloadLen,
			"coinbase payload is less than the minimum length of %d", minLength)
	}

	blueScore = byteOrder.Uint64(coinbaseTx.Payload[:uint64Len])
	scriptPubKeyLength := coinbaseTx.Payload[uint64Len]

	if uint64(scriptPubKeyLength) > c.coinbasePayloadScriptPublicKeyMaxLength {
		return 0, nil, errors.Wrapf(ruleerrors.ErrBadCoinbasePayloadLen, "coinbase's payload script public key is "+
			"longer than the max allowed length of %d", c.coinbasePayloadScriptPublicKeyMaxLength)
	}

	if len(coinbaseTx.Payload) < minLength+int(scriptPubKeyLength) {
		return 0, nil, errors.Wrapf(ruleerrors.ErrBadCoinbasePayloadLen,
			"coinbase payload doesn't have enough bytes to contain a script public key of %d bytes", scriptPubKeyLength)
	}

	return blueScore, &externalapi.DomainCoinbaseData{
		ScriptPublicKey: coinbaseTx.Payload[uint64Len+lengthOfscriptPubKeyLength : uint64Len+lengthOfscriptPubKeyLength+scriptPubKeyLength],
		ExtraData:       coinbaseTx.Payload[uint64Len+lengthOfscriptPubKeyLength+scriptPubKeyLength:],
	}, nil
}
