package domainconverters

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// DomainTransactionToMsgTx converts a DomainTransaction into an appmessage.MsgTx
func DomainTransactionToMsgTx(domainTransaction *externalapi.DomainTransaction) *appmessage.MsgTx {
	txIns := make([]*appmessage.TxIn, 0, len(domainTransaction.Inputs))
	for _, input := range domainTransaction.Inputs {
		txIns = append(txIns, domainTransactionInputToTxIn(input))
	}

	txOuts := make([]*appmessage.TxOut, 0, len(domainTransaction.Outputs))
	for _, output := range domainTransaction.Outputs {
		txOuts = append(txOuts, domainTransactionOutputToTxOut(output))
	}

	return &appmessage.MsgTx{
		Version:      domainTransaction.Version,
		TxIn:         txIns,
		TxOut:        txOuts,
		LockTime:     domainTransaction.LockTime,
		SubnetworkID: domainTransaction.SubnetworkID,
		Gas:          domainTransaction.Gas,
		PayloadHash:  &domainTransaction.PayloadHash,
		Payload:      domainTransaction.Payload,
	}
}

func domainTransactionOutputToTxOut(domainTransactionOutput *externalapi.DomainTransactionOutput) *appmessage.TxOut {
	return &appmessage.TxOut{
		Value:        domainTransactionOutput.Value,
		ScriptPubKey: domainTransactionOutput.ScriptPublicKey,
	}
}

func domainTransactionInputToTxIn(domainTransactionInput *externalapi.DomainTransactionInput) *appmessage.TxIn {
	return &appmessage.TxIn{
		PreviousOutpoint: *domainOutpointToOutpoint(domainTransactionInput.PreviousOutpoint),
		SignatureScript:  domainTransactionInput.SignatureScript,
		Sequence:         domainTransactionInput.Sequence,
	}
}

func domainOutpointToOutpoint(domainOutpoint externalapi.DomainOutpoint) *appmessage.Outpoint {
	return appmessage.NewOutpoint(
		&domainOutpoint.TransactionID,
		domainOutpoint.Index)
}

func MsgTxToDomainTransaction(msgTx *appmessage.MsgTx) *externalapi.DomainTransaction {
	transactionInputs := make([]*externalapi.DomainTransactionInput, 0, len(msgTx.TxIn))
	for _, txIn := range msgTx.TxIn {
		transactionInputs = append(transactionInputs, txInToDomainTransactionInput(txIn))
	}

	transactionOutputs := make([]*externalapi.DomainTransactionOutput, 0, len(msgTx.TxOut))
	for _, txOut := range msgTx.TxOut {
		transactionOutputs = append(transactionOutputs, txOutToDomainTransactionOutput(txOut))
	}
	return &externalapi.DomainTransaction{
		Version:      msgTx.Version,
		Inputs:       transactionInputs,
		Outputs:      transactionOutputs,
		LockTime:     msgTx.LockTime,
		SubnetworkID: msgTx.SubnetworkID,
		Gas:          msgTx.Gas,
		PayloadHash:  *msgTx.PayloadHash,
		Payload:      msgTx.Payload,
	}
}

func txOutToDomainTransactionOutput(txOut *appmessage.TxOut) *externalapi.DomainTransactionOutput {
	return &externalapi.DomainTransactionOutput{
		Value:           txOut.Value,
		ScriptPublicKey: txOut.ScriptPubKey,
	}
}

func txInToDomainTransactionInput(txIn *appmessage.TxIn) *externalapi.DomainTransactionInput {
	return &externalapi.DomainTransactionInput{
		PreviousOutpoint: *outpointToDomainOutpoint(&txIn.PreviousOutpoint), //TODO
		SignatureScript:  txIn.SignatureScript,
		Sequence:         txIn.Sequence,
	}
}

func outpointToDomainOutpoint(outpoint *appmessage.Outpoint) *externalapi.DomainOutpoint {
	return &externalapi.DomainOutpoint{
		TransactionID: outpoint.TxID,
		Index:         outpoint.Index,
	}
}
