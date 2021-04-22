package bip32

import (
	"encoding/binary"
	"github.com/btcsuite/btcutil/base58"
	"github.com/kaspanet/go-secp256k1"
	"github.com/pkg/errors"
)

const (
	versionSerializationLen     = 4
	depthSerializationLen       = 1
	fingerprintSerializationLen = 4
	childNumberSerializationLen = 4
	chainCodeSerializationLen   = 32
	keySerializationLen         = 33
	checkSumLen                 = 4
)

const extendedKeySerializationLen = versionSerializationLen +
	depthSerializationLen +
	fingerprintSerializationLen +
	childNumberSerializationLen +
	chainCodeSerializationLen +
	keySerializationLen +
	checkSumLen

func DeserializeExtendedPrivateKey(extKeyString string) (*ExtendedKey, error) {
	serializedBytes := base58.Decode(extKeyString)
	return deserializeExtendedPrivateKey(serializedBytes)
}

func deserializeExtendedPrivateKey(serialized []byte) (*ExtendedKey, error) {
	if len(serialized) != extendedKeySerializationLen {
		return nil, errors.Errorf("key length must be %d bytes but got %d", extendedKeySerializationLen, len(serialized))
	}

	err := validateChecksum(serialized)
	if err != nil {
		return nil, err
	}

	extKey := &ExtendedKey{}

	copy(extKey.Version[:], serialized[:versionSerializationLen])
	extKey.Depth = serialized[versionSerializationLen]
	copy(extKey.Fingerprint[:], serialized[versionSerializationLen+depthSerializationLen:])
	extKey.ChildNumber = binary.BigEndian.Uint32(
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen:],
	)
	copy(
		extKey.ChainCode[:],
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen:],
	)

	isPrivate := isPrivateVersion(extKey.Version)
	if isPrivate {
		privateKeyPadding := serialized[versionSerializationLen+
			depthSerializationLen+
			fingerprintSerializationLen+
			childNumberSerializationLen+
			chainCodeSerializationLen]
		if privateKeyPadding != 0 {
			return nil, errors.Errorf("expected 0 padding for private key but got %d", privateKeyPadding)
		}

		extKey.privateKey, err = secp256k1.DeserializeECDSAPrivateKeyFromSlice(serialized[versionSerializationLen+
			depthSerializationLen+
			fingerprintSerializationLen+
			childNumberSerializationLen+
			chainCodeSerializationLen+1 : versionSerializationLen+
			depthSerializationLen+
			fingerprintSerializationLen+
			childNumberSerializationLen+
			chainCodeSerializationLen+
			keySerializationLen])
		if err != nil {
			return nil, err
		}
	} else {
		extKey.publicKey, err = secp256k1.DeserializeECDSAPubKey(serialized[versionSerializationLen+
			depthSerializationLen+
			fingerprintSerializationLen+
			childNumberSerializationLen+
			chainCodeSerializationLen : versionSerializationLen+
			depthSerializationLen+
			fingerprintSerializationLen+
			childNumberSerializationLen+
			chainCodeSerializationLen+
			keySerializationLen])
		if err != nil {
			return nil, err
		}
	}

	return extKey, nil
}

func (extKey *ExtendedKey) serialize() ([]byte, error) {
	var serialized [extendedKeySerializationLen]byte
	copy(serialized[:versionSerializationLen], extKey.Version[:])
	serialized[versionSerializationLen] = extKey.Depth
	copy(serialized[versionSerializationLen+depthSerializationLen:], extKey.Fingerprint[:])
	binary.BigEndian.PutUint32(
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen:],
		extKey.ChildNumber,
	)
	copy(
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen:],
		extKey.ChainCode[:],
	)

	if extKey.IsPrivate() {
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen+chainCodeSerializationLen] = 0
		copy(
			serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen+chainCodeSerializationLen+1:],
			extKey.privateKey.Serialize()[:],
		)
	} else {
		publicKey, err := extKey.PublicKey()
		if err != nil {
			return nil, err
		}

		serializedPublicKey, err := publicKey.Serialize()
		if err != nil {
			return nil, err
		}

		copy(
			serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen+chainCodeSerializationLen:],
			serializedPublicKey[:],
		)
	}

	checkSum := doubleSha256(serialized[:len(serialized)-checkSumLen])
	copy(
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen+chainCodeSerializationLen+keySerializationLen:],
		checkSum,
	)
	return serialized[:], nil
}
