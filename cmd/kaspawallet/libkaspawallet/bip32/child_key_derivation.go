package bip32

import (
	"encoding/binary"
	"github.com/kaspanet/go-secp256k1"
	"github.com/pkg/errors"
)

const hardenedIndexStart = 0x80000000

func NewMaster(seed []byte, version [4]byte) (*ExtendedKey, error) {
	mac := newHMACWriter([]byte("Bitcoin seed"))
	mac.InfallibleWrite(seed)
	I := mac.Sum(nil)

	var iL, iR [32]byte
	copy(iL[:], I[:32])
	copy(iR[:], I[32:])

	privateKey, err := secp256k1.DeserializeECDSAPrivateKeyFromSlice(iL[:])
	if err != nil {
		return nil, err
	}

	return &ExtendedKey{
		privateKey:        privateKey,
		Version:           version,
		Depth:             0,
		ParentFingerprint: [4]byte{},
		ChildNumber:       0,
		ChainCode:         iR,
	}, nil
}

func isHardened(i uint32) bool {
	return i >= hardenedIndexStart
}

func (extKey *ExtendedKey) Child(i uint32) (*ExtendedKey, error) {
	I, err := extKey.calcI(i)
	if err != nil {
		return nil, err
	}

	var iL, iR [32]byte
	copy(iL[:], I[:32])
	copy(iR[:], I[32:])

	fingerPrint, err := extKey.calcFingerprint()
	if err != nil {
		return nil, err
	}

	childExt := &ExtendedKey{
		Version:           extKey.Version,
		Depth:             extKey.Depth + 1,
		ParentFingerprint: fingerPrint,
		ChildNumber:       i,
		ChainCode:         iR,
	}

	if extKey.privateKey != nil {
		childExt.privateKey, err = privateKeyAdd(extKey.privateKey, iL)
		if err != nil {
			return nil, err
		}
	} else {
		publicKey, err := extKey.PublicKey()
		if err != nil {
			return nil, err
		}

		childExt.publicKey, err = pointAdd(publicKey, iL)
		if err != nil {
			return nil, err
		}
	}

	return childExt, nil
}

func (extKey *ExtendedKey) calcFingerprint() ([4]byte, error) {
	publicKey, err := extKey.PublicKey()
	if err != nil {
		return [4]byte{}, err
	}

	serializedPoint, err := publicKey.Serialize()
	if err != nil {
		return [4]byte{}, err
	}

	hash := hash160(serializedPoint[:])
	var fingerprint [4]byte
	copy(fingerprint[:], hash[:4])
	return fingerprint, nil
}

func privateKeyAdd(k *secp256k1.ECDSAPrivateKey, tweak [32]byte) (*secp256k1.ECDSAPrivateKey, error) {
	kCopy := *k
	err := kCopy.Add(tweak)
	if err != nil {
		return nil, err
	}

	return &kCopy, nil
}

func (extKey *ExtendedKey) calcI(i uint32) ([]byte, error) {
	if isHardened(i) && !extKey.IsPrivate() {
		return nil, errors.Errorf("Cannot calculate hardened child for public key")
	}

	mac := newHMACWriter(extKey.ChainCode[:])
	if isHardened(i) {
		mac.InfallibleWrite([]byte{0x00})
		mac.InfallibleWrite(extKey.privateKey.Serialize()[:])
	} else {
		publicKey, err := extKey.PublicKey()
		if err != nil {
			return nil, err
		}

		serializedPublicKey, err := publicKey.Serialize()
		if err != nil {
			return nil, errors.Wrap(err, "error serializing public key")
		}

		mac.InfallibleWrite(serializedPublicKey[:])
	}

	mac.InfallibleWrite(serializeUint32(i))
	return mac.Sum(nil), nil
}

func serializeUint32(v uint32) []byte {
	serialized := make([]byte, 4)
	binary.BigEndian.PutUint32(serialized, v)
	return serialized
}

func pointAdd(point *secp256k1.ECDSAPublicKey, tweak [32]byte) (*secp256k1.ECDSAPublicKey, error) {
	pointCopy := *point
	err := pointCopy.Add(tweak)
	if err != nil {
		return nil, err
	}

	return &pointCopy, nil
}
