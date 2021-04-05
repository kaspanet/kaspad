package libkaspawallet

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

// CreateKeyPair generates a private-public key pair
func CreateKeyPair() ([]byte, []byte, error) {
	privateKey, err := secp256k1.GenerateSchnorrKeyPair()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to generate private key")
	}
	publicKey, err := privateKey.SchnorrPublicKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to generate public key")
	}
	publicKeySerialized, err := publicKey.Serialize()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to serialize public key")
	}

	return privateKey.SerializePrivateKey()[:], publicKeySerialized[:], nil
}

func addressFromPublicKey(params *dagconfig.Params, publicKeySerialized []byte) (util.Address, error) {
	addr, err := util.NewAddressPublicKey(publicKeySerialized[:], params.Prefix)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate p2pkh address")
	}

	return addr, nil
}

// Address returns the address associated with the given public keys and minimum signatures parameters.
func Address(params *dagconfig.Params, pubKeys [][]byte, minimumSignatures uint32) (util.Address, error) {
	sortPublicKeys(pubKeys)
	if uint32(len(pubKeys)) < minimumSignatures {
		return nil, errors.Errorf("The minimum amount of signatures (%d) is greater than the amount of "+
			"provided public keys (%d)", minimumSignatures, len(pubKeys))
	}
	if len(pubKeys) == 1 {
		return addressFromPublicKey(params, pubKeys[0])
	}

	redeemScript, err := multiSigRedeemScript(pubKeys, minimumSignatures)
	if err != nil {
		return nil, err
	}

	return util.NewAddressScriptHash(redeemScript, params.Prefix)
}
