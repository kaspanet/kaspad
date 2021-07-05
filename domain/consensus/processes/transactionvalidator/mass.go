package transactionvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/estimatedsize"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
)

func (v *transactionValidator) transactionMassStandalonePart(tx *externalapi.DomainTransaction) uint64 {
	size := estimatedsize.TransactionEstimatedSerializedSize(tx)

	totalScriptPubKeySize := uint64(0)
	for _, output := range tx.Outputs {
		totalScriptPubKeySize += 2 //output.ScriptPublicKey.Version (uint16)
		totalScriptPubKeySize += uint64(len(output.ScriptPublicKey.Script))
	}

	return size*v.massPerTxByte + totalScriptPubKeySize*v.massPerScriptPubKeyByte
}

func (v *transactionValidator) transactionMass(tx *externalapi.DomainTransaction) uint64 {
	if transactionhelper.IsCoinBase(tx) {
		return 0
	}

	// calculate mass for size
	size := estimatedsize.TransactionEstimatedSerializedSize(tx)
	massForSize := size * v.massPerTxByte

	// calculate mass for scriptPubKey
	totalScriptPubKeySize := uint64(0)
	for _, output := range tx.Outputs {
		totalScriptPubKeySize += 2 //output.ScriptPublicKey.Version (uint16)
		totalScriptPubKeySize += uint64(len(output.ScriptPublicKey.Script))
	}
	massForScriptPubKey := totalScriptPubKeySize * v.massPerScriptPubKeyByte

	// calculate mass for SigOps
	totalSigOpCount := uint64(0)
	for _, input := range tx.Inputs {
		totalSigOpCount += uint64(input.SigOpCount)
	}
	massForSigOps := totalSigOpCount * v.massPerSigOp

	// Sum all components of mass
	return massForSize + massForScriptPubKey + massForSigOps
}
