package bip32

import (
	"encoding/binary"
	"github.com/btcsuite/btcutil/base58"
	"github.com/kaspanet/go-secp256k1"
	"github.com/pkg/errors"
)

type ExtendedPrivateKey struct {
	PrivateKey *secp256k1.ECDSAPrivateKey
	*extendedKey
}

func (extPrv *ExtendedPrivateKey) Public() (*ExtendedPublicKey, error) {
	point, err := extPrv.PrivateKey.ECDSAPublicKey()
	if err != nil {
		return nil, errors.Wrap(err, "error calculating point")
	}

	version, err := toPublicVersion(extPrv.Version)
	if err != nil {
		return nil, err
	}

	return &ExtendedPublicKey{
		PublicKey: point,
		extendedKey: &extendedKey{
			Version:     version,
			Depth:       extPrv.Depth,
			Fingerprint: extPrv.Fingerprint,
			ChildNumber: extPrv.ChildNumber,
			ChainCode:   extPrv.ChainCode,
		},
	}, nil
}

func (extPrv *ExtendedPrivateKey) Path(pathString string) (*ExtendedPrivateKey, error) {
	path, err := parsePath(pathString)
	if err != nil {
		return nil, err
	}

	return extPrv.path(path)
}

func (extPrv *ExtendedPrivateKey) path(path *path) (*ExtendedPrivateKey, error) {
	if !path.isPrivate {
		return nil, errors.Errorf("path() cannot handle a public path")
	}

	descendantExtKey := extPrv
	for _, index := range path.indexes {
		var err error
		descendantExtKey, err = descendantExtKey.Child(index)
		if err != nil {
			return nil, err
		}
	}

	return descendantExtKey, nil
}

func (extPrv *ExtendedPrivateKey) pathPublic(path *path) (*ExtendedPublicKey, error) {
	if path.isPrivate {
		return nil, errors.Errorf("pathPublic() cannot handle a private path")
	}

	pathPrivate := *path
	pathPrivate.isPrivate = true

	descendantExtKey, err := extPrv.path(&pathPrivate)
	if err != nil {
		return nil, err
	}

	return descendantExtKey.Public()
}

func (extPrv *ExtendedPrivateKey) serialize() []byte {
	var serialized [extendedKeySerializationLen]byte
	copy(serialized[:versionSerializationLen], extPrv.Version[:])
	serialized[versionSerializationLen] = extPrv.Depth
	copy(serialized[versionSerializationLen+depthSerializationLen:], extPrv.Fingerprint[:])
	binary.BigEndian.PutUint32(
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen:],
		extPrv.ChildNumber,
	)
	copy(
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen:],
		extPrv.ChainCode[:],
	)
	serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen+chainCodeSerializationLen] = 0
	copy(
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen+chainCodeSerializationLen+1:],
		extPrv.PrivateKey.Serialize()[:],
	)
	checkSum := doubleSha256(serialized[:len(serialized)-checkSumLen])
	copy(
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen+chainCodeSerializationLen+keySerializationLen:],
		checkSum,
	)
	return serialized[:]
}

func (extPrv *ExtendedPrivateKey) String() string {
	serialized := extPrv.serialize()
	return base58.Encode(serialized)
}

func DeserializeExtendedPrivateKey(extPrvString string) (*ExtendedPrivateKey, error) {
	serializedBytes := base58.Decode(extPrvString)
	return deserializeExtendedPrivateKey(serializedBytes)
}

func deserializeExtendedPrivateKey(serialized []byte) (*ExtendedPrivateKey, error) {
	if len(serialized) != extendedKeySerializationLen {
		return nil, errors.Errorf("key length must be %d bytes but got %d", extendedKeySerializationLen, len(serialized))
	}

	err := validateChecksum(serialized)
	if err != nil {
		return nil, err
	}

	extPrv := &ExtendedPrivateKey{
		PrivateKey:  nil,
		extendedKey: &extendedKey{},
	}

	copy(extPrv.Version[:], serialized[:versionSerializationLen])
	extPrv.Depth = serialized[versionSerializationLen]
	copy(extPrv.Fingerprint[:], serialized[versionSerializationLen+depthSerializationLen:])
	extPrv.ChildNumber = binary.BigEndian.Uint32(
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen:],
	)
	copy(
		extPrv.ChainCode[:],
		serialized[versionSerializationLen+depthSerializationLen+fingerprintSerializationLen+childNumberSerializationLen:],
	)

	privateKeyPadding := serialized[versionSerializationLen+
		depthSerializationLen+
		fingerprintSerializationLen+
		childNumberSerializationLen+
		chainCodeSerializationLen]
	if privateKeyPadding != 0 {
		return nil, errors.Errorf("expected 0 padding for private key but got %d", privateKeyPadding)
	}

	extPrv.PrivateKey, err = secp256k1.DeserializeECDSAPrivateKeyFromSlice(serialized[versionSerializationLen+
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

	return extPrv, nil
}
