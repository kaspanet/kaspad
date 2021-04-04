package libkaspawallet

import (
	"bytes"
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"sort"
)

// Payment contains a recipient payment details
type Payment struct {
	Address util.Address
	Amount  uint64
}

func sortPublicKeys(publicKeys [][]byte) {
	sort.Slice(publicKeys, func(i, j int) bool {
		return bytes.Compare(publicKeys[i], publicKeys[j]) < 0
	})
}

// CreateUnsignedTransaction creates an unsigned transaction
func CreateUnsignedTransaction(
	pubKeys [][]byte,
	minimumSignatures uint32,
	payments []*Payment,
	selectedUTXOs []*externalapi.OutpointAndUTXOEntryPair) ([]byte, error) {

	sortPublicKeys(pubKeys)
	unsignedTransaction, err := createUnsignedTransaction(pubKeys, minimumSignatures, payments, selectedUTXOs)
	if err != nil {
		return nil, err
	}

	return serialization.SerializePartiallySignedTransaction(unsignedTransaction)
}

// Sign signs the transaction with the given private keys
func Sign(privateKeys [][]byte, serializedPSTx []byte) ([]byte, error) {
	keyPairs := make([]*secp256k1.SchnorrKeyPair, len(privateKeys))
	for i, privateKey := range privateKeys {
		var err error
		keyPairs[i], err = secp256k1.DeserializeSchnorrPrivateKeyFromSlice(privateKey)
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

func multiSigRedeemScript(pubKeys [][]byte, minimumSignatures uint32) ([]byte, error) {
	scriptBuilder := txscript.NewScriptBuilder()
	scriptBuilder.AddInt64(int64(minimumSignatures))
	for _, key := range pubKeys {
		scriptBuilder.AddData(key)
	}
	scriptBuilder.AddInt64(int64(len(pubKeys)))
	scriptBuilder.AddOp(txscript.OpCheckMultiSig)
	return scriptBuilder.Script()
}

func createUnsignedTransaction(
	pubKeys [][]byte,
	minimumSignatures uint32,
	payments []*Payment,
	selectedUTXOs []*externalapi.OutpointAndUTXOEntryPair) (*serialization.PartiallySignedTransaction, error) {

	var redeemScript []byte
	if len(pubKeys) > 1 {
		var err error
		redeemScript, err = multiSigRedeemScript(pubKeys, minimumSignatures)
		if err != nil {
			return nil, err
		}
	}

	inputs := make([]*externalapi.DomainTransactionInput, len(selectedUTXOs))
	partiallySignedInputs := make([]*serialization.PartiallySignedInput, len(selectedUTXOs))
	for i, utxo := range selectedUTXOs {
		emptyPubKeySignaturePairs := make([]*serialization.PubKeySignaturePair, len(pubKeys))
		for i, pubKey := range pubKeys {
			emptyPubKeySignaturePairs[i] = &serialization.PubKeySignaturePair{
				PubKey: pubKey,
			}
		}

		inputs[i] = &externalapi.DomainTransactionInput{PreviousOutpoint: *utxo.Outpoint}
		partiallySignedInputs[i] = &serialization.PartiallySignedInput{
			RedeeemScript: redeemScript,
			PrevOutput: &externalapi.DomainTransactionOutput{
				Value:           utxo.UTXOEntry.Amount(),
				ScriptPublicKey: utxo.UTXOEntry.ScriptPublicKey(),
			},
			MinimumSignatures:    minimumSignatures,
			PubKeySignaturePairs: emptyPubKeySignaturePairs,
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

func sign(keyPair *secp256k1.SchnorrKeyPair, psTx *serialization.PartiallySignedTransaction) error {
	if isTransactionFullySigned(psTx) {
		return nil
	}

	publicKey, err := keyPair.SchnorrPublicKey()
	if err != nil {
		return err
	}

	serializedPublicKey, err := publicKey.Serialize()
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
				pair.Signature, err = txscript.RawTxInSignature(psTx.Tx, i, consensushashing.SigHashAll, keyPair, sighashReusedValues)
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

// IsTransactionFullySigned returns whether the transaction is fully signed and ready to broadcast.
func IsTransactionFullySigned(psTxBytes []byte) (bool, error) {
	partiallySignedTransaction, err := serialization.DeserializePartiallySignedTransaction(psTxBytes)
	if err != nil {
		return false, err
	}

	return isTransactionFullySigned(partiallySignedTransaction), nil
}

func isTransactionFullySigned(psTx *serialization.PartiallySignedTransaction) bool {
	for _, input := range psTx.PartiallySignedInputs {
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
func ExtractTransaction(psTxBytes []byte) (*externalapi.DomainTransaction, error) {
	partiallySignedTransaction, err := serialization.DeserializePartiallySignedTransaction(psTxBytes)
	if err != nil {
		return nil, err
	}

	return extractTransaction(partiallySignedTransaction)
}

func extractTransaction(psTx *serialization.PartiallySignedTransaction) (*externalapi.DomainTransaction, error) {
	for i, input := range psTx.PartiallySignedInputs {
		isMultisig := input.RedeeemScript != nil
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

			scriptBuilder.AddData(input.RedeeemScript)
			sigScript, err := scriptBuilder.Script()
			if err != nil {
				return nil, err
			}

			psTx.Tx.Inputs[i].SignatureScript = sigScript
		} else {
			if len(input.PubKeySignaturePairs) > 1 {
				return nil, errors.Errorf("Cannot sign on P2PKH when len(input.PubKeySignaturePairs) > 1")
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
			psTx.Tx.Inputs[i].SignatureScript = sigScript
		}
	}
	return psTx.Tx, nil
}
