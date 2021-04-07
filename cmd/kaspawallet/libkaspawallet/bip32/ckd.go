package bip32

import (
	"crypto/sha256"
	"encoding/binary"
	"github.com/kaspanet/go-secp256k1"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ripemd160"
)

const hardenedIndexStart = 0x80000000

func NewMaster(seed []byte, version [4]byte) (*ExtendedPrivateKey, error) {
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

	return &ExtendedPrivateKey{
		PrivateKey: privateKey,
		extendedKey: &extendedKey{
			Version:     version,
			Depth:       0,
			Fingerprint: [4]byte{},
			ChildNumber: 0,
			ChainCode:   iR,
		},
	}, nil
}

func isHardened(i uint32) bool {
	return i >= hardenedIndexStart
}

func (extPrv *ExtendedPrivateKey) Child(i uint32) (*ExtendedPrivateKey, error) {
	I, err := ckdPrivCalcI(extPrv, i)
	if err != nil {
		return nil, err
	}

	var iL, iR [32]byte
	copy(iL[:], I[:32])
	copy(iR[:], I[32:])

	fingerPrint, err := fingerprintFromPrivateKey(extPrv.PrivateKey)
	if err != nil {
		return nil, err
	}

	childPrivateKey, err := privateKeyAdd(extPrv.PrivateKey, iL)
	if err != nil {
		return nil, err
	}

	return &ExtendedPrivateKey{
		PrivateKey: childPrivateKey,
		extendedKey: &extendedKey{
			Version:     extPrv.Version,
			Depth:       extPrv.Depth + 1,
			Fingerprint: fingerPrint,
			ChildNumber: i,
			ChainCode:   iR,
		},
	}, nil
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

func ckdPrivCalcI(extPrv *ExtendedPrivateKey, i uint32) ([]byte, error) {
	if isHardened(i) {
		mac := newHMACWriter(extPrv.ChainCode[:])
		mac.InfallibleWrite([]byte{0x00})
		mac.InfallibleWrite(extPrv.PrivateKey.Serialize()[:])
		mac.InfallibleWrite(ser32(i))
		return mac.Sum(nil), nil
	}

	extPub, err := extPrv.Public()
	if err != nil {
		return nil, err
	}

	return ckdPubCalcI(extPub, i)
}

func ckdPubCalcI(extPub *ExtendedPublicKey, i uint32) ([]byte, error) {
	mac := newHMACWriter(extPub.ChainCode[:])
	serializedPoint, err := extPub.PublicKey.Serialize()
	if err != nil {
		return nil, errors.Wrap(err, "error serializing point")
	}

	mac.InfallibleWrite(serializedPoint[:])
	mac.InfallibleWrite(ser32(i))
	return mac.Sum(nil), nil
}

func ser32(v uint32) []byte {
	serialized := make([]byte, 4)
	binary.BigEndian.PutUint32(serialized, v)
	return serialized
}

func (extPub *ExtendedPublicKey) Child(i uint32) (*ExtendedPublicKey, error) {
	if isHardened(i) {
		return nil, errors.Errorf("CKDpub cannot operate with hardened indexes (0x%x)", i)
	}

	I, err := ckdPubCalcI(extPub, i)
	if err != nil {
		return nil, err
	}

	var iL, iR [32]byte
	copy(iL[:], I[:32])
	copy(iR[:], I[32:])

	childPoint, err := pointAdd(extPub.PublicKey, iL)
	if err != nil {
		return nil, err
	}

	fingerPrint, err := fingerPrintFromPoint(childPoint)
	if err != nil {
		return nil, err
	}

	return &ExtendedPublicKey{
		PublicKey: childPoint,
		extendedKey: &extendedKey{
			Version:     extPub.Version,
			Depth:       extPub.Depth + 1,
			Fingerprint: fingerPrint,
			ChildNumber: i,
			ChainCode:   iR,
		}}, nil
}

func pointAdd(point *secp256k1.ECDSAPublicKey, tweak [32]byte) (*secp256k1.ECDSAPublicKey, error) {
	pointCopy := *point
	err := pointCopy.Add(tweak)
	if err != nil {
		return nil, err
	}

	return &pointCopy, nil
}
