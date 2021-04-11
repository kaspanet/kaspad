package libkaspawallet

import (
	"bytes"
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/pkg/errors"
)

type signer interface {
	rawTxInSignature(tx *externalapi.DomainTransaction, idx int, hashType consensushashing.SigHashType,
		sighashReusedValues *consensushashing.SighashReusedValues) ([]byte, error)
	serializedPublicKey() ([]byte, error)
}

type schnorrSigner secp256k1.SchnorrKeyPair

func (s *schnorrSigner) rawTxInSignature(tx *externalapi.DomainTransaction, idx int, hashType consensushashing.SigHashType,
	sighashReusedValues *consensushashing.SighashReusedValues) ([]byte, error) {
	return txscript.RawTxInSignature(tx, idx, hashType, (*secp256k1.SchnorrKeyPair)(s), sighashReusedValues)
}

func (s *schnorrSigner) serializedPublicKey() ([]byte, error) {
	publicKey, err := (*secp256k1.SchnorrKeyPair)(s).SchnorrPublicKey()
	if err != nil {
		return nil, err
	}

	serializedPublicKey, err := publicKey.Serialize()
	if err != nil {
		return nil, err
	}

	return serializedPublicKey[:], nil
}

type ecdsaSigner secp256k1.ECDSAPrivateKey

func (e *ecdsaSigner) rawTxInSignature(tx *externalapi.DomainTransaction, idx int, hashType consensushashing.SigHashType,
	sighashReusedValues *consensushashing.SighashReusedValues) ([]byte, error) {
	return txscript.RawTxInSignatureECDSA(tx, idx, hashType, (*secp256k1.ECDSAPrivateKey)(e), sighashReusedValues)
}

func (e *ecdsaSigner) serializedPublicKey() ([]byte, error) {
	publicKey, err := (*secp256k1.ECDSAPrivateKey)(e).ECDSAPublicKey()
	if err != nil {
		return nil, err
	}

	serializedPublicKey, err := publicKey.Serialize()
	if err != nil {
		return nil, err
	}

	return serializedPublicKey[:], nil
}

func deserializeECDSAPrivateKey(privateKey []byte, ecdsa bool) (signer, error) {
	if ecdsa {
		keyPair, err := secp256k1.DeserializeECDSAPrivateKeyFromSlice(privateKey)
		if err != nil {
			return nil, errors.Wrap(err, "Error deserializing private key")
		}

		return (*ecdsaSigner)(keyPair), nil
	}

	keyPair, err := secp256k1.DeserializeSchnorrPrivateKeyFromSlice(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "Error deserializing private key")
	}

	return (*schnorrSigner)(keyPair), nil
}

// Sign signs the transaction with the given private keys
func Sign(privateKeys [][]byte, serializedPSTx []byte, ecdsa bool) ([]byte, error) {
	keyPairs := make([]signer, len(privateKeys))
	for i, privateKey := range privateKeys {
		var err error
		keyPairs[i], err = deserializeECDSAPrivateKey(privateKey, ecdsa)
		if err != nil {
			return nil, errors.Wrap(err, "Error deserializing private key")
		}
	}

	partiallySignedTransaction, err := serialization.DeserializePartiallySignedTransaction(serializedPSTx)
	if err != nil {
		return nil, err
	}

	for _, keyPair := range keyPairs {
		err = sign(keyPair, partiallySignedTransaction)
		if err != nil {
			return nil, err
		}
	}

	return serialization.SerializePartiallySignedTransaction(partiallySignedTransaction)
}

func sign(keyPair signer, psTx *serialization.PartiallySignedTransaction) error {
	if isTransactionFullySigned(psTx) {
		return nil
	}

	serializedPublicKey, err := keyPair.serializedPublicKey()
	if err != nil {
		return err
	}

	sighashReusedValues := &consensushashing.SighashReusedValues{}
	for i, partiallySignedInput := range psTx.PartiallySignedInputs {
		prevOut := partiallySignedInput.PrevOutput
		psTx.Tx.Inputs[i].UTXOEntry = utxo.NewUTXOEntry(
			prevOut.Value,
			prevOut.ScriptPublicKey,
			false, // This is a fake value, because it's irrelevant for the signature
			0,     // This is a fake value, because it's irrelevant for the signature
		)
	}

	signed := false
	for i, partiallySignedInput := range psTx.PartiallySignedInputs {
		for _, pair := range partiallySignedInput.PubKeySignaturePairs {
			if bytes.Equal(pair.PubKey, serializedPublicKey[:]) {
				pair.Signature, err = keyPair.rawTxInSignature(psTx.Tx, i, consensushashing.SigHashAll, sighashReusedValues)
				if err != nil {
					return err
				}

				signed = true
			}
		}
	}

	if !signed {
		return errors.Errorf("Public key doesn't match any of the transaction public keys")
	}

	return nil
}
