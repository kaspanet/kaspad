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
	"sort"
	"strings"
)

// Payment contains a recipient payment details
type Payment struct {
	Address util.Address
	Amount  uint64
}

func sortPublicKeys(extendedPublicKeys []string) {
	sort.Slice(extendedPublicKeys, func(i, j int) bool {
		return strings.Compare(extendedPublicKeys[i], extendedPublicKeys[j]) < 0
	})
}

// CreateUnsignedTransaction creates an unsigned transaction
func CreateUnsignedTransaction(
	extendedPublicKeys []string,
	minimumSignatures uint32,
	ecdsa bool,
	payments []*Payment,
	selectedUTXOs []*externalapi.OutpointAndUTXOEntryPair) ([]byte, error) {

	sortPublicKeys(extendedPublicKeys)
	unsignedTransaction, err := createUnsignedTransaction(extendedPublicKeys, minimumSignatures, ecdsa, payments, selectedUTXOs)
	if err != nil {
		return nil, err
	}

	return serialization.SerializePartiallySignedTransaction(unsignedTransaction)
}

func multiSigRedeemScript(extendedPublicKeys []string, minimumSignatures uint32, ecdsa bool) ([]byte, error) {
	scriptBuilder := txscript.NewScriptBuilder()
	scriptBuilder.AddInt64(int64(minimumSignatures))
	for _, key := range extendedPublicKeys {
		extendedKey, err := bip32.DeserializeExtendedKey(key)
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
	ecdsa bool,
	payments []*Payment,
	selectedUTXOs []*externalapi.OutpointAndUTXOEntryPair) (*serialization.PartiallySignedTransaction, error) {

	var redeemScript []byte
	if len(extendedPublicKeys) > 1 {
		var err error
		redeemScript, err = multiSigRedeemScript(extendedPublicKeys, minimumSignatures, ecdsa)
		if err != nil {
			return nil, err
		}
	}

	inputs := make([]*externalapi.DomainTransactionInput, len(selectedUTXOs))
	partiallySignedInputs := make([]*serialization.PartiallySignedInput, len(selectedUTXOs))
	for i, utxo := range selectedUTXOs {
		emptyPubKeySignaturePairs := make([]*serialization.PubKeySignaturePair, len(extendedPublicKeys))
		for i, extendedPublicKey := range extendedPublicKeys {
			extendedKey, err := bip32.DeserializeExtendedKey(extendedPublicKey)
			if err != nil {
				return nil, err
			}

			// TODO: Implement no-reuse address policy
			firstChild, err := extendedKey.Child(0)
			if err != nil {
				return nil, err
			}

			emptyPubKeySignaturePairs[i] = &serialization.PubKeySignaturePair{
				ExtendedPublicKey: firstChild.String(),
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
			DerivationPath:       "m/0", // TODO: Implement no-reuse address policy
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
			psTx.Tx.Inputs[i].SignatureScript = sigScript
		}
	}
	return psTx.Tx, nil
}
