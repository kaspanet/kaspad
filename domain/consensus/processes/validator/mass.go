package validator

import "github.com/kaspanet/kaspad/domain/consensus/model"

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

func (bv *Validator) transactionMassStandalonePart(tx *model.DomainTransaction) uint64 {
	size := bv.transactionEstimatedSerializedSize(tx)

	totalScriptPubKeySize := uint64(0)
	for _, output := range tx.Outputs {
		totalScriptPubKeySize += uint64(len(output.ScriptPublicKey))
	}

	return size*MassPerTxByte + totalScriptPubKeySize*MassPerScriptPubKeyByte
}
