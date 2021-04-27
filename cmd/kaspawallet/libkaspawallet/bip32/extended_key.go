package bip32

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/kaspanet/go-secp256k1"
	"github.com/pkg/errors"
)

// ExtendedKey is a bip32 extended key
type ExtendedKey struct {
	privateKey        *secp256k1.ECDSAPrivateKey
	publicKey         *secp256k1.ECDSAPublicKey
	Version           [4]byte
	Depth             uint8
	ParentFingerprint [4]byte
	ChildNumber       uint32
	ChainCode         [32]byte
}

// PrivateKey returns the ECDSA private key associated with the extended key
func (extKey *ExtendedKey) PrivateKey() *secp256k1.ECDSAPrivateKey {
	return extKey.privateKey
}

// PublicKey returns the ECDSA public key associated with the extended key
func (extKey *ExtendedKey) PublicKey() (*secp256k1.ECDSAPublicKey, error) {
	if extKey.publicKey != nil {
		return extKey.publicKey, nil
	}

	publicKey, err := extKey.privateKey.ECDSAPublicKey()
	if err != nil {
		return nil, err
	}

	extKey.publicKey = publicKey
	return publicKey, nil
}

// IsPrivate returns whether the extended key is private
func (extKey *ExtendedKey) IsPrivate() bool {
	return extKey.privateKey != nil
}

// Public returns public version of the extended key
func (extKey *ExtendedKey) Public() (*ExtendedKey, error) {
	if !extKey.IsPrivate() {
		return extKey, nil
	}

	publicKey, err := extKey.PublicKey()
	if err != nil {
		return nil, errors.Wrap(err, "error calculating publicKey")
	}

	version, err := toPublicVersion(extKey.Version)
	if err != nil {
		return nil, err
	}

	return &ExtendedKey{
		publicKey:         publicKey,
		Version:           version,
		Depth:             extKey.Depth,
		ParentFingerprint: extKey.ParentFingerprint,
		ChildNumber:       extKey.ChildNumber,
		ChainCode:         extKey.ChainCode,
	}, nil
}

// DeriveFromPath returns the extended key derived from the given path
func (extKey *ExtendedKey) DeriveFromPath(pathString string) (*ExtendedKey, error) {
	path, err := parsePath(pathString)
	if err != nil {
		return nil, err
	}

	return extKey.path(path)
}

func (extKey *ExtendedKey) path(path *path) (*ExtendedKey, error) {
	if path.isPrivate && !extKey.IsPrivate() {
		return nil, errors.Errorf("extended public keys cannot handle a private path")
	}

	descendantExtKey := extKey
	for _, index := range path.indexes {
		var err error
		descendantExtKey, err = descendantExtKey.Child(index)
		if err != nil {
			return nil, err
		}
	}

	if !path.isPrivate {
		return descendantExtKey.Public()
	}

	return descendantExtKey, nil
}

func (extKey *ExtendedKey) String() string {
	serialized, err := extKey.serialize()
	if err != nil {
		panic(errors.Wrap(err, "error serializing key"))
	}
	return base58.Encode(serialized)
}
