package coinbasemanager

import (
	"encoding/binary"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

const uint64Len = 8
const uint16Len = 2
const lengthOfSubsidy = uint64Len
const lengthOfScriptPubKeyLength = 1
const lengthOfVersionScriptPubKey = uint16Len

// serializeCoinbasePayload builds the coinbase payload based on the provided scriptPubKey and extra data.
func (c *coinbaseManager) serializeCoinbasePayload(blueScore uint64,
	coinbaseData *externalapi.DomainCoinbaseData, subsidy uint64) ([]byte, error) {

	scriptLengthOfScriptPubKey := len(coinbaseData.ScriptPublicKey.Script)
	if scriptLengthOfScriptPubKey > int(c.coinbasePayloadScriptPublicKeyMaxLength) {
		return nil, errors.Wrapf(ruleerrors.ErrBadCoinbasePayloadLen, "coinbase's payload script public key is "+
			"longer than the max allowed length of %d", c.coinbasePayloadScriptPublicKeyMaxLength)
	}

	payload := make([]byte, uint64Len+lengthOfVersionScriptPubKey+lengthOfScriptPubKeyLength+scriptLengthOfScriptPubKey+len(coinbaseData.ExtraData)+lengthOfSubsidy)
	binary.LittleEndian.PutUint64(payload[:uint64Len], blueScore)
	binary.LittleEndian.PutUint64(payload[uint64Len:], subsidy)

	payload[uint64Len+lengthOfSubsidy] = uint8(coinbaseData.ScriptPublicKey.Version)
	payload[uint64Len+lengthOfSubsidy+lengthOfVersionScriptPubKey] = uint8(len(coinbaseData.ScriptPublicKey.Script))
	copy(payload[uint64Len+lengthOfSubsidy+lengthOfVersionScriptPubKey+lengthOfScriptPubKeyLength:], coinbaseData.ScriptPublicKey.Script)
	copy(payload[uint64Len+lengthOfSubsidy+lengthOfVersionScriptPubKey+lengthOfScriptPubKeyLength+scriptLengthOfScriptPubKey:], coinbaseData.ExtraData)

	return payload, nil
}

// ExtractCoinbaseDataBlueScoreAndSubsidy deserializes the coinbase payload to its component (scriptPubKey, extra data, and subsidy).
func (c *coinbaseManager) ExtractCoinbaseDataBlueScoreAndSubsidy(coinbaseTx *externalapi.DomainTransaction) (
	blueScore uint64, coinbaseData *externalapi.DomainCoinbaseData, subsidy uint64, err error) {

	minLength := uint64Len + lengthOfSubsidy + lengthOfVersionScriptPubKey + lengthOfScriptPubKeyLength
	if len(coinbaseTx.Payload) < minLength {
		return 0, nil, 0, errors.Wrapf(ruleerrors.ErrBadCoinbasePayloadLen,
			"coinbase payload is less than the minimum length of %d", minLength)
	}

	blueScore = binary.LittleEndian.Uint64(coinbaseTx.Payload[:uint64Len])
	subsidy = binary.LittleEndian.Uint64(coinbaseTx.Payload[uint64Len:])

	scriptPubKeyVersion := uint16(coinbaseTx.Payload[uint64Len+lengthOfSubsidy])
	scriptPubKeyScriptLength := coinbaseTx.Payload[uint64Len+lengthOfSubsidy+lengthOfVersionScriptPubKey]

	if scriptPubKeyScriptLength > c.coinbasePayloadScriptPublicKeyMaxLength {
		return 0, nil, 0, errors.Wrapf(ruleerrors.ErrBadCoinbasePayloadLen, "coinbase's payload script public key is "+
			"longer than the max allowed length of %d", c.coinbasePayloadScriptPublicKeyMaxLength)
	}

	if len(coinbaseTx.Payload) < minLength+int(scriptPubKeyScriptLength) {
		return 0, nil, 0, errors.Wrapf(ruleerrors.ErrBadCoinbasePayloadLen,
			"coinbase payload doesn't have enough bytes to contain a script public key of %d bytes", scriptPubKeyScriptLength)
	}
	scriptPubKeyScript := coinbaseTx.Payload[uint64Len+lengthOfSubsidy+lengthOfVersionScriptPubKey+lengthOfScriptPubKeyLength : uint64Len+lengthOfSubsidy+lengthOfVersionScriptPubKey+lengthOfScriptPubKeyLength+scriptPubKeyScriptLength]

	return blueScore, &externalapi.DomainCoinbaseData{
		ScriptPublicKey: &externalapi.ScriptPublicKey{Script: scriptPubKeyScript, Version: scriptPubKeyVersion},
		ExtraData:       coinbaseTx.Payload[uint64Len+lengthOfSubsidy+lengthOfVersionScriptPubKey+lengthOfScriptPubKeyLength+scriptPubKeyScriptLength:],
	}, subsidy, nil
}
