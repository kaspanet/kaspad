package bip32

import (
	"crypto/sha256"
	"encoding/binary"
	"github.com/kaspanet/go-secp256k1"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ripemd160"
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
		PrivateKey:  privateKey,
		Version:     version,
		Depth:       0,
		Fingerprint: [4]byte{},
		ChildNumber: 0,
		ChainCode:   iR,
	}, nil
}

func isHardened(i uint32) bool {
	return i >= hardenedIndexStart
}

func (extKey *ExtendedKey) Child(i uint32) (*ExtendedKey, error) {
	I, err := calcI(extKey, i)
	if err != nil {
		return nil, err
	}

	var iL, iR [32]byte
	copy(iL[:], I[:32])
	copy(iR[:], I[32:])

	publicKey, err := extKey.PublicKey()
	if err != nil {
		return nil, err
	}

	fingerPrint, err := fingerPrintFromPoint(publicKey)
	if err != nil {
		return nil, err
	}

	childExt := &ExtendedKey{
		Version:     extKey.Version,
		Depth:       extKey.Depth + 1,
		Fingerprint: fingerPrint,
		ChildNumber: i,
		ChainCode:   iR,
	}

	if extKey.PrivateKey != nil {
		childExt.PrivateKey, err = privateKeyAdd(extKey.PrivateKey, iL)
		if err != nil {
			return nil, err
		}
	} else {
		childExt.publicKey, err = pointAdd(publicKey, iL)
		if err != nil {
			return nil, err
		}
	}

	return childExt, nil
}

func fingerprintFromPrivateKey(privateKey *secp256k1.ECDSAPrivateKey) ([4]byte, error) {
	point, err := privateKey.ECDSAPublicKey()
	if err != nil {
		return [4]byte{}, err
	}

	return fingerPrintFromPoint(point)
}

func fingerPrintFromPoint(point *secp256k1.ECDSAPublicKey) ([4]byte, error) {
	serializedPoint, err := point.Serialize()
	if err != nil {
		return [4]byte{}, err
	}

	hash := hash160(serializedPoint[:])
	var fingerprint [4]byte
	copy(fingerprint[:], hash[:4])
	return fingerprint, nil
}

func hash160(data []byte) []byte {
	sha := sha256.New()
	ripe := ripemd160.New()
	sha.Write(data)
	ripe.Write(sha.Sum(nil))
	return ripe.Sum(nil)
}

func privateKeyAdd(k *secp256k1.ECDSAPrivateKey, tweak [32]byte) (*secp256k1.ECDSAPrivateKey, error) {
	kCopy := *k
	err := kCopy.Add(tweak)
	if err != nil {
		return nil, err
	}

	return &kCopy, nil
}

func calcI(extKey *ExtendedKey, i uint32) ([]byte, error) {
	if isHardened(i) {
		if !extKey.IsPrivate() {
			return nil, errors.Errorf("Cannot calculate hardened child for public key")
		}

		mac := newHMACWriter(extKey.ChainCode[:])
		mac.InfallibleWrite([]byte{0x00})
		mac.InfallibleWrite(extKey.PrivateKey.Serialize()[:])
		mac.InfallibleWrite(ser32(i))
		return mac.Sum(nil), nil
	}

	mac := newHMACWriter(extKey.ChainCode[:])
	publicKey, err := extKey.PublicKey()
	if err != nil {
		return nil, err
	}

	serializedPublicKey, err := publicKey.Serialize()
	if err != nil {
		return nil, errors.Wrap(err, "error serializing public key")
	}

	mac.InfallibleWrite(serializedPublicKey[:])
	mac.InfallibleWrite(ser32(i))
	return mac.Sum(nil), nil
}

func ser32(v uint32) []byte {
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
