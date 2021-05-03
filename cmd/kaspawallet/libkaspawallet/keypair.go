package libkaspawallet

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/bip32"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"math"
	"sort"
	"strings"
)

// CreateKeyPair generates a private-public key pair
func CreateKeyPair(ecdsa bool) ([]byte, []byte, error) {
	if ecdsa {
		return createKeyPairECDSA()
	}

	return createKeyPair()
}

func createKeyPair() ([]byte, []byte, error) {
	keyPair, err := secp256k1.GenerateSchnorrKeyPair()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to generate private key")
	}
	publicKey, err := keyPair.SchnorrPublicKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to generate public key")
	}
	publicKeySerialized, err := publicKey.Serialize()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to serialize public key")
	}

	return keyPair.SerializePrivateKey()[:], publicKeySerialized[:], nil
}

func createKeyPairECDSA() ([]byte, []byte, error) {
	keyPair, err := secp256k1.GenerateECDSAPrivateKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to generate private key")
	}
	publicKey, err := keyPair.ECDSAPublicKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to generate public key")
	}
	publicKeySerialized, err := publicKey.Serialize()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to serialize public key")
	}

	return keyPair.Serialize()[:], publicKeySerialized[:], nil
}

// PublicKeyFromPrivateKey returns the public key associated with a private key
func PublicKeyFromPrivateKey(privateKeyBytes []byte) ([]byte, error) {
	keyPair, err := secp256k1.DeserializeSchnorrPrivateKeyFromSlice(privateKeyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to deserialize private key")
	}

	publicKey, err := keyPair.SchnorrPublicKey()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate public key")
	}

	publicKeySerialized, err := publicKey.Serialize()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to serialize public key")
	}

	return publicKeySerialized[:], nil
}

// Address returns the address associated with the given public keys and minimum signatures parameters.
func Address(params *dagconfig.Params, extendedPublicKeys []string, minimumSignatures uint32, path string, ecdsa bool) (util.Address, error) {
	sortPublicKeys(extendedPublicKeys)
	if uint32(len(extendedPublicKeys)) < minimumSignatures {
		return nil, errors.Errorf("The minimum amount of signatures (%d) is greater than the amount of "+
			"provided public keys (%d)", minimumSignatures, len(extendedPublicKeys))
	}

	if len(extendedPublicKeys) == 1 {
		return p2pkAddress(params, extendedPublicKeys[0], path, ecdsa)
	}

	redeemScript, err := multiSigRedeemScript(extendedPublicKeys, minimumSignatures, path, ecdsa)
	if err != nil {
		return nil, err
	}

	return util.NewAddressScriptHash(redeemScript, params.Prefix)
}

func p2pkAddress(params *dagconfig.Params, extendedPublicKey string, path string, ecdsa bool) (util.Address, error) {
	extendedKey, err := bip32.DeserializeExtendedKey(extendedPublicKey)
	if err != nil {
		return nil, err
	}

	derivedKey, err := extendedKey.DeriveFromPath(path)
	if err != nil {
		return nil, err
	}

	publicKey, err := derivedKey.PublicKey()
	if err != nil {
		return nil, err
	}

	if ecdsa {
		serializedECDSAPublicKey, err := publicKey.Serialize()
		if err != nil {
			return nil, err
		}
		return util.NewAddressPublicKeyECDSA(serializedECDSAPublicKey[:], params.Prefix)
	}

	schnorrPublicKey, err := publicKey.ToSchnorr()
	if err != nil {
		return nil, err
	}

	serializedSchnorrPublicKey, err := schnorrPublicKey.Serialize()
	if err != nil {
		return nil, err
	}

	return util.NewAddressPublicKey(serializedSchnorrPublicKey[:], params.Prefix)
}

func sortPublicKeys(extendedPublicKeys []string) {
	sort.Slice(extendedPublicKeys, func(i, j int) bool {
		return strings.Compare(extendedPublicKeys[i], extendedPublicKeys[j]) < 0
	})
}

func cosignerIndex(extendedPublicKey string, sortedExtendedPublicKeys []string) (uint32, error) {
	cosignerIndex := sort.SearchStrings(sortedExtendedPublicKeys, extendedPublicKey)
	if cosignerIndex == len(sortedExtendedPublicKeys) {
		return 0, errors.Errorf("couldn't find extended public key %s", extendedPublicKey)
	}

	return uint32(cosignerIndex), nil
}

func MinimumCosignerIndex(signerExtendedPublicKeys, allExtendedPublicKeys []string) (uint32, error) {
	allExtendedPublicKeysCopy := make([]string, len(allExtendedPublicKeys))
	copy(allExtendedPublicKeysCopy, allExtendedPublicKeys)
	sortPublicKeys(allExtendedPublicKeysCopy)

	min := uint32(math.MaxUint32)
	for _, extendedPublicKey := range signerExtendedPublicKeys {
		cosignerIndex, err := cosignerIndex(extendedPublicKey, allExtendedPublicKeysCopy)
		if err != nil {
			return 0, err
		}

		if cosignerIndex < min {
			min = cosignerIndex
		}
	}

	return min, nil
}
