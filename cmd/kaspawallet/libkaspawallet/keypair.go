package libkaspawallet

import (
	"github.com/kaspanet/go-secp256k1"
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
		return nil, errors.Wrap(err, "Failed to deserialized private key")
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
func Address(params *dagconfig.Params, pubKeys [][]byte, minimumSignatures uint32, ecdsa bool) (util.Address, error) {
	sortPublicKeys(pubKeys)
	if uint32(len(pubKeys)) < minimumSignatures {
		return nil, errors.Errorf("The minimum amount of signatures (%d) is greater than the amount of "+
			"provided public keys (%d)", minimumSignatures, len(pubKeys))
	}
	if len(pubKeys) == 1 {
		if ecdsa {
			return util.NewAddressPublicKeyECDSA(pubKeys[0][:], params.Prefix)
		}
		return util.NewAddressPublicKey(pubKeys[0][:], params.Prefix)
	}

	redeemScript, err := multiSigRedeemScript(pubKeys, minimumSignatures, ecdsa)
	if err != nil {
		return nil, err
	}

	return util.NewAddressScriptHash(redeemScript, params.Prefix)
}
