package libkaspawallet

import (
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/bip32"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
)

func rawTxInSignature(extendedKey *bip32.ExtendedKey, tx *externalapi.DomainTransaction, idx int, hashType consensushashing.SigHashType,
	sighashReusedValues *consensushashing.SighashReusedValues, ecdsa bool) ([]byte, error) {

	privateKey := extendedKey.PrivateKey()
	if ecdsa {
		return txscript.RawTxInSignatureECDSA(tx, idx, hashType, privateKey, sighashReusedValues)
	}

	schnorrKeyPair, err := privateKey.ToSchnorr()
	if err != nil {
		return nil, err
	}

	return txscript.RawTxInSignature(tx, idx, hashType, schnorrKeyPair, sighashReusedValues)
}

// Sign signs the transaction with the given private keys
func Sign(params *dagconfig.Params, mnemonics []string, serializedPSTx []byte, ecdsa bool) ([]byte, error) {
	partiallySignedTransaction, err := serialization.DeserializePartiallySignedTransaction(serializedPSTx)
	if err != nil {
		return nil, err
	}

	for _, mnemonic := range mnemonics {
		err = sign(params, mnemonic, partiallySignedTransaction, ecdsa)
		if err != nil {
			return nil, err
		}
	}

	return serialization.SerializePartiallySignedTransaction(partiallySignedTransaction)
}

func sign(params *dagconfig.Params, mnemonic string, partiallySignedTransaction *serialization.PartiallySignedTransaction, ecdsa bool) error {
	if isTransactionFullySigned(partiallySignedTransaction) {
		return nil
	}

	sighashReusedValues := &consensushashing.SighashReusedValues{}
	for i, partiallySignedInput := range partiallySignedTransaction.PartiallySignedInputs {
		prevOut := partiallySignedInput.PrevOutput
		partiallySignedTransaction.Tx.Inputs[i].UTXOEntry = utxo.NewUTXOEntry(
			prevOut.Value,
			prevOut.ScriptPublicKey,
			false, // This is a fake value, because it's irrelevant for the signature
			0,     // This is a fake value, because it's irrelevant for the signature
		)
		partiallySignedTransaction.Tx.Inputs[i].SigOpCount = byte(len(partiallySignedInput.PubKeySignaturePairs))
	}

	signed := false
	for i, partiallySignedInput := range partiallySignedTransaction.PartiallySignedInputs {
		isMultisig := len(partiallySignedInput.PubKeySignaturePairs) > 1
		path := defaultPath(isMultisig)
		extendedKey, err := extendedKeyFromMnemonicAndPath(mnemonic, path, params)
		if err != nil {
			return err
		}

		derivedKey, err := extendedKey.DeriveFromPath(partiallySignedInput.DerivationPath)
		if err != nil {
			return err
		}

		derivedPublicKey, err := derivedKey.Public()
		if err != nil {
			return err
		}

		for _, pair := range partiallySignedInput.PubKeySignaturePairs {
			if pair.ExtendedPublicKey == derivedPublicKey.String() {
				pair.Signature, err = rawTxInSignature(derivedKey, partiallySignedTransaction.Tx, i, consensushashing.SigHashAll, sighashReusedValues, ecdsa)
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
