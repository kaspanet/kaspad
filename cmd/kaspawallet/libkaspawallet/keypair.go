package libkaspawallet

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

func CreateKeyPair(params *dagconfig.Params) ([]byte, []byte, util.Address, error) {
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Failed to generate private key")
	}
	publicKey, err := privateKey.SchnorrPublicKey()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Failed to generate public key")
	}
	publicKeySerialized, err := publicKey.Serialize()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Failed to serialize public key")
	}

	addr, err := util.NewAddressPubKeyHashFromPublicKey(publicKeySerialized[:], params.Prefix)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Failed to generate p2pkh address")
	}

	return privateKey.SerializePrivateKey()[:], publicKeySerialized[:], addr, nil
}

func AddressFromPrivateKey(params *dagconfig.Params, privateKey []byte) (util.Address, error) {
	keyPair, err := secp256k1.DeserializePrivateKeyFromSlice(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "Error deserializing private key")
	}

	publicKey, err := keyPair.SchnorrPublicKey()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating public key")
	}

	publicKeySerialized, err := publicKey.Serialize()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to serialize public key")
	}

	return addressFromPublicKey(params, publicKeySerialized[:])
}

func addressFromPublicKey(params *dagconfig.Params, publicKeySerialized []byte) (util.Address, error) {
	addr, err := util.NewAddressPubKeyHashFromPublicKey(publicKeySerialized[:], params.Prefix)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate p2pkh address")
	}

	return addr, nil
}

func MultiSigAddress(params *dagconfig.Params, pubKeys [][]byte, minimumSignatures uint32) (util.Address, error) {
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
