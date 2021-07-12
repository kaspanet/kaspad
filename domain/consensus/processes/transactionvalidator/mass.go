package transactionvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/estimatedsize"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
)

func (v *transactionValidator) PopulateMass(transaction *externalapi.DomainTransaction) {
	if transaction.Mass == 0 {
		return
	}
	transaction.Mass = v.transactionMass(transaction)
}

func (v *transactionValidator) transactionMass(transaction *externalapi.DomainTransaction) uint64 {
	if transactionhelper.IsCoinBase(transaction) {
		return 0
	}

	// calculate mass for size
	size := estimatedsize.TransactionEstimatedSerializedSize(transaction)
	massForSize := size * v.massPerTxByte

	// calculate mass for scriptPubKey
	totalScriptPubKeySize := uint64(0)
	for _, output := range transaction.Outputs {
		totalScriptPubKeySize += 2 //output.ScriptPublicKey.Version (uint16)
		totalScriptPubKeySize += uint64(len(output.ScriptPublicKey.Script))
	}
	massForScriptPubKey := totalScriptPubKeySize * v.massPerScriptPubKeyByte

	// calculate mass for SigOps
	totalSigOpCount := uint64(0)
	for _, input := range transaction.Inputs {
		totalSigOpCount += uint64(input.SigOpCount)
	}
	massForSigOps := totalSigOpCount * v.massPerSigOp

	// Sum all components of mass
	return massForSize + massForScriptPubKey + massForSigOps
}
