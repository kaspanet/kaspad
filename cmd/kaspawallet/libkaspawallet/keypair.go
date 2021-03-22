package libkaspawallet

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

func CreateKeyPair(params *dagconfig.Params) ([]byte, util.Address, error) {
	privateKey, err := secp256k1.GeneratePrivateKey()
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

	addr, err := util.NewAddressPubKeyHashFromPublicKey(publicKeySerialized[:], params.Prefix)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to generate p2pkh address")
	}

	return privateKey.SerializePrivateKey()[:], addr, nil
}

func parsePrivateKey(privateKey []byte) (*secp256k1.SchnorrKeyPair, *secp256k1.SchnorrPublicKey, error) {
	keyPair, err := secp256k1.DeserializePrivateKeyFromSlice(privateKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error deserializing private key")
	}
	publicKey, err := keyPair.SchnorrPublicKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error generating public key")
	}
	return keyPair, publicKey, nil
}
