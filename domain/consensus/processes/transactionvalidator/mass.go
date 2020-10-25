package transactionvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/estimatedsize"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
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

func (v *transactionValidator) transactionMass(tx *externalapi.DomainTransaction) uint64 {
	standaloneMass := v.transactionMassStandalonePart(tx)
	sigOpsCount := uint64(0)
	for _, input := range tx.Inputs {
		// Count the precise number of signature operations in the
		// referenced public key script.
		sigScript := input.SignatureScript
		isP2SH := txscript.IsPayToScriptHash(input.UTXOEntry.ScriptPublicKey)
		sigOpsCount += uint64(txscript.GetPreciseSigOpCount(sigScript, input.UTXOEntry.ScriptPublicKey, isP2SH))
	}

	return standaloneMass + sigOpsCount*MassPerSigOp
}
