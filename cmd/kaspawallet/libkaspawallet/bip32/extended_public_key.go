package bip32

import (
	"encoding/binary"
	"github.com/btcsuite/btcutil/base58"
	"github.com/kaspanet/go-secp256k1"
	"github.com/pkg/errors"
)

type ExtendedPublicKey struct {
	PublicKey *secp256k1.ECDSAPublicKey
	*extendedKey
}

func (extPub *ExtendedPublicKey) Path(pathString string) (*ExtendedPublicKey, error) {
	path, err := parsePath(pathString)
	if err != nil {
		return nil, err
	}

	return extPub.path(path)
}

func (extPub *ExtendedPublicKey) path(path *path) (*ExtendedPublicKey, error) {
	if path.isPrivate {
		return nil, errors.Errorf("path() cannot handle a private path")
	}

	descendantExtKey := extPub
	for _, index := range path.indexes {
		var err error
		descendantExtKey, err = descendantExtKey.Child(index)
		if err != nil {
			return nil, err
		}
	}

	return descendantExtKey, nil
}

func (extPub *ExtendedPublicKey) serialize() ([]byte, error) {
	var serialized [extendedKeySerializationLen]byte
	copy(serialized[:versionSerializationLen], extPub.Version[:])
	serialized[versionSerializationLen] = extPub.Depth
	copy(serialized[versionSerializationLen+depthSerializationLen:], extPub.Fingerprint[:])
	binary.BigEndian.PutUint32(
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen:],
		extPub.ChildNumber,
	)
	copy(
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen:],
		extPub.ChainCode[:],
	)

	serializedPoint, err := extPub.PublicKey.Serialize()
	if err != nil {
		return nil, err
	}

	copy(
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen+chainCodeSerializationLen:],
		serializedPoint[:],
	)
	checkSum := doubleSha256(serialized[:len(serialized)-checkSumLen])
	copy(
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen+chainCodeSerializationLen+keySerializationLen:],
		checkSum,
	)
	return serialized[:], nil
}

func DeserializeExtendedPublicKey(extPubString string) (*ExtendedPublicKey, error) {
	serializedBytes := base58.Decode(extPubString)
	return deserializeExtendedPublicKey(serializedBytes)
}

func deserializeExtendedPublicKey(serialized []byte) (*ExtendedPublicKey, error) {
	if len(serialized) != extendedKeySerializationLen {
		return nil, errors.Errorf("key length must be %d bytes but got %d", extendedKeySerializationLen, len(serialized))
	}

	err := validateChecksum(serialized)
	if err != nil {
		return nil, err
	}

	extPub := &ExtendedPublicKey{
		PublicKey:   nil,
		extendedKey: &extendedKey{},
	}

	copy(extPub.Version[:], serialized[:versionSerializationLen])
	extPub.Depth = serialized[versionSerializationLen]
	copy(extPub.Fingerprint[:], serialized[versionSerializationLen+depthSerializationLen:])
	extPub.ChildNumber = binary.BigEndian.Uint32(
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen:],
	)
	copy(
		extPub.ChainCode[:],
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen:],
	)

	extPub.PublicKey, err = secp256k1.DeserializeECDSAPubKey(serialized[versionSerializationLen+
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

	return extPub, nil
}

func (extPub *ExtendedPublicKey) String() string {
	serialized, err := extPub.serialize()
	if err != nil {
		panic(errors.Wrap(err, "error serializing key"))
	}
	return base58.Encode(serialized)
}
