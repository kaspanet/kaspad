package transactionvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/estimatedsize"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
)

func (v *transactionValidator) transactionMassStandalonePart(tx *externalapi.DomainTransaction) uint64 {
	size := estimatedsize.TransactionEstimatedSerializedSize(tx)

	totalScriptPubKeySize := uint64(0)
	for _, output := range tx.Outputs {
		totalScriptPubKeySize += uint64(len(output.ScriptPublicKey))
	}

	return size*v.massPerTxByte + totalScriptPubKeySize*v.massPerScriptPubKeyByte
}

func (v *transactionValidator) transactionMass(tx *externalapi.DomainTransaction) (uint64, error) {
	if transactionhelper.IsCoinBase(tx) {
		return 0, nil
	}

	standaloneMass := v.transactionMassStandalonePart(tx)
	sigOpsCount := uint64(0)
	var missingOutpoints []*externalapi.DomainOutpoint
	for _, input := range tx.Inputs {
		utxoEntry := input.UTXOEntry
		if utxoEntry == nil {
			missingOutpoints = append(missingOutpoints, &input.PreviousOutpoint)
			continue
		}
		// Count the precise number of signature operations in the
		// referenced public key script.
		sigScript := input.SignatureScript
		isP2SH := txscript.IsPayToScriptHash(utxoEntry.ScriptPublicKey())
		sigOpsCount += uint64(txscript.GetPreciseSigOpCount(sigScript, utxoEntry.ScriptPublicKey(), isP2SH))
	}
	if len(missingOutpoints) > 0 {
		return 0, ruleerrors.NewErrMissingTxOut(missingOutpoints)
	}

	return standaloneMass + sigOpsCount*v.massPerSigOp, nil
}
