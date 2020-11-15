package testutils

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
)

// CreateTransaction create a transaction that spends the first output of provided transaction.
// Assumes that the output being spent has opTrueScript as it's scriptPublicKey
// Creates the value of the spent output minus 1 sompi
func CreateTransaction(txToSpend *externalapi.DomainTransaction) (*externalapi.DomainTransaction, error) {
	opTrueScript := OpTrueScript()

	scriptPublicKey := opTrueScript
	signatureScript, err := txscript.PayToScriptHashSignatureScript(opTrueScript, nil)
	if err != nil {
		return nil, err
	}
	input := &externalapi.DomainTransactionInput{
		PreviousOutpoint: externalapi.DomainOutpoint{
			TransactionID: *consensusserialization.TransactionID(txToSpend),
			Index:         0,
		},
		SignatureScript: signatureScript,
		Sequence:        constants.MaxTxInSequenceNum,
	}
	output := &externalapi.DomainTransactionOutput{
		ScriptPublicKey: scriptPublicKey,
		Value:           txToSpend.Outputs[0].Value - 1,
	}
	return &externalapi.DomainTransaction{
		Version: constants.TransactionVersion,
		Inputs:  []*externalapi.DomainTransactionInput{input},
		Outputs: []*externalapi.DomainTransactionOutput{output},
		Payload: []byte{},
	}, nil
}
