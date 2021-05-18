package libkaspawallet

import (
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/bip32"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

// Payment contains a recipient payment details
type Payment struct {
	Address util.Address
	Amount  uint64
}

// UTXO is a type that stores a UTXO and meta data
// that is needed in order to sign it and create
// transactions with it.
type UTXO struct {
	Outpoint       *externalapi.DomainOutpoint
	UTXOEntry      externalapi.UTXOEntry
	DerivationPath string
}

// CreateUnsignedTransaction creates an unsigned transaction
func CreateUnsignedTransaction(
	extendedPublicKeys []string,
	minimumSignatures uint32,
	payments []*Payment,
	selectedUTXOs []*UTXO) ([]byte, error) {

	sortPublicKeys(extendedPublicKeys)
	unsignedTransaction, err := createUnsignedTransaction(extendedPublicKeys, minimumSignatures, payments, selectedUTXOs)
	if err != nil {
		return nil, err
	}

	return serialization.SerializePartiallySignedTransaction(unsignedTransaction)
}

func multiSigRedeemScript(extendedPublicKeys []string, minimumSignatures uint32, path string, ecdsa bool) ([]byte, error) {
	scriptBuilder := txscript.NewScriptBuilder()
	scriptBuilder.AddInt64(int64(minimumSignatures))
	for _, key := range extendedPublicKeys {
		extendedKey, err := bip32.DeserializeExtendedKey(key)
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

		var serializedPublicKey []byte
		if ecdsa {
			serializedECDSAPublicKey, err := publicKey.Serialize()
			if err != nil {
				return nil, err
			}
			serializedPublicKey = serializedECDSAPublicKey[:]
		} else {
			schnorrPublicKey, err := publicKey.ToSchnorr()
			if err != nil {
				return nil, err
			}

			serializedSchnorrPublicKey, err := schnorrPublicKey.Serialize()
			if err != nil {
				return nil, err
			}
			serializedPublicKey = serializedSchnorrPublicKey[:]
		}

		scriptBuilder.AddData(serializedPublicKey)
	}
	scriptBuilder.AddInt64(int64(len(extendedPublicKeys)))

	if ecdsa {
		scriptBuilder.AddOp(txscript.OpCheckMultiSigECDSA)
	} else {
		scriptBuilder.AddOp(txscript.OpCheckMultiSig)
	}

	return scriptBuilder.Script()
}

func createUnsignedTransaction(
	extendedPublicKeys []string,
	minimumSignatures uint32,
	payments []*Payment,
	selectedUTXOs []*UTXO) (*serialization.PartiallySignedTransaction, error) {

	inputs := make([]*externalapi.DomainTransactionInput, len(selectedUTXOs))
	partiallySignedInputs := make([]*serialization.PartiallySignedInput, len(selectedUTXOs))
	for i, utxo := range selectedUTXOs {
		emptyPubKeySignaturePairs := make([]*serialization.PubKeySignaturePair, len(extendedPublicKeys))
		for i, extendedPublicKey := range extendedPublicKeys {
			extendedKey, err := bip32.DeserializeExtendedKey(extendedPublicKey)
			if err != nil {
				return nil, err
			}

			derivedKey, err := extendedKey.DeriveFromPath(utxo.DerivationPath)
			if err != nil {
				return nil, err
			}

			emptyPubKeySignaturePairs[i] = &serialization.PubKeySignaturePair{
				ExtendedPublicKey: derivedKey.String(),
			}
		}

		inputs[i] = &externalapi.DomainTransactionInput{PreviousOutpoint: *utxo.Outpoint}
		partiallySignedInputs[i] = &serialization.PartiallySignedInput{
			PrevOutput: &externalapi.DomainTransactionOutput{
				Value:           utxo.UTXOEntry.Amount(),
				ScriptPublicKey: utxo.UTXOEntry.ScriptPublicKey(),
			},
			MinimumSignatures:    minimumSignatures,
			PubKeySignaturePairs: emptyPubKeySignaturePairs,
			DerivationPath:       utxo.DerivationPath,
		}
	}

	outputs := make([]*externalapi.DomainTransactionOutput, len(payments))
	for i, payment := range payments {
		scriptPublicKey, err := txscript.PayToAddrScript(payment.Address)
		if err != nil {
			return nil, err
		}

		outputs[i] = &externalapi.DomainTransactionOutput{
			Value:           payment.Amount,
			ScriptPublicKey: scriptPublicKey,
		}
	}

	domainTransaction := &externalapi.DomainTransaction{
		Version:      constants.MaxTransactionVersion,
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     0,
		SubnetworkID: subnetworks.SubnetworkIDNative,
		Gas:          0,
		Payload:      nil,
	}

	return &serialization.PartiallySignedTransaction{
		Tx:                    domainTransaction,
		PartiallySignedInputs: partiallySignedInputs,
	}, nil
}

// IsTransactionFullySigned returns whether the transaction is fully signed and ready to broadcast.
func IsTransactionFullySigned(partiallySignedTransactionBytes []byte) (bool, error) {
	partiallySignedTransaction, err := serialization.DeserializePartiallySignedTransaction(partiallySignedTransactionBytes)
	if err != nil {
		return false, err
	}

	return isTransactionFullySigned(partiallySignedTransaction), nil
}

func isTransactionFullySigned(partiallySignedTransaction *serialization.PartiallySignedTransaction) bool {
	for _, input := range partiallySignedTransaction.PartiallySignedInputs {
		numSignatures := 0
		for _, pair := range input.PubKeySignaturePairs {
			if pair.Signature != nil {
				numSignatures++
			}
		}
		if uint32(numSignatures) < input.MinimumSignatures {
			return false
		}
	}
	return true
}

// ExtractTransaction extracts a domain transaction from partially signed transaction after all of the
// relevant parties have signed it.
func ExtractTransaction(partiallySignedTransactionBytes []byte, ecdsa bool) (*externalapi.DomainTransaction, error) {
	partiallySignedTransaction, err := serialization.DeserializePartiallySignedTransaction(partiallySignedTransactionBytes)
	if err != nil {
		return nil, err
	}

	return extractTransaction(partiallySignedTransaction, ecdsa)
}

func extractTransaction(partiallySignedTransaction *serialization.PartiallySignedTransaction, ecdsa bool) (*externalapi.DomainTransaction, error) {
	for i, input := range partiallySignedTransaction.PartiallySignedInputs {
		isMultisig := len(input.PubKeySignaturePairs) > 1
		scriptBuilder := txscript.NewScriptBuilder()
		if isMultisig {
			signatureCount := 0
			for _, pair := range input.PubKeySignaturePairs {
				if pair.Signature != nil {
					scriptBuilder.AddData(pair.Signature)
					signatureCount++
				}
			}
			if uint32(signatureCount) < input.MinimumSignatures {
				return nil, errors.Errorf("missing %d signatures", input.MinimumSignatures-uint32(signatureCount))
			}

			redeemScript, err := partiallySignedInputMultisigRedeemScript(input, ecdsa)
			if err != nil {
				return nil, err
			}

			scriptBuilder.AddData(redeemScript)
			sigScript, err := scriptBuilder.Script()
			if err != nil {
				return nil, err
			}

			partiallySignedTransaction.Tx.Inputs[i].SignatureScript = sigScript
		} else {
			if len(input.PubKeySignaturePairs) > 1 {
				return nil, errors.Errorf("Cannot sign on P2PK when len(input.PubKeySignaturePairs) > 1")
			}

			if input.PubKeySignaturePairs[0].Signature == nil {
				return nil, errors.Errorf("missing signature")
			}

			sigScript, err := txscript.NewScriptBuilder().
				AddData(input.PubKeySignaturePairs[0].Signature).
				Script()
			if err != nil {
				return nil, err
			}
			partiallySignedTransaction.Tx.Inputs[i].SignatureScript = sigScript
		}
	}
	return partiallySignedTransaction.Tx, nil
}

func partiallySignedInputMultisigRedeemScript(input *serialization.PartiallySignedInput, ecdsa bool) ([]byte, error) {
	extendedPublicKeys := make([]string, len(input.PubKeySignaturePairs))
	for i, pair := range input.PubKeySignaturePairs {
		extendedPublicKeys[i] = pair.ExtendedPublicKey
	}

	return multiSigRedeemScript(extendedPublicKeys, input.MinimumSignatures, "m", ecdsa)
}
