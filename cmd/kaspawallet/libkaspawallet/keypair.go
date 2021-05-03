package libkaspawallet

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/bip32"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
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
func Address(params *dagconfig.Params, extendedPublicKeys []string, minimumSignatures uint32, ecdsa bool) (util.Address, error) {
	sortPublicKeys(extendedPublicKeys)
	if uint32(len(extendedPublicKeys)) < minimumSignatures {
		return nil, errors.Errorf("The minimum amount of signatures (%d) is greater than the amount of "+
			"provided public keys (%d)", minimumSignatures, len(extendedPublicKeys))
	}
	if len(extendedPublicKeys) == 1 {
		return p2pkAddress(params, extendedPublicKeys[0], ecdsa)
	}

	redeemScript, err := multiSigRedeemScript(extendedPublicKeys, minimumSignatures, ecdsa)
	if err != nil {
		return nil, err
	}

	return util.NewAddressScriptHash(redeemScript, params.Prefix)
}

func p2pkAddress(params *dagconfig.Params, extendedPublicKey string, ecdsa bool) (util.Address, error) {
	extendedKey, err := bip32.DeserializeExtendedKey(extendedPublicKey)
	if err != nil {
		return nil, err
	}

	// TODO: Implement no-reuse address policy
	firstChild, err := extendedKey.Child(0)
	if err != nil {
		return nil, err
	}

	publicKey, err := firstChild.PublicKey()
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
