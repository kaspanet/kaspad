package transactionvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/estimatedsize"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/pkg/errors"
)

const (
	// MassPerTxByte is the number of grams that any byte
	// adds to a transaction.
	MassPerTxByte = 1

	// MassPerScriptPubKeyByte is the number of grams that any
	// scriptPubKey byte adds to a transaction.
	MassPerScriptPubKeyByte = 10

	// MassPerSigOp is the number of grams that any
	// signature operation adds to a transaction.
	MassPerSigOp = 10000
)

func (v *transactionValidator) transactionMassStandalonePart(tx *externalapi.DomainTransaction) uint64 {
	size := estimatedsize.TransactionEstimatedSerializedSize(tx)

	totalScriptPubKeySize := uint64(0)
	for _, output := range tx.Outputs {
		totalScriptPubKeySize += uint64(len(output.ScriptPublicKey))
	}

	return size*MassPerTxByte + totalScriptPubKeySize*MassPerScriptPubKeyByte
}

func (v *transactionValidator) transactionMass(tx *externalapi.DomainTransaction) (uint64, error) {
	standaloneMass := v.transactionMassStandalonePart(tx)
	sigOpsCount := uint64(0)
	for _, input := range tx.Inputs {
		utxoEntry := input.UTXOEntry
		if utxoEntry == nil {
			return 0, errors.Wrapf(ruleerrors.ErrMissingTxOut, "output %s "+
				"either does not exist or "+
				"has already been spent", input.PreviousOutpoint)
		}
		// Count the precise number of signature operations in the
		// referenced public key script.
		sigScript := input.SignatureScript
		isP2SH := txscript.IsPayToScriptHash(utxoEntry.ScriptPublicKey)
		sigOpsCount += uint64(txscript.GetPreciseSigOpCount(sigScript, utxoEntry.ScriptPublicKey, isP2SH))
	}

	return standaloneMass + sigOpsCount*MassPerSigOp, nil
}
