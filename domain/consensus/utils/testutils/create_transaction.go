package testutils

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
)

// opTrueScript is script returning TRUE
var opTrueScript = []byte{txscript.OpTrue}

// SimpleCoinbaseData can be used as a default CoinbaseData for TestConsensus.AddBlock in tests that don't care about
// their coinbase data
var SimpleCoinbaseData = &externalapi.DomainCoinbaseData{ScriptPublicKey: opTrueScript, ExtraData: []byte{}}

// CreateTransaction create a transaction that spends the first output of provided transaction.
// Assumes that the output being spent has opTrueScript as it's scriptPublicKey
// Creates the value of the spent output minus 1 sompi
func CreateTransaction(txToSpend *externalapi.DomainTransaction) (*externalapi.DomainTransaction, error) {
	scriptPublicKey, err := txscript.PayToScriptHashScript(opTrueScript)
	if err != nil {
		return nil, err
	}
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
